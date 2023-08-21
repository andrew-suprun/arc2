package controller

import (
	m "arc/model"
	v "arc/view"
	w "arc/widgets"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
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
			folder.size += event.Size
			if folder.modTime.Before(event.ModTime) {
				folder.modTime = event.ModTime
			}
			folder.modTime = event.ModTime
			folder.state = v.Scanned
		}
		folder.files[base] = &file{
			size:    event.Size,
			modTime: event.ModTime,
			state:   v.Scanned,
		}

	case m.ArchiveScanned:
		c.archives[string(event)].scanned = true

	case m.FileHashed:
		path, base := parseName(event.Path)
		archive := c.archive(event.Root)
		folder := archive.folder(path)
		file := folder.file(base)
		file.hash = event.Hash
		file.setState(v.Hashed)

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
		folder.updateState()

	case m.HashingProgress:
		path, base := parseName(event.Path)
		file := c.archive(event.Root).folder(path).file(base)
		file.state = v.Hashing
		file.progress = event.Hashed

	case m.CopyingProgress:
		path, base := parseName(event.Path)
		file := c.archive(event.Root).folder(path).file(base)
		file.state = v.Copying
		file.progress = event.Copied

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
		archive := c.archive(c.root)
		path := filepath.Join(archive.currentPath...)
		folder := archive.folder(archive.currentPath)
		fileName := filepath.Join(c.root, path, folder.selectedName)
		exec.Command("open", fileName).Start()

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
		folder.makeSelectedVisible(c.fileTreeLines)

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
			archive.selectFile(cmd)

		case v.SelectFolder:
			archive.currentPath = cmd

		case v.SortColumn:
			folder := archive.curFolder()
			if cmd == folder.sortColumn {
				folder.sortAscending[folder.sortColumn] = !folder.sortAscending[folder.sortColumn]
			} else {
				folder.sortColumn = cmd
			}
			folder.makeSelectedVisible(c.fileTreeLines)
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
		c.curArchive().curFolder().makeSelectedVisible(c.fileTreeLines)

	case m.ResolveOne:
		c.resolveSelected()

	case m.ResolveAll:
		c.resolveFolder(c.curArchive().curFolder())

	case m.KeepAll:
		// TODO: Implement, maybe?

	case m.Delete:
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
	panic("IMPLEMENT c.tab()")
	// selected := c.curArchive().curFolder().selected()
	// file, ok := selected.(*m.File)
	// if !ok {
	// 	return
	// }

	// sameHash := c.byHash[file.Hash]

	// slices.SortFunc(sameHash, func(a, b *m.File) int {
	// 	result := cmp.Compare(c.archives[a.Root].idx, c.archives[b.Root].idx)
	// 	if result != 0 {
	// 		return result
	// 	}
	// 	result = cmp.Compare(a.Path, b.Path)
	// 	if result != 0 {
	// 		return result
	// 	}
	// 	return cmp.Compare(a.Base, b.Base)
	// })

	// idx := slices.Index(sameHash, file)
	// var newSelected m.Entry = sameHash[(idx+1)%len(sameHash)]
	// c.archive = c.archives[newSelected.Meta().Root]
	// c.archive.currentPath = newSelected.Meta().Path
	// folder := c.curArchive().curFolder()
	// folder.selectedIdx = slices.Index(folder.entries, newSelected)
	// folder.makeSelectedVisible(c.fileTreeLines)
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

var noName = m.Path{}

func (c *controller) analyzeDiscrepancy(hash m.Hash) {
	files := c.byHash[hash]
	divergent := false
	if len(files) != len(c.roots) {
		divergent = true
	}
	if !divergent {
		name := m.Path{}
		for _, file := range files {
			if name == noName {
				name = file.Name
			}
			if file.Name != name {
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
	panic("IMPLEMENT c.resolveFolder()")
	// for _, entry := range c.archive.folders[path].entries {
	// 	switch entry := entry.(type) {
	// 	case *m.File:
	// 		if entry.State == m.Divergent && entry.Counts[c.archive.idx] == 1 {
	// 			c.resolveFile(entry)
	// 		}
	// 	case *m.Folder:
	// 		c.resolveFolder(m.Path(entry.Name.String()))
	// 	}
	// }
}

func (c *controller) setStates(files []*m.File, state m.State) {
	for _, file := range files {
		file.State = state
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

func (c *controller) resolveRegularFile(file *m.File) {
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

func (c *controller) clearName(file *m.File) {
	if c.nameCollidesWithPath(file.Name) {
		var newName m.Path
		for i := 1; ; i++ {
			newName = newSuffix(file.Name, i)
			if !c.nameCollidesWithPath(newName) {
				break
			}
		}
		if newName != file.Name {
			c.archives[file.Root].renameEntry(file, newName)
		}
	}
}

func (c *controller) nameCollidesWithPath(name m.Path) bool {
	path := name.ChildPath()
	for _, archive := range c.archives {
		if _, ok := archive.folders[path]; ok {
			return true
		}
	}
	return false
}

func newSuffix(name m.Path, idx int) m.Path {
	parts := strings.Split(name.Base.String(), ".")

	var part string
	if len(parts) == 1 {
		part = stripIdx(parts[0])
	} else {
		part = stripIdx(parts[len(parts)-2])
	}
	var newBase m.Base
	if len(parts) == 1 {
		newBase = m.Base(fmt.Sprintf("%s%c%d", part, '`', idx))
	} else {
		parts[len(parts)-2] = fmt.Sprintf("%s%c%d", part, '`', idx)
		newBase = m.Base(strings.Join(parts, "."))
	}
	return m.Path{Path: name.Path, Base: newBase}
}

type stripIdxState int

const (
	expectDigit stripIdxState = iota
	expectDigitOrBacktick
)

func stripIdx(name string) string {
	state := expectDigit
	i := len(name) - 1
	for ; i >= 0; i-- {
		ch := name[i]
		if ch >= '0' && ch <= '9' && (state == expectDigit || state == expectDigitOrBacktick) {
			state = expectDigitOrBacktick
		} else if ch == '`' && state == expectDigitOrBacktick {
			return name[:i]
		} else {
			return name
		}
	}
	return name
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

func (a *archive) open() {
	id := id{Root: a.root, Name: m.Path{Path: a.currentPath, Base: a.currentFolder().selectedName}}
	exec.Command("open", id.String()).Start()
}

func (a *archive) revealInFinder() {
	id := id{Root: a.root, Name: m.Path{Path: a.currentPath, Base: a.currentFolder().selectedName}}
	exec.Command("open", "-R", id.String()).Start()
}

func (a *archive) selectFile(cmd m.SelectFile) {
	folder := a.currentFolder()
	if folder.selectedName == cmd.Base && time.Since(folder.lastMouseEventTime).Seconds() < 0.5 {
		a.open()
	} else {
		folder.selectedName = cmd.Base
	}
	folder.lastMouseEventTime = time.Now()
}

func (f *folder) makeSelectedVisible(fileTreeLines int) {
	if f.offsetIdx > f.selectedIdx {
		f.offsetIdx = f.selectedIdx
	}
	if f.offsetIdx < f.selectedIdx+1-fileTreeLines {
		f.offsetIdx = f.selectedIdx + 1 - fileTreeLines
	}
}
