package controller

import (
	m "arc/model"
	"fmt"
	"log"
	"strings"
	"time"
)

type archive struct {
	root        m.Root
	idx         int
	folders     map[m.Path]*folder
	currentPath m.Path

	*shared

	state         archiveState
	totalSize     uint64
	speed         float64
	timeRemaining time.Duration
	progressInfo  *progressInfo
	fileTreeLines int
	scanned       bool
}

type archiveState int

const (
	scanning archiveState = iota
	hashing
	ready
	copying
)

type progressInfo struct {
	tab           string
	value         float64
	speed         float64
	timeRemaining time.Duration
}

func newArchive(root m.Root, idx int, shared *shared) *archive {
	return &archive{
		root:    root,
		idx:     idx,
		folders: map[m.Path]*folder{},
		shared:  shared,
	}
}

func (a *archive) printTo(buf *strings.Builder) {
	fmt.Fprintf(buf, "  Archive: %q\n", a.root)
	fmt.Fprintf(buf, "    Current Path: %q\n", a.currentPath)
	for _, folder := range a.folders {
		folder.printTo(buf)
	}
}

func (a *archive) renameEntry(file *m.File, newName m.Name) {
	log.Printf("renameEntry: >>> from: %q, to: %q", file.Id, newName)
	defer log.Printf("renameEntry: <<< from: %q, to: %q", file.Id, newName)
	a.shared.fs.Send(m.RenameFile{
		Hash: file.Hash,
		From: file.Id,
		To:   newName,
	})

	delete(a.folders[file.Path].files, file.Base)
	file.Name = newName
	file.SetState(m.Resolved)
	a.folders[file.Path].files[file.Base] = file
}

func (a *archive) currentFolder() *folder {
	return a.getFolder(a.currentPath)
}

func (a *archive) getFolder(path m.Path) *folder {
	pathFolder, ok := a.folders[path]
	if !ok {
		pathFolder = newFolder(path)
		a.folders[path] = pathFolder
	}
	return pathFolder
}

func (a *archive) parents(file *m.File, proc func(parent *m.Folder)) {
	panic("ERROR")
	// name := file.ParentName()
	// for name.Base != "." {
	// 	proc(a.getFolder(name.Path).entry(name.Base).(*m.Folder))
	// 	name = name.Path.ParentName()
	// }
}
