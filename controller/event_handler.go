package controller

import (
	m "arc/model"
	v "arc/view"
	w "arc/widgets"
	"cmp"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

func (c *controller) handleEvent(event any) {
	// log.Printf("### event: %T: %s", event, event)
	if event == nil {
		return
	}
	switch event := event.(type) {
	case m.FileScanned:
		path, base := parseName(event.Path)
		archive := c.archive(event.Root)
		folder := archive.rootFolder
		for _, name := range path[:len(path)-1] {
			folder = folder.child(name)
		}
		folder.files[base] = &file{
			meta: meta{
				root:    event.Root,
				name:    base,
				parent:  folder,
				size:    event.Size,
				modTime: event.ModTime,
				state:   v.Scanned,
			},
		}

	case m.ArchiveScanned:
		c.archives[string(event)].scanned = true

	case m.FileHashed:
		path, base := parseName(event.Path)
		archive := c.archive(event.Root)
		folder := archive.folder(path)
		file := folder.file(base)
		file.hash = event.Hash
		file.state = v.Hashed
		c.byHash[file.hash] = append(c.byHash[file.hash], file)

	case m.ArchiveHashed:
		c.archives[string(event)].state = ready
		for _, archive := range c.archives {
			if archive.state != ready {
				return
			}
		}
		c.analyzeDiscrepancies()

	case m.FileDeleted:
		// No Action Needed

	case m.FileRenamed:
		// No Action Needed

	case m.FileCopied:
		path, base := parseName(event.From)
		folder := c.archive(event.Root).folder(path)
		folder.file(base).state = v.Resolved
		// TODO Change all v.Copied to v.Resolved if possible

	case m.HashingProgress:
		path, base := parseName(event.Path)
		file := c.archive(event.Root).folder(path).file(base)
		file.state = v.Hashing
		file.progress = &v.Progress{
			Size: file.size,
			Done: event.Hashed,
		}

	case m.CopyingProgress:
		path, base := parseName(event.Path)
		file := c.archive(event.Root).folder(path).file(base)
		file.state = v.Copying
		file.progress = &v.Progress{
			Size: file.size,
			Done: event.Copied,
		}

	case m.ScreenSize:
		c.screenSize = w.Size{Width: event.Width, Height: event.Height}

	case m.SelectArchive:
		if event.Idx < len(c.roots) {
			c.root = c.roots[event.Idx]
		}

	case m.Enter:
		archive := c.archive(c.root)
		folder := archive.folder(archive.currentPath)
		if folder != nil && folder.selectedName != "" {
			archive.currentPath = append(archive.currentPath, folder.selectedName)
		}

	case m.Exit:
		archive := c.archive(c.root)
		archive.currentPath = archive.currentPath[:len(archive.currentPath)-1]

	case m.Open:
		c.open()

	case m.RevealInFinder:
		archive := c.archive(c.root)
		folder := archive.folder(archive.currentPath)
		if folder.selectedName != "" {
			path := filepath.Join(archive.currentPath...)
			fileName := filepath.Join(c.root, path, folder.selectedName)
			exec.Command("open", "-R", fileName).Start()
		}

	case m.MoveSelection:
		log.Printf("m.MoveSelection: Lines: %d", event.Lines)
		folder := c.curArchive().curFolder()
		folder.moveSelection(event.Lines)
		c.makeSelectedVisible = true

	case m.SelectFirst:
		folder := c.curArchive().curFolder()
		folder.selectedName = ""
		folder.selectedIdx = 0

	case m.SelectLast:
		folder := c.curArchive().curFolder()
		folder.selectedName = ""
		folder.selectedIdx = len(folder.children) + len(folder.files) - 1

	case m.Scroll:
		c.curArchive().curFolder().moveOffset(event.Lines, c.fileTreeLines)

	case m.MouseTarget:
		archive := c.curArchive()
		switch cmd := event.Command.(type) {
		case v.SelectFile:
			folder := c.curArchive().folder(cmd.Path)
			if folder.selectedName == cmd.Name && time.Since(c.lastMouseEventTime).Seconds() < 0.5 {
				c.open()
			} else {
				folder.selectedName = cmd.Name
			}
			c.lastMouseEventTime = time.Now()

		case v.SelectFolder:
			archive.currentPath = cmd

		case v.SortColumn:
			folder := archive.curFolder()
			if cmd == folder.sortColumn {
				folder.sortAscending[folder.sortColumn] = !folder.sortAscending[folder.sortColumn]
			} else {
				folder.sortColumn = cmd
			}
			c.makeSelectedVisible = true
		}

	case m.PgUp:
		folder := c.curArchive().curFolder()
		folder.moveOffset(-c.fileTreeLines, c.fileTreeLines)
		folder.moveSelection(-c.fileTreeLines)

	case m.PgDn:
		folder := c.curArchive().curFolder()
		folder.moveOffset(c.fileTreeLines, c.fileTreeLines)
		folder.moveSelection(c.fileTreeLines)

	case m.Tab:
		c.tab()
		c.makeSelectedVisible = true

	case m.ResolveOne:
		c.resolveSelected()

	case m.ResolveAll:
		c.resolveFolder(c.curArchive().curFolder())

	case m.KeepAll:
		// TODO: Implement, maybe?

	case m.Delete:
		// TODO
		// folder := c.currentFolder()
		// c.deleteFile(folder.selectedEntry)

	case m.Error:
		log.Printf("### Error: %s", event)
		c.errors = append(c.errors, event)

	case m.Quit:
		c.quit = true

	case m.DebugPrintState:
		log.Println(c.String())

	case m.DebugPrintRootWidget:
		log.Println(c.view().RootWidget())

	default:
		log.Panicf("### unhandled event: %#v", event)
	}
}

func (a *archive) archiveScanned() {
	a.scanned = true
}

func (c *controller) tab() {
	folder := c.curArchive().curFolder()
	curFile := folder.files[folder.selectedName]
	if curFile == nil {
		return
	}

	sameHash := c.byHash[curFile.hash]

	slices.SortFunc(sameHash, func(a, b *file) int {
		result := cmp.Compare(c.archives[a.root].idx, c.archives[b.root].idx)
		if result != 0 {
			return result
		}
		return cmp.Compare(a.fullName(), b.fullName())
	})

	idx := slices.Index(sameHash, curFile)
	var newSelected = sameHash[(idx+1)%len(sameHash)]
	c.root = newSelected.root
	archive := c.curArchive()
	folder = newSelected.parent
	archive.currentPath = folder.path()
	folder.selectedName = newSelected.name
	c.makeSelectedVisible = true
}

func (c *controller) String() string {
	buf := &strings.Builder{}
	fmt.Fprintln(buf, "Controller:")
	for _, archive := range c.archives {
		archive.printTo(buf)
	}

	return buf.String()
}

func (c *controller) analyzeDiscrepancies() {
	for hash := range c.byHash {
		c.analyzeDiscrepancy(hash)
	}
}

func (c *controller) analyzeDiscrepancy(hash string) {
	files := c.byHash[hash]
	divergent := false
	if len(files) != len(c.roots) {
		divergent = true
	}
	if !divergent {
		name := files[0].name
		for _, file := range files {
			if file.name != name {
				divergent = true
				break
			}
		}
	}

	if divergent {
		c.setStates(files, v.Divergent)
		c.setCounts(files, v.Divergent)
	}
}

func (c *controller) resolveSelected() {
	folder := c.curArchive().curFolder()
	c.resolveFile(folder, folder.selectedName)
}

func (c *controller) resolveFile(folder *folder, name string) {
	subfolder := folder.children[name]
	if subfolder != nil {
		c.resolveFolder(subfolder)
		return
	}
	file := folder.file(name)
	if file != nil {
		c.resolveRegularFile(file)
	}
}

func (c *controller) resolveFolder(folder *folder) {
	for _, sub := range folder.children {
		c.resolveFolder(sub)
	}
	for _, file := range folder.files {
		c.resolveRegularFile(file)
	}
}

func (c *controller) setStates(files []*file, state v.State) {
	for _, file := range files {
		file.state = state
	}
}

func (c *controller) setCounts(files []*file, state v.State) {
	if state != v.Divergent {
		for _, file := range files {
			file.counts = nil
		}
		return
	}

	counts := make([]int, len(c.roots))

	for _, file := range files {
		for i, root := range c.roots {
			if root == file.root {
				counts[i]++
			}
		}
	}

	for _, file := range files {
		file.counts = counts
	}
}

func (c *controller) resolveRegularFile(file *file) {
	panic("IMPLEMENT: c.resolveRegularFile()")
	// c.clearName(file)

	// for _, archive := range c.archives {
	// 	if folder, ok := archive.folders[file.Path]; ok {
	// 		entry := folder.entry(file.Base)
	// 		if archiveFile, ok := entry.(*m.File); ok && archiveFile.Hash != file.Hash {

	// 		}
	// 	}
	// }

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

	// copyRoots := []id{}
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
	// 		copyRoots = append(copyRoots, id{Root: archive.root, Name: file.Name})
	// 		newFile := &m.Entry{
	// 			Meta: m.Meta{
	// 				Id:      id{Root: archive.root, Name: file.Name},
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

func (f *folder) moveSelection(lines int) {
	f.selectedName = ""
	f.selectedIdx += lines
	entries := len(f.children) + len(f.files)

	if f.selectedIdx >= entries {
		f.selectedIdx = entries - 1
	}
	if f.selectedIdx < 0 {
		f.selectedIdx = 0
	}
	log.Printf("moveSelection: selectedIdx: %d", f.selectedIdx)
}

func (f *folder) moveOffset(lines, fileTreeLines int) {
	f.offsetIdx += lines
	entries := len(f.children) + len(f.files)

	if f.offsetIdx >= entries+1-fileTreeLines {
		f.offsetIdx = entries - fileTreeLines
	}
	if f.offsetIdx < 0 {
		f.offsetIdx = 0
	}
}

func (c *controller) open() {
	archive := c.archive(c.root)
	path := filepath.Join(archive.currentPath...)
	folder := archive.folder(archive.currentPath)
	fileName := filepath.Join(c.root, path, folder.selectedName)
	exec.Command("open", fileName).Start()
}
