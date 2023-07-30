package controller

import (
	m "arc/model"
	"path/filepath"
	"strings"
)

func (a *archive) archiveScanned(event m.ArchiveScanned) {
	a.addFiles(event)

	for _, file := range event.Files {
		a.totalSize += file.Size
	}
}

func (a *archive) addFiles(event m.ArchiveScanned) {
	for _, file := range event.Files {
		a.addFile(&m.File{Meta: *file, State: m.Scanned})
	}
	a.currentPath = ""
}

func (a *archive) addFile(file *m.File) {
	a.addEntry(file)
	name := file.ParentName()
	for name.Base != "." {
		parentFolder := a.getFolder(name.Path)
		folderEntry := parentFolder.entry(name.Base)
		if folderEntry != nil {
			folderEntry.Size += file.Size
			if folderEntry.ModTime.Before(file.ModTime) {
				folderEntry.ModTime = file.ModTime
			}
		} else {
			folderEntry := &m.File{
				Meta: m.Meta{
					Id: m.Id{
						Root: file.Root,
						Name: name,
					},
					Size:    file.Size,
					ModTime: file.ModTime,
				},
				Kind:  m.FileFolder,
				State: m.Scanned,
			}
			a.addEntry(folderEntry)

		}

		name = name.Path.ParentName()
	}
}

func (a *archive) fileHashedEvent(event m.FileHashed) {
	folder := a.getFolder(event.Path)
	file := folder.entry(event.Base)
	file.Hash = event.Hash
	file.State = m.Resolved

	a.markDuplicates()
	a.updateFolderStates("")

	a.parents(file, func(parent *m.File) {
		parent.Hashed = 0
		parent.TotalHashed += file.Size
	})

	a.totalHashed += file.Size
	a.fileHashed = 0
}

func (a *archive) enter() {
	file := a.currentFolder().selected()
	if file != nil && file.Kind == m.FileFolder {
		a.currentPath = m.Path(file.Name.String())
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
		folder.needsSorting = true
		folder.makeSelectedVisible(a.fileTreeLines)
	}
}
