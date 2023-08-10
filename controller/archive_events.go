package controller

import (
	m "arc/model"
	"path/filepath"
	"strings"
)

func (a *archive) fileScanned(file *m.File) {
	a.addFile(file)
	a.totalSize += file.Size
}

func (a *archive) archiveScanned() {
	// TODO Need this?
}

func (a *archive) addFile(file *m.File) {
	a.insertEntry(file)
	name := file.ParentName()
	for name.Base != "." {
		parentFolder := a.getFolder(name.Path)
		entry := parentFolder.entry(name.Base)

		if folderEntry, ok := entry.(*m.Folder); ok {
			folderEntry.Size += file.Size
			if folderEntry.ModTime.Before(file.ModTime) {
				folderEntry.ModTime = file.ModTime
			}
		} else {
			folderEntry := m.NewFolder(m.Meta{
				Id: m.Id{
					Root: file.Root,
					Name: name,
				},
				Size:    file.Size,
				ModTime: file.ModTime,
			},
				m.Scanned,
			)
			a.insertEntry(folderEntry)

		}

		name = name.Path.ParentName()
	}
}

func (a *archive) insertEntry(entry m.Entry) {
	folder := a.getFolder(entry.Meta().Path)
	folder.insertEntry(entry)
}

func (a *archive) fileHashedEvent(event m.FileHashed) {
	folder := a.getFolder(event.Path)
	file := folder.entry(event.Base).(*m.File)
	file.Hash = event.Hash
	file.SetState(m.Resolved)

	a.updateFolderStates("")

	a.parents(file, func(parent *m.Folder) {
		parent.Hashed = 0
		parent.TotalHashed += file.Size
	})

	a.totalHashed += file.Size
	a.fileHashed = 0
}

func (a *archive) enter() {
	entry := a.currentFolder().selected()
	if folder, ok := entry.(*m.Folder); ok {
		a.currentPath = m.Path(folder.Name.String())

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
		a.currentFolder().selectFile(cmd)

	case m.SelectFolder:
		a.currentPath = m.Path(cmd)

	case m.SortColumn:
		folder := a.currentFolder()
		folder.selectSortColumn(cmd)
		folder.makeSelectedVisible(a.fileTreeLines)
	}
}
