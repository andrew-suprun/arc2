package controller

import (
	m "arc/model"
	"arc/stream"
	w "arc/widgets"
	"fmt"
	"log"
	"runtime/debug"
	"slices"
	"strings"
	"time"
)

type controller struct {
	roots    []m.Root
	archives map[m.Root]*archive
	byId     map[m.Id]*m.File
	byHash   map[m.Hash][]*m.File
	archive  *archive
	hashed   int

	*shared

	screenSize w.Size
	frames     int
	prevTick   time.Time

	copySize        uint64
	totalCopiedSize uint64
	fileCopiedSize  uint64
	prevCopied      uint64
	copySpeed       float64
	timeRemaining   time.Duration

	errors []m.Error

	quit bool
}

type shared struct {
	fps int
	fs  m.FS
}

func Run(fs m.FS, renderer w.Renderer, events *stream.Stream[m.Event], roots []m.Root) (err any, stack []byte) {
	defer func() {
		err = recover()
		stack = debug.Stack()
	}()
	run(fs, renderer, events, roots)
	return nil, nil
}

func run(fs m.FS, renderer w.Renderer, events *stream.Stream[m.Event], roots []m.Root) {
	c := newController(roots)
	c.shared.fs = fs

	go ticker(events)

	for idx, root := range roots {
		c.archives[root] = newArchive(root, idx, c.shared)
		c.shared.fs.Scan(root)
	}
	c.archive = c.archives[roots[0]]

	for !c.quit {
		events, _ := events.Pull()
		for _, event := range events {
			c.handleEvent(event)
		}

		c.frames++
		screen := w.NewScreen(c.screenSize)
		c.archive.rootWidget().Render(screen, w.Position{X: 0, Y: 0}, c.screenSize)
		renderer.Push(screen)
	}
}

func newController(roots []m.Root) *controller {
	c := &controller{
		roots:    roots,
		archives: map[m.Root]*archive{},
		byId:     map[m.Id]*m.File{},
		byHash:   map[m.Hash][]*m.File{},
		shared:   &shared{},
	}
	return c
}

// Events

func (c *controller) tab() {
	selected := c.archive.currentFolder().selected()
	file, ok := selected.(*m.File)
	if !ok {
		return
	}

	sameHash := c.byHash[file.Hash]

	slices.SortFunc(sameHash, func(a, b *m.File) int {
		aIdx := c.archives[a.Root].idx
		bIdx := c.archives[b.Root].idx
		if aIdx < bIdx {
			return -1
		} else if aIdx > bIdx {
			return 1
		}
		if a.Path < b.Path {
			return -1
		} else if a.Path > b.Path {
			return 1
		}
		if a.Base < b.Base {
			return -1
		} else if a.Base > b.Base {
			return 1
		}
		return 0
	})

	idx := slices.Index(sameHash, file)
	var newSelected m.Entry = sameHash[(idx+1)%len(sameHash)]
	c.archive = c.archives[newSelected.Meta().Root]
	c.archive.currentPath = newSelected.Meta().Path
	folder := c.archive.currentFolder()
	folder.selectedIdx = slices.Index(folder.entries, newSelected)
	folder.makeSelectedVisible(c.archive.fileTreeLines)
}

func (c *controller) rootIdx(root m.Root) int {
	for idx := range c.roots {
		if root == c.roots[idx] {
			return idx
		}
	}
	return 0
}

func (c *controller) String() string {
	buf := &strings.Builder{}
	fmt.Fprintln(buf, "Controller:")
	for _, archive := range c.archives {
		archive.printTo(buf)
	}

	return buf.String()
}

func (c *controller) archiveHashed(event m.ArchiveHashed) {
	archive := c.archives[event.Root]
	archive.progressInfo = nil
	c.hashed++

	if c.hashed == len(c.roots) {
		c.analyzeDiscrepancies()
	}
}

func (c *controller) fileDeleted(event m.FileDeleted) {}

func (c *controller) fileRenamed(event m.FileRenamed) {
	c.analyzeDiscrepancy(event.Hash)
	c.updateFolderStates()
	log.Printf("fileRenamed: analyzeDiscrepancies")
}

func (c *controller) fileCopied(event m.FileCopied) {
	c.analyzeDiscrepancy(event.Hash)
	c.updateFolderStates()
}

func (c *controller) handleHashingProgress(event m.HashingProgress) {
	archive := c.archives[event.Root]
	archive.fileHashed = event.Hashed
	folder := archive.folders[event.Path]
	file := folder.entry(event.Base).(*m.File)
	file.SetState(m.Hashing)
	file.Hashed = event.Hashed

	c.archives[event.Root].progressInfo = &progressInfo{
		tab:           " Hashing",
		value:         float64(archive.totalHashed+uint64(archive.fileHashed)) / float64(archive.totalSize),
		speed:         archive.speed,
		timeRemaining: archive.timeRemaining,
	}

	archive.parents(file, func(file *m.Folder) {
		file.SetState(m.Hashing)
		file.Hashed = event.Hashed
	})

}

func (c *controller) handleCopyingProgress(event m.CopyingProgress) {
	c.fileCopiedSize = uint64(event)
	info := &progressInfo{
		tab:           " Copying",
		value:         float64(c.totalCopiedSize+uint64(c.fileCopiedSize)) / float64(c.copySize),
		speed:         c.copySpeed,
		timeRemaining: c.timeRemaining,
	}
	for _, archive := range c.archives {
		archive.progressInfo = info
	}
}

func (c *controller) analyzeDiscrepancies() {
	for hash := range c.byHash {
		c.analyzeDiscrepancy(hash)
	}
	c.updateFolderStates()
}

var noName = m.Name{}

func (c *controller) analyzeDiscrepancy(hash m.Hash) {
	files := c.byHash[hash]
	state := m.Resolved
	if len(files) != len(c.roots) {
		state = m.Divergent
	}
	if state != m.Divergent {
		name := m.Name{}
		for _, file := range files {
			if name == noName {
				name = file.Name
			}
			if file.Name != name {
				state = m.Divergent
				break
			}
		}
	}

	c.setStates(files, state)
	c.setCounts(files, state)
}

func (c *controller) resolveSelected() {
	selected := c.archive.currentFolder().selected()
	c.resolveFile(selected)
}

func (c *controller) resolveAll() {
	c.resolveFolder(c.archive.currentPath)
}

func (c *controller) resolveFile(entry m.Entry) {
	switch entry := entry.(type) {
	case *m.File:
		c.resolveRegularFile(entry)
	case *m.Folder:
		c.resolveFolder(entry.Path)
	}
}

func (c *controller) resolveFolder(path m.Path) {
	for _, entry := range c.archive.folders[path].entries {
		switch entry := entry.(type) {
		case *m.File:
			if entry.State() == m.Divergent && entry.Counts[c.archive.idx] == 1 {
				c.resolveFile(entry)
			}
		case *m.Folder:
			c.resolveFolder(m.Path(entry.Name.String()))
		}
	}
}

func (c *controller) setStates(files []*m.File, state m.State) {
	for _, file := range files {
		file.SetState(state)
	}
}

func (c *controller) setCounts(files []*m.File, state m.State) {
	if state != m.Divergent {
		for _, file := range files {
			file.Counts = nil
		}
		return
	}

	counts := make([]int, len(c.roots))

	for _, file := range files {
		for i, root := range c.roots {
			if root == file.Root {
				counts[i]++
			}
		}
	}

	for _, file := range files {
		file.Counts = counts
	}
}

func (c *controller) updateFolderStates() {
	for _, archive := range c.archives {
		archive.updateFolderStates("")
	}
}

func (c *controller) resolveRegularFile(file *m.File) {
	// c.cleanName(file)
	// c.cleanPath(file)

	// log.Printf("resolveFile: file: %s", file)
	// for _, archive := range c.archives {
	// 	log.Printf("resolveFile: archive:1: %q", archive.root)
	// 	archive.clearPath(file.Path)

	// 	var sameName m.Entry
	// 	folder := archive.folders[file.Path]
	// 	if folder == nil {
	// 		continue
	// 	}

	// 	for _, entry := range folder.entries {
	// 		if entry.Name == file.Name {
	// 			sameName = entry
	// 			break
	// 		}
	// 	}
	// 	if sameName == nil {
	// 		continue
	// 	}

	// 	if sameName.Hash != file.Hash {
	// 		log.Printf("resolveFile: found deverdent: id: %q, hash: %q", sameName.Id, sameName.Hash)
	// 		newBase := folder.uniqueName(file.Base)
	// 		newName := m.Name{Path: file.Path, Base: newBase}
	// 		archive.renameEntry(sameName, newName)
	// 		sameName.State = m.Pending
	// 		file.State = m.Pending

	// 		log.Printf("resolveFile: handled deverdent: id: %q, hash: %q", sameName.Id, sameName.Hash)
	// 	}
	// }

	// copyRoots := []m.Id{}
	// for _, archive := range c.archives {
	// 	log.Printf("resolveFile: archive:2: %q", archive.root)
	// 	sameHash := []*m.File{}
	// 	for _, entry := range c.byHash[file.Hash] {
	// 		if entry.Root == archive.root {
	// 			sameHash = append(sameHash, entry)
	// 			log.Printf("resolveFile: archive:2: sameHash: %q", entry.Id)
	// 		}
	// 	}
	// 	if len(sameHash) == 0 {
	// 		log.Printf("resolveFile: archive:2: no entries")
	// 		copyRoots = append(copyRoots, m.Id{Root: archive.root, Name: file.Name})
	// 		newFile := &m.Entry{
	// 			Meta: m.Meta{
	// 				Id:      m.Id{Root: archive.root, Name: file.Name},
	// 				Size:    file.Size,
	// 				ModTime: file.ModTime,
	// 			},
	// 			Hash:  file.Hash,
	// 			Kind:  m.FileRegular,
	// 			State: m.Pending,
	// 		}
	// 		file.State = m.Pending
	// 		log.Printf("resolveFile: archive:2: add entry: %q", newFile.Id)
	// 		archive.addFile(newFile)
	// 		c.byHash[newFile.Hash] = append(c.byHash[newFile.Hash], newFile)
	// 		continue
	// 	}
	// 	var keep m.Entry
	// 	for _, entry := range sameHash {
	// 		if entry.Name == file.Name {
	// 			keep = entry
	// 		}
	// 	}
	// 	if keep == nil {
	// 		keep = sameHash[0]
	// 	}
	// 	log.Printf("resolveFile: archive:2: keep: %q", keep.Id)
	// 	if keep.Name != file.Name {
	// 		archive.renameEntry(keep, file.Name)
	// 		keep.State = m.Pending
	// 		file.State = m.Pending
	// 	}

	// 	for _, other := range sameHash {
	// 		if other == keep {
	// 			continue
	// 		}
	// 		log.Printf("resolveFile: archive:2: other: %q", other.Id)

	// 		c.shared.fs.Send(m.DeleteFile{
	// 			Hash: keep.Hash,
	// 			Id:   other.Id,
	// 		})
	// 		file.State = m.Pending
	// 		archive.folders[other.Path].deleteEntry(other.Base)
	// 		files := c.byHash[keep.Hash]
	// 		for i, file := range files {
	// 			if file.Id == other.Id {
	// 				files[i] = files[len(files)-1]
	// 				c.byHash[keep.Hash] = files[:len(files)-1]
	// 				break
	// 			}
	// 		}
	// 	}
	// }
	// if len(copyRoots) > 0 {
	// 	c.shared.fs.Send(m.CopyFile{
	// 		Hash: file.Hash,
	// 		From: file.Id,
	// 		To:   copyRoots,
	// 	})
	// 	file.State = m.Pending
	// }
}
