package controller

import (
	m "arc/model"
	"arc/stream"
	w "arc/widgets"
	"fmt"
	"log"
	"runtime/debug"
	"sort"
	"strings"
	"time"
)

type controller struct {
	roots    []m.Root
	archives map[m.Root]*archive
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

	errors []any

	quit bool
}

type shared struct {
	fps int
	fs  m.FS
}

func Run(fs m.FS, renderer w.Renderer, events *stream.Stream[m.Event], roots []m.Root) {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("PANIC: %#v", err)
			debug.PrintStack()
		}
	}()
	run(fs, renderer, events, roots)
}

func run(fs m.FS, renderer w.Renderer, events *stream.Stream[m.Event], roots []m.Root) {
	c := newController(roots)
	c.shared.fs = fs

	go ticker(events)

	for _, root := range roots {
		c.archives[root] = newArchive(root, c.shared)
		c.shared.fs.Scan(root)
	}
	c.archive = c.archives[roots[0]]

	for !c.quit {
		for _, event := range events.Pull() {
			c.handleEvent(event)
		}

		c.frames++
		c.archive.currentFolder().sort()
		screen := w.NewScreen(c.screenSize)
		c.archive.rootWidget().Render(screen, w.Position{X: 0, Y: 0}, c.screenSize)
		renderer.Push(screen)
	}
}

func newController(roots []m.Root) *controller {
	c := &controller{
		roots:    roots,
		archives: map[m.Root]*archive{},
		shared:   &shared{},
	}
	return c
}

// Events

func (c *controller) tab() {
	selected := c.archive.currentFolder().selected()

	if selected == nil || selected.Kind != m.FileRegular {
		return
	}
	sameHash := []*m.File{}
	for _, archive := range c.archives {
		for _, folder := range archive.folders {
			for _, entry := range folder.entries {
				if entry.Hash == selected.Hash {
					sameHash = append(sameHash, entry)
				}
			}
		}
	}

	sort.Slice(sameHash, func(i, j int) bool {
		iFile := sameHash[i]
		jFile := sameHash[j]
		iRootIdx := c.rootIdx(iFile.Root)
		jRootIdx := c.rootIdx(jFile.Root)

		if iRootIdx != jRootIdx {
			return iRootIdx < jRootIdx
		}

		iName := strings.ToLower(iFile.Id.String())
		jName := strings.ToLower(jFile.Id.String())

		return iName < jName
	})

	idx := 0
	for idx = range sameHash {
		if sameHash[idx] == selected {
			break
		}
	}

	newSelected := sameHash[(idx+1)%len(sameHash)]
	c.archive = c.archives[newSelected.Root]
	c.archive.currentPath = newSelected.Path

	for idx, entry := range c.archive.currentFolder().entries {
		if newSelected == entry {
			c.archive.currentFolder().selectedIdx = idx
			break
		}
	}
	c.archive.currentFolder().makeSelectedVisible(c.archive.fileTreeLines)
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

func (c *controller) fileRenamed(event m.FileRenamed) {
	c.setState(event.To, m.Resolved)
	c.setState(m.Id{Root: event.From.Root, Name: event.To.Name}, m.Resolved)
	c.analyzeDiscrepancies()
}

func (c *controller) fileCopied(event m.FileCopied) {
	c.setState(event.From, m.Resolved)
	for _, to := range event.To {
		c.setState(to, m.Resolved)
	}
	c.analyzeDiscrepancies()
}

func (c *controller) setState(id m.Id, state m.State) {
	folder := c.archives[id.Root].folders[id.Path]
	for _, entry := range folder.entries {
		if entry.Base == id.Base {
			entry.State = state
		}
	}
}

func (c *controller) handleHashingProgress(event m.HashingProgress) {
	archive := c.archives[event.Root]
	archive.fileHashed = event.Hashed
	folder := archive.folders[event.Path]
	file := folder.entry(event.Base)
	file.State = m.Hashing
	file.Hashed = event.Hashed

	c.archives[event.Root].progressInfo = &progressInfo{
		tab:           " Hashing",
		value:         float64(archive.totalHashed+uint64(archive.fileHashed)) / float64(archive.totalSize),
		speed:         archive.speed,
		timeRemaining: archive.timeRemaining,
	}

	archive.parents(file, func(file *m.File) {
		file.State = m.Hashing
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
	allNames := map[m.Name]m.Hash{}
	archiveNames := map[m.Root]map[m.Name]m.Hash{}
	for _, root := range c.roots {
		archiveNames[root] = map[m.Name]m.Hash{}
	}

	for _, archive := range c.archives {
		for _, folder := range archive.folders {
			for _, entry := range folder.entries {
				allNames[entry.Name] = entry.Hash
				archiveNames[entry.Root][entry.Name] = entry.Hash
			}
		}
	}

	divergency := map[m.Name]struct{}{}
	for name, hash := range allNames {
		for _, archNames := range archiveNames {
			if archNames[name] != hash {
				divergency[name] = struct{}{}
			}
		}
	}

	for _, archive := range c.archives {
		for _, folder := range archive.folders {
			for _, entry := range folder.entries {
				if _, ok := divergency[entry.Name]; ok && entry.State != m.Duplicate {
					entry.State = m.Divergent
				}
			}
		}
		archive.updateFolderStates("")
	}
}

func (c *controller) keepSelected() {
	selected := c.archive.currentFolder().selected()
	c.keepFile(selected)
}

func (c *controller) keepFile(file *m.File) {
	for _, archive := range c.archives {
		var sameName *m.File
		folder := archive.folders[file.Path]
		for _, entry := range folder.entries {
			if entry.Name == file.Name {
				sameName = entry
				break
			}
		}
		if sameName == nil {
			continue
		}
		// TODO What if a sameName file is a folder?
		if sameName.Hash != file.Hash {
			newBase := folder.uniqueName(file.Base)
			newId := m.Id{Root: file.Root, Name: m.Name{Path: file.Path, Base: newBase}}
			c.shared.fs.Send(m.RenameFile{
				Hash: sameName.Hash,
				From: sameName.Id,
				To:   newId,
			})
			archive.renameEntry(sameName.Name, newId.Name)
			sameName.State = m.Pending
			file.State = m.Pending
		}
	}

	copyRoots := []m.Id{}
	for _, archive := range c.archives {
		sameHash := []*m.File{}
		for _, folder := range archive.folders {
			for _, entry := range folder.entries {
				if entry.Hash == file.Hash {
					sameHash = append(sameHash, entry)
				}
			}
		}
		if len(sameHash) == 0 {
			copyRoots = append(copyRoots, m.Id{Root: archive.root, Name: file.Name})
			newFile := &m.File{
				Meta: m.Meta{
					Id:      m.Id{Root: archive.root, Name: file.Name},
					Size:    file.Size,
					ModTime: file.ModTime,
					Hash:    file.Hash,
				},
				Kind:  m.FileRegular,
				State: m.Pending,
			}
			file.State = m.Pending
			archive.addEntry(newFile)
			continue
		}
		var keep *m.File
		for _, entry := range sameHash {
			if entry.Name == file.Name {
				keep = entry
			}
		}
		if keep == nil {
			keep = sameHash[0]
		}
		if keep.Id != file.Id {
			c.shared.fs.Send(m.RenameFile{
				Hash: keep.Hash,
				From: keep.Id,
				To:   file.Id,
			})
			archive.renameEntry(keep.Name, file.Name)
			keep.State = m.Pending
			file.State = m.Pending
		}

		for _, other := range sameHash {
			if other == keep {
				continue
			}

			c.shared.fs.Send(m.DeleteFile{
				Hash: keep.Hash,
				Id:   other.Id,
			})
			archive.removeEntry(other.Name)
		}
	}
	if len(copyRoots) > 0 {
		c.shared.fs.Send(m.CopyFile{
			Hash: file.Hash,
			From: file.Id,
			To:   copyRoots,
		})
		file.State = m.Pending
	}
}
