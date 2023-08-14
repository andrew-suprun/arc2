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

	state           archiveState
	totalSize       uint64
	totalHashedSize uint64
	fileHashedSize  uint64
	prevHashedSize  uint64
	speed           float64
	timeRemaining   time.Duration
	progressInfo    *progressInfo
	fileTreeLines   int
	scanned         bool
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

// TODO: ??? Need it?
// func (a *archive) clearPath(path m.Path) {
// log.Printf("clearPath: >>> root: %q, path: %q", a.root, path)
// defer log.Printf("clearPath: <<< root: %q, path: %q", a.root, path)

// parentName := path.ParentName()
// if parentName.Base == "." {
// 	return
// }
// folder := a.getFolder(parentName.Path)
// for _, entry := range folder.entries {
// 	if entry.Meta().Base == parentName.Base {
// 		file, ok := entry.(*m.File)
// 		if !ok {
// 			return
// 		}

// 		newBase := folder.uniqueName(entry.Meta().Base)
// 		newName := m.Name{Path: entry.Meta().Path, Base: newBase}
// 		a.renameEntry(file, newName)
// 		entry.SetState(m.Pending)
// 	}
// }
// a.clearPath(parentName.Path)
// }

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
	file.SetState(m.Pending)
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
