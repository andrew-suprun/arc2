package controller

import (
	v "arc/view"
)

func (c *controller) view() *v.View {
	archive := c.curArchive()
	folder := archive.curFolder()
	view := &v.View{
		Archive: c.root,
		Path:    folder.path(),
	}

	archive.updateMetas()

	for _, sub := range folder.children {
		view.Entries = append(view.Entries, v.Entry{
			Name:     sub.name,
			Size:     sub.size,
			ModTime:  sub.modTime,
			Kind:     v.Folder,
			State:    sub.state,
			Progress: sub.progress,
		})
	}

	for _, file := range folder.files {
		view.Entries = append(view.Entries, v.Entry{
			Name:     file.name,
			Size:     file.size,
			ModTime:  file.modTime,
			Kind:     v.Folder,
			State:    file.state,
			Counts:   file.counts,
			Progress: file.progress,
		})
	}

	view.Sort(folder.sortColumn, folder.sortAscending[folder.sortColumn])
	if c.makeSelectedVisible {
		if folder.offsetIdx > folder.selectedIdx {
			folder.offsetIdx = folder.selectedIdx
		}
		if folder.offsetIdx < folder.selectedIdx+1-c.fileTreeLines {
			folder.offsetIdx = folder.selectedIdx + 1 - c.fileTreeLines
		}
		c.makeSelectedVisible = false
	}

	if archive.state == scanning || len(view.Entries) == 0 {
		return view
	}

	validSelected := false
	for idx, entry := range view.Entries {
		if folder.selectedName == entry.Name {
			folder.selectedIdx = idx
			validSelected = true
			break
		}
	}
	if !validSelected {
		if folder.selectedIdx >= len(view.Entries) {
			folder.selectedIdx = len(view.Entries)
		}
		if folder.selectedIdx < 0 {
			folder.selectedIdx = 0
		}

		folder.selectedName = view.Entries[folder.selectedIdx].Name
	}
	view.SelectedName = folder.selectedName

	return view
}
