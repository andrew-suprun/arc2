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
		file := &file{Meta: event.Meta, State: m.Scanned, progressSize: event.Meta.Size}
		c.archive(event.Id).getFolder(event.Path).files[event.Base] = file

	case m.ArchiveScanned:
		c.archives[event.Root].state = hashing

	case m.FileHashed:
		file := c.file(event.Id)
		file.Hash = event.Hash
		file.State = m.Hashed
		c.byHash[event.Hash] = append(c.byHash[event.Hash], file)

	case m.ArchiveHashed:
		c.archives[event.Root].state = ready
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
		c.file(event.From).State = m.Copied

	case m.HashingProgress:
		file := c.file(event.Id)
		file.State = m.Hashing
		file.progressDone = event.Hashed

	case m.CopyingProgress:
		file := c.file(event.Id)
		file.State = m.Copying
		file.progressDone = event.Copied

	case m.ScreenSize:
		c.screenSize = w.Size{Width: event.Width, Height: event.Height}

	case m.SelectArchive:
		if event.Idx < len(c.roots) {
			c.currRoot = c.roots[event.Idx]
		}

	case m.Enter:
		archive := c.currArchive()
		folder := c.currFolder()
		base := folder.selectedBase
		if folder.files[base] == nil {
			archive.currentPath = m.Path(filepath.Join(archive.currentPath.String(), base.String()))
		}

	case m.Exit:
		archive := c.currArchive()
		if archive.currentPath == "" {
			return
		}
		parts := strings.Split(archive.currentPath.String(), "/")
		if len(parts) == 1 {
			archive.currentPath = ""
		}
		archive.currentPath = m.Path(filepath.Join(parts[:len(parts)-1]...))

	case m.Open:
		c.open()

	case m.RevealInFinder:
		c.reveal()

	case m.MoveSelection:
		folder := c.currFolder()
		folder.moveSelection(event.Lines)
		folder.makeSelectedVisible(c.currArchive().fileTreeLines)

	case m.SelectFirst:
		folder := c.currFolder()
		folder.selectedBase = ""
		folder.selectedIdx = 0

	case m.SelectLast:
		folder := c.currFolder()
		folder.selectedBase = ""
		folder.selectedIdx = folder.entries - 1

	case m.Scroll:
		c.currFolder().moveOffset(event.Lines, c.currArchive().fileTreeLines)

	case m.MouseTarget:
		switch cmd := event.Command.(type) {
		case m.SelectFile:
			c.selectFile(cmd)

		case m.SelectFolder:
			c.currArchive().currentPath = m.Path(cmd)

		case m.SortColumn:
			folder := c.currFolder()
			folder.selectSortColumn(cmd)
			folder.makeSelectedVisible(c.currArchive().fileTreeLines)
		}

	case m.PgUp:
		archive := c.currArchive()
		folder := c.currFolder()
		folder.moveOffset(-archive.fileTreeLines, archive.fileTreeLines)
		folder.moveSelection(-archive.fileTreeLines)

	case m.PgDn:
		archive := c.currArchive()
		folder := c.currFolder()
		folder.moveOffset(archive.fileTreeLines, archive.fileTreeLines)
		folder.moveSelection(archive.fileTreeLines)

	case m.Tab:
		c.tab()
		c.currFolder().makeSelectedVisible(c.currArchive().fileTreeLines)

	case m.ResolveOne:
		c.resolveSelected()

	case m.ResolveAll:
		c.resolveAll()

	case m.KeepAll:
		// TODO: Implement, maybe?

	case m.Delete:
		// TODO
		// folder := c.currFolder()
		// c.deleteFile(folder.selectedEntry)

	case m.Error:
		log.Printf("### Error: %s", event)
		c.errors = append(c.errors, event)

	case m.Quit:
		c.quit = true

	case m.DebugPrintView:
		log.Println(c.view().String())

	case m.DebugPrintRootWidget:
		log.Println(c.view().RootWidget())

	default:
		log.Panicf("### unhandled event: %#v", event)
	}
}

func (c *controller) tab() {
	panic("IMPLEMENT c.tab()")
	// selected := c.archive.currFolder().selected()
	// file, ok := selected.(*File)
	// if !ok {
	// 	return
	// }

	// sameHash := c.byHash[file.Hash]

	// slices.SortFunc(sameHash, func(a, b *File) int {
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
	// folder := c.archive.currFolder()
	// folder.selectedIdx = slices.Index(folder.entries, newSelected)
	// folder.makeSelectedVisible(c.archive.fileTreeLines)
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

var noName = m.Name{}

func (c *controller) analyzeDiscrepancy(hash m.Hash) {
	files := c.byHash[hash]
	divergent := false
	if len(files) != len(c.roots) {
		divergent = true
	}
	if !divergent {
		name := m.Name{}
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
		c.setStates(files, m.Divergent)
		c.setCounts(files, m.Divergent)
	}
}

func (c *controller) resolveSelected() {
	panic("IMPLEMENT c.resolveSelected()")
	// c.resolveFile(c.archive.currFolder().selectedEntry)
}

func (c *controller) resolveAll() {
	c.resolveFolder(c.currArchive().currentPath)
}

func (c *controller) resolveFile(entry v.Entry) {
	// TODO
	// switch entry.Kind {
	// case v.Regular:
	// 	c.resolveRegularFile(entry.File)
	// case v.Folder:
	// 	c.resolveFolder(entry.Path)
	// }
}

func (c *controller) resolveFolder(path m.Path) {
	panic("IMPLEMENT c.resolveFolder()")
	// for _, entry := range c.archive.folders[path].entries {
	// 	switch entry := entry.(type) {
	// 	case *File:
	// 		if entry.State == m.Divergent && entry.Counts[c.archive.idx] == 1 {
	// 			c.resolveFile(entry)
	// 		}
	// 	case *m.Folder:
	// 		c.resolveFolder(m.Path(entry.Name.String()))
	// 	}
	// }
}

func (c *controller) setStates(files []*file, state m.State) {
	for _, file := range files {
		file.State = state
	}
}

func (c *controller) setCounts(files []*file, state m.State) {
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

func (c *controller) resolveRegularFile(meta *file) {
	panic("IMPLEMENT: c.resolveRegularFile()")
	// c.clearName(file)

	// for _, archive := range c.archives {
	// 	if folder, ok := archive.folders[file.Path]; ok {
	// 		entry := folder.entry(file.Base)
	// 		if archiveFile, ok := entry.(*File); ok && archiveFile.Hash != file.Hash {

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

	// copyRoots := []m.Id{}
	// for _, archive := range c.archives {
	// 	log.Printf("resolveFile: archive:2: %q", archive.root)
	// 	sameHash := []*File{}
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
	// 			Kind:  FileRegular,
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

func (c *controller) clearName(meta *file) {
	if c.nameCollidesWithPath(meta.Name) {
		var newName m.Name
		for i := 1; ; i++ {
			newName = newSuffix(meta.Name, i)
			if !c.nameCollidesWithPath(newName) {
				break
			}
		}
		if newName != meta.Name {
			c.renameFile(meta, newName)
		}
	}
}

func (c *controller) nameCollidesWithPath(name m.Name) bool {
	path := name.ChildPath()
	for _, archive := range c.archives {
		if _, ok := archive.folders[path]; ok {
			return true
		}
	}
	return false
}

func newSuffix(name m.Name, idx int) m.Name {
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
	return m.Name{Path: name.Path, Base: newBase}
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
	f.selectedBase = ""
	f.selectedIdx += lines

	if f.selectedIdx >= f.entries {
		f.selectedIdx = f.entries - 1
	}
	if f.selectedIdx < 0 {
		f.selectedIdx = 0
	}
}

func (f *folder) moveOffset(lines, fileTreeLines int) {
	f.offsetIdx += lines

	if f.offsetIdx >= f.entries+1-fileTreeLines {
		f.offsetIdx = f.entries - fileTreeLines
	}
	if f.offsetIdx < 0 {
		f.offsetIdx = 0
	}
}

func (c *controller) reveal() {
	archive := c.currArchive()
	folder := c.currFolder()
	id := m.Id{Root: c.currRoot, Name: m.Name{Path: archive.currentPath, Base: folder.selectedBase}}
	exec.Command("open", "-R", id.String()).Start()
}

func (c *controller) open() {
	archive := c.currArchive()
	folder := c.currFolder()
	id := m.Id{Root: c.currRoot, Name: m.Name{Path: archive.currentPath, Base: folder.selectedBase}}
	exec.Command("open", id.String()).Start()
}

func (c *controller) selectFile(cmd m.SelectFile) {
	folder := c.currFolder()
	if folder.selectedBase == cmd.Base && time.Since(folder.lastMouseEventTime).Seconds() < 0.5 {
		c.open()
	} else {
		folder.selectedBase = cmd.Base
	}
	folder.lastMouseEventTime = time.Now()
}

func (f *folder) selectSortColumn(cmd m.SortColumn) {
	if cmd == f.sortColumn {
		f.sortAscending[f.sortColumn] = !f.sortAscending[f.sortColumn]
	} else {
		f.sortColumn = cmd
	}
}

func (f *folder) makeSelectedVisible(fileTreeLines int) {
	if f.offsetIdx > f.selectedIdx {
		f.offsetIdx = f.selectedIdx
	}
	if f.offsetIdx < f.selectedIdx+1-fileTreeLines {
		f.offsetIdx = f.selectedIdx + 1 - fileTreeLines
	}
}
