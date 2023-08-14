package controller

import (
	m "arc/model"
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
		archive := c.archives[event.Root]
		file := m.NewFile(event.Meta, m.Scanned)
		archive.totalSize += file.Size
		folder := archive.getFolder(file.Path)
		folder.files[file.Base] = file

	case m.ArchiveScanned:
		c.archives[event.Root].scanned = true

	case m.FileHashed:
		file := c.archives[event.Root].getFolder(event.Path).files[event.Base]
		file.Hash = event.Hash
		file.SetState(m.Resolved)
		c.byHash[event.Hash] = append(c.byHash[event.Hash], file)

	case m.ArchiveHashed:
		c.archives[event.Root].state = ready

	case m.FileDeleted:
		// No Action Needed

	case m.FileRenamed:
		c.archives[event.From.Root].folders[event.To.Path].files[event.To.Base].SetState(m.Resolved)

	case m.FileCopied:
		c.archives[event.From.Root].folders[event.From.Path].files[event.From.Base].SetState(m.Resolved)
		for _, to := range event.To {
			c.archives[to.Root].folders[to.Path].files[to.Base].SetState(m.Resolved)
		}

	case m.HashingProgress:
		c.handleHashingProgress(event)

	case m.CopyingProgress:
		c.handleCopyingProgress(event)

	case m.Tick:
		c.handleTick(event)

	case m.ScreenSize:
		c.screenSize = w.Size{Width: event.Width, Height: event.Height}

	case m.SelectArchive:
		if event.Idx < len(c.roots) {
			c.archive = c.archives[c.roots[event.Idx]]
		}

	case m.Enter:
		c.archive.enter()

	case m.Exit:
		c.archive.exit()

	case m.Open:
		c.archive.open()

	case m.RevealInFinder:
		c.archive.revealInFinder()

	case m.MoveSelection:
		log.Printf("m.MoveSelection: Lines: %d", event.Lines)
		folder := c.archive.currentFolder()
		folder.moveSelection(event.Lines)
		folder.makeSelectedVisible(c.archive.fileTreeLines)

	case m.SelectFirst:
		c.archive.currentFolder().selectFirst()

	case m.SelectLast:
		c.archive.currentFolder().selectLast()

	case m.Scroll:
		c.archive.currentFolder().moveOffset(event.Lines, c.archive.fileTreeLines)

	case m.MouseTarget:
		c.archive.mouseTarget(event.Command)

	case m.PgUp:
		folder := c.archive.currentFolder()
		folder.moveOffset(-c.archive.fileTreeLines, c.archive.fileTreeLines)
		folder.moveSelection(-c.archive.fileTreeLines)

	case m.PgDn:
		folder := c.archive.currentFolder()
		folder.moveOffset(c.archive.fileTreeLines, c.archive.fileTreeLines)
		folder.moveSelection(c.archive.fileTreeLines)

	case m.Tab:
		c.tab()
		c.archive.currentFolder().makeSelectedVisible(c.archive.fileTreeLines)

	case m.ResolveOne:
		c.resolveSelected()

	case m.ResolveAll:
		c.resolveAll()

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

func (c *controller) fileHashed(event m.FileHashed) {
	folder := c.archive.folders[event.Path]
	file := folder.files[event.Base]
	file.Hash = event.Hash
	c.byHash[event.Hash] = append(c.byHash[event.Hash], file)
}

func (a *archive) archiveHashed() {
	a.state = ready
}

func (a *archive) addFile(file *m.File) {
	panic("ERROR")
	// a.insertEntry(file)
	// name := file.ParentName()
	// for name.Base != "." {
	// 	parentFolder := a.getFolder(name.Path)
	// 	entry := parentFolder.entry(name.Base)

	// 	if folderEntry, ok := entry.(*m.Folder); ok {
	// 		folderEntry.Size += file.Size
	// 		if folderEntry.ModTime.Before(file.ModTime) {
	// 			folderEntry.ModTime = file.ModTime
	// 		}
	// 	} else {
	// 		folderEntry := m.NewFolder(m.Meta{
	// 			Id: m.Id{
	// 				Root: file.Root,
	// 				Name: name,
	// 			},
	// 			Size:    file.Size,
	// 			ModTime: file.ModTime,
	// 		},
	// 			m.Scanned,
	// 		)
	// 		a.insertEntry(folderEntry)

	// 	}

	// 	name = name.Path.ParentName()
	// }
}

// func (a *archive) insertEntry(entry m.Entry) {
// 	folder := a.getFolder(entry.Meta().Path)
// 	folder.insertEntry(entry)
// }

func (a *archive) enter() {
	folder := a.currentFolder()
	base := folder.selectedBase
	file := folder.files[base]
	if file == nil {
		a.currentPath = m.Path(filepath.Join(a.currentPath.String(), base.String()))
	}
}

func (a *archive) exit() {
	if a.currentPath == "" {
		return
	}
	parts := strings.Split(a.currentPath.String(), "/")
	if len(parts) == 1 {
		a.currentPath = ""
	}
	a.currentPath = m.Path(filepath.Join(parts[:len(parts)-1]...))
}

func (a *archive) mouseTarget(cmd any) {
	switch cmd := cmd.(type) {
	case m.SelectFile:
		a.selectFile(cmd)

	case m.SelectFolder:
		a.currentPath = m.Path(cmd)

	case m.SortColumn:
		folder := a.currentFolder()
		folder.selectSortColumn(cmd)
		folder.makeSelectedVisible(a.fileTreeLines)
	}
}

func (c *controller) tab() {
	panic("IMPLEMENT c.tab()")
	// selected := c.archive.currentFolder().selected()
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
	// folder := c.archive.currentFolder()
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

func (c *controller) handleHashingProgress(event m.HashingProgress) {
	archive := c.archives[event.Root]
	archive.fileHashedSize = event.Hashed
	folder := archive.folders[event.Path]
	file := folder.files[event.Base]
	file.SetState(m.Hashing)
	file.Hashed = event.Hashed

	// c.archives[event.Root].progressInfo = &progressInfo{
	// 	tab:           " Hashing",
	// 	value:         float64(archive.totalHashedSize+uint64(archive.fileHashedSize)) / float64(archive.totalSize),
	// 	speed:         archive.speed,
	// 	timeRemaining: archive.timeRemaining,
	// }

	// archive.parents(file, func(file *m.Folder) {
	// 	file.SetState(m.Hashing)
	// 	file.Hashed = event.Hashed
	// })
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
	panic("IMPLEMENT c.resolveSelected()")
	// c.resolveFile(c.archive.currentFolder().selectedEntry)
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
	panic("IMPLEMENT c.resolveFolder()")
	// for _, entry := range c.archive.folders[path].entries {
	// 	switch entry := entry.(type) {
	// 	case *m.File:
	// 		if entry.State() == m.Divergent && entry.Counts[c.archive.idx] == 1 {
	// 			c.resolveFile(entry)
	// 		}
	// 	case *m.Folder:
	// 		c.resolveFolder(m.Path(entry.Name.String()))
	// 	}
	// }
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

func (c *controller) clearName(file *m.File) {
	if c.nameCollidesWithPath(file.Name) {
		var newName m.Name
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

func (f *folder) selectFirst() {
	f.selectedBase = ""
	f.selectedIdx = 0
}

func (f *folder) selectLast() {
	f.selectedBase = ""
	f.selectedIdx = f.entries - 1
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
	log.Printf("moveSelection: selectedIdx: %d", f.selectedIdx)
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

func (a *archive) open() {
	id := m.Id{Root: a.root, Name: m.Name{Path: a.currentPath, Base: a.currentFolder().selectedBase}}
	exec.Command("open", id.String()).Start()
}

func (a *archive) revealInFinder() {
	id := m.Id{Root: a.root, Name: m.Name{Path: a.currentPath, Base: a.currentFolder().selectedBase}}
	exec.Command("open", "-R", id.String()).Start()
}

func (a *archive) selectFile(cmd m.SelectFile) {
	folder := a.currentFolder()
	if folder.selectedBase == cmd.Base && time.Since(folder.lastMouseEventTime).Seconds() < 0.5 {
		a.open()
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
	if f.sortAscending[f.sortColumn] {
		switch f.sortColumn {
		case m.SortByName:
			f.cmpFunc = cmpByAscendingName
		case m.SortByTime:
			f.cmpFunc = cmpByAscendingTime
		case m.SortBySize:
			f.cmpFunc = cmpByAscendingSize
		}
	} else {
		switch f.sortColumn {
		case m.SortByName:
			f.cmpFunc = cmpByDescendingName
		case m.SortByTime:
			f.cmpFunc = cmpByDescendingTime
		case m.SortBySize:
			f.cmpFunc = cmpByDescendingSize
		}
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
