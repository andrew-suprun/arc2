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

	totalSize      uint64
	totalHashed    uint64
	fileHashed     uint64
	prevHashed     uint64
	speed          float64
	timeRemaining  time.Duration
	progressInfo   *progressInfo
	pendingFiles   int
	duplicateFiles int
	absentFiles    int
	fileTreeLines  int
}

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

func (a *archive) removeEntry(name m.Name) *m.File {
	var removed *m.File
	folder := a.getFolder(name.Path)
	for idx, entry := range folder.entries {
		if entry.Base == name.Base {
			removed = entry
			if idx < len(folder.entries)-1 {
				folder.entries[idx] = folder.entries[len(folder.entries)-1]
			}
			folder.entries = folder.entries[:len(folder.entries)-1]
		}
	}
	folder.needsSorting = true
	return removed
}

func (a *archive) clearPath(path m.Path) {
	log.Printf("clearPath: >>> root: %q, path: %q", a.root, path)
	defer log.Printf("clearPath: <<< root: %q, path: %q", a.root, path)

	parentName := path.ParentName()
	if parentName.Base == "." {
		return
	}
	folder := a.getFolder(parentName.Path)
	for _, file := range folder.entries {
		if file.Base == parentName.Base {
			if file.Kind == m.FileFolder {
				return
			}

			newBase := folder.uniqueName(file.Base)
			newName := m.Name{Path: file.Path, Base: newBase}
			a.renameEntry(file, newName)
			file.State = m.Pending
		}
	}
	a.clearPath(parentName.Path)
}

func (a *archive) renameEntry(file *m.File, newName m.Name) {
	log.Printf("renameEntry: >>> from: %q, to: %q", file.Id, newName)
	defer log.Printf("renameEntry: <<< from: %q, to: %q", file.Id, newName)
	a.shared.fs.Send(m.RenameFile{
		Hash: file.Hash,
		From: file.Id,
		To:   newName,
	})

	if file.Path == newName.Path {
		folder := a.getFolder(file.Path)
		for _, entry := range folder.entries {
			if entry.Base == file.Base {
				entry.Base = newName.Base
				folder.needsSorting = true
				return
			}
		}
	}
	removed := a.removeEntry(file.Name)
	removed.Name = newName
	a.addFile(removed)
}

func (a *archive) currentFolder() *folder {
	return a.getFolder(a.currentPath)
}

func (a *archive) getFolder(path m.Path) *folder {
	pathFolder, ok := a.folders[path]
	if !ok {
		pathFolder = &folder{
			path:          path,
			sortAscending: []bool{true, false, false},
			needsSorting:  true,
		}
		a.folders[path] = pathFolder
	}
	return pathFolder
}

func (a *archive) parents(file *m.File, proc func(parent *m.File)) {
	name := file.ParentName()
	for name.Base != "." {
		proc(a.getFolder(name.Path).entry(name.Base))
		name = name.Path.ParentName()
	}
}

func (a *archive) updateFolderStates(path m.Path) m.State {
	state := m.Resolved
	folder := a.getFolder(path)
	for _, entry := range folder.entries {
		if entry.Kind == m.FileFolder {
			entry.State = a.updateFolderStates(m.Path(entry.Name.String()))
			state = state.Merge(entry.State)
		} else {
			state = state.Merge(entry.State)
		}
	}
	return state
}
