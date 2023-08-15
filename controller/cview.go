package controller

import (
	m "arc/model"
	v "arc/view"
	"slices"
	"strings"
)

func (c *controller) view() *v.View {
	archive := c.archive
	currentFolder := archive.currentFolder()
	view := &v.View{
		Archive:   archive.root,
		Path:      archive.currentPath,
		OffsetIdx: currentFolder.offsetIdx,
	}

	subFolders := map[m.Base]m.Entry{}
	for path, folder := range archive.folders {
		var totalSize, progress uint64
		switch archive.state {
		case hashing:
			view.Progress = &v.Progress{
				Tab: " Hashing",
			}
			for _, file := range folder.files {
				totalSize += file.Size
				switch file.State() {
				case m.Hashing:
					progress += file.Progress
				case m.Hashed:
					progress += file.Size
				}
			}
		case copying:
			view.Progress = &v.Progress{
				Tab: " Copying",
			}
			for _, file := range folder.files {
				totalSize += file.Size
				switch file.State() {
				case m.Pending:
					totalSize += file.Size
					progress += file.Size
				case m.Copying:
					totalSize += file.Size
					progress += file.Progress
				}
			}
		default:
			view.Progress = nil
		}
		if path == archive.currentPath {
			archive.populateFiles(view, folder)
		} else if strings.HasPrefix(path.String(), archive.currentPath.String()) {
			archive.populateSubFolder(view, folder, subFolders)
		}
	}
	for _, subFolder := range subFolders {
		view.Entries = append(view.Entries, subFolder)
	}

	currentFolder.entries = len(view.Entries)

	slices.SortFunc(view.Entries, currentFolder.cmpFunc)

	if archive.state == scanning || len(view.Entries) == 0 {
		return view
	}

	validSelected := false
	for idx, entry := range view.Entries {
		if currentFolder.selectedBase == entry.Meta().Base {
			currentFolder.selectedIdx = idx
			validSelected = true
			break
		}
	}
	if !validSelected {
		if currentFolder.selectedIdx >= len(view.Entries) {
			currentFolder.selectedIdx = len(view.Entries)
		}
		if currentFolder.selectedIdx < 0 {
			currentFolder.selectedIdx = 0
		}

		currentFolder.selectedBase = view.Entries[currentFolder.selectedIdx].Meta().Base
	}
	view.SelectedBase = currentFolder.selectedBase

	return view
}

func (a *archive) populateFiles(view *v.View, folder *folder) {
	for _, file := range folder.files {
		view.Entries = append(view.Entries, file)
	}
}

func (a *archive) populateSubFolder(view *v.View, folder *folder, subFolders map[m.Base]m.Entry) {
	var currentPathParts []string
	if a.currentPath != "" {
		currentPathParts = strings.Split(a.currentPath.String(), "/")
	}
	folderPathParts := strings.Split(folder.path.String(), "/")
	base := m.Base(folderPathParts[len(currentPathParts)])
	entry, ok := subFolders[base]
	if !ok {
		entry = m.NewFolder(
			m.Meta{
				Id: m.Id{Root: a.root, Name: m.Name{Path: a.currentPath, Base: base}},
			},
			m.Scanned,
		)
		subFolders[base] = entry
	}
	for _, file := range folder.files {
		entry.Meta().Size += file.Size
		if entry.Meta().ModTime.Before(file.ModTime) {
			entry.Meta().ModTime = file.ModTime
		}
		entry.SetState(entry.State().Merge(file.State()))
	}
}
