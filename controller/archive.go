package controller

import (
	m "arc/model"
	"time"
)

type archive struct {
	root        m.Root
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

func newArchive(root m.Root, shared *shared) *archive {
	return &archive{
		root:    root,
		folders: map[m.Path]*folder{},
		shared:  shared,
	}
}

func (a *archive) archiveScanned(event m.ArchiveScanned) {
	a.addFiles(event)
	a.sort()

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
	folder := a.getFolder(file.Path)
	folder.addEntry(file)
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
			parentFolder.addEntry(folderEntry)

		}

		name = name.Path.ParentName()
	}
}

func (a *archive) sort() {
	for _, folder := range a.folders {
		folder.sort()
	}
}

func (a *archive) currentFolder() *folder {
	return a.getFolder(a.currentPath)
}

func (a *archive) getFolder(path m.Path) *folder {
	pathFolder, ok := a.folders[path]
	if !ok {
		pathFolder = &folder{
			sortAscending: []bool{true, false, false},
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

func (a *archive) markDuplicates() {
	hashes := map[m.Hash]int{}
	for _, folder := range a.folders {
		for _, entry := range folder.entries {
			if entry.Hash != "" {
				hashes[entry.Hash]++
			}
		}
	}

	for _, folder := range a.folders {
		for _, entry := range folder.entries {
			if hashes[entry.Hash] > 1 {
				entry.State = m.Duplicate
			}
		}
	}

	a.duplicateFiles = 0
	for _, count := range hashes {
		if count > 1 {
			a.duplicateFiles++
		}
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
