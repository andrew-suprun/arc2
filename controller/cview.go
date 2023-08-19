package controller

import (
	m "arc/model"
	v "arc/view"
	"strings"
)

func (c *controller) view() *v.View {
	archive := c.currArchive()
	folder := archive.currFolder()
	view := &v.View{
		Archive:       archive.root,
		Path:          archive.currentPath,
		SortColumn:    folder.sortColumn,
		SortAscending: folder.sortAscending[folder.sortColumn],
	}

	subFolders := map[m.Base]*v.Entry{}
	var totalSize, progress uint64
	for path, folder := range archive.folders {
		switch archive.state {
		case hashing:
			for _, file := range folder.files {
				totalSize += file.Size
				switch file.State {
				case m.Hashing:
					progress += file.progressDone
				case m.Hashed:
					progress += file.Size
				}
			}
		case copying:
			for _, file := range folder.files {
				totalSize += file.Size
				switch file.State {
				case m.Pending:
					totalSize += file.Size
					progress += file.Size
				case m.Copying:
					totalSize += file.Size
					progress += file.progressDone
				}
			}
		}
		if path == archive.currentPath {
			archive.populateFiles(view, folder)
		} else if strings.HasPrefix(path.String(), archive.currentPath.String()) {
			archive.populateSubFolder(view, folder, subFolders)
		}
	}

	switch archive.state {
	case hashing:
		view.Progress = &v.Progress{
			Tab:   " Hashing",
			Value: float64(progress) / float64(totalSize),
		}
	case copying:
		view.Progress = &v.Progress{
			Tab:   " Copying",
			Value: float64(progress) / float64(totalSize),
		}
	default:
		view.Progress = nil
	}

	for _, subFolder := range subFolders {
		view.Entries = append(view.Entries, subFolder)
	}

	view.Sort(folder.sortColumn, folder.sortAscending[folder.sortColumn])

	if archive.state == scanning || len(view.Entries) == 0 {
		return view
	}

	validSelected := false
	for idx, entry := range view.Entries {
		if folder.selectedId == entry.Id {
			folder.selectedIdx = idx
			validSelected = true
			break
		}
	}
	if !validSelected {
		if folder.selectedIdx >= len(view.Entries) {
			folder.selectedIdx = len(view.Entries) - 1
		}
		if folder.selectedIdx < 0 {
			folder.selectedIdx = 0
		}

		folder.selectedId = view.Entries[folder.selectedIdx].Id
	}
	view.SelectedId = folder.selectedId

	if folder.selectedIdx >= len(view.Entries) {
		folder.selectedIdx = len(view.Entries) - 1
	}
	if folder.selectedIdx < 0 {
		folder.selectedIdx = 0
	}

	if c.makeSelectedVisible {
		if folder.offsetIdx > folder.selectedIdx {
			folder.offsetIdx = folder.selectedIdx
		}
		if folder.offsetIdx < folder.selectedIdx+1-c.fileTreeLines {
			folder.offsetIdx = folder.selectedIdx + 1 - c.fileTreeLines
		}
		c.makeSelectedVisible = false
	}
	view.OffsetIdx = folder.offsetIdx

	return view
}

func (a *archive) populateFiles(view *v.View, folder *folder) {
	for _, file := range folder.files {
		view.Entries = append(view.Entries, &v.Entry{
			Meta:         file.Meta,
			Kind:         v.Regular,
			State:        file.State,
			ProgressSize: file.progressSize,
			ProgressDone: file.progressDone,
		})
	}
}

func (a *archive) populateSubFolder(view *v.View, folder *folder, subFolders map[m.Base]*v.Entry) {
	var currentPathParts []string
	if a.currentPath != "" {
		currentPathParts = strings.Split(a.currentPath.String(), "/")
	}
	folderPathParts := strings.Split(folder.path.String(), "/")
	base := m.Base(folderPathParts[len(currentPathParts)])
	entry, ok := subFolders[base]
	if !ok {
		entry = &v.Entry{
			Meta: m.Meta{
				Id: m.Id{Root: a.root, Name: m.Name{Path: a.currentPath, Base: base}},
			},
			Kind:  v.Folder,
			State: m.Scanned,
		}
		subFolders[base] = entry
	}
	for _, file := range folder.files {
		entry.Size += file.Size
		if entry.ModTime.Before(file.ModTime) {
			entry.ModTime = file.ModTime
		}
		entry.State = entry.State.Merge(file.State)
		entry.ProgressSize += file.progressSize
		entry.ProgressDone += file.progressDone
	}
}
