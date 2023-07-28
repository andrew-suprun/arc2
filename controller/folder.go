package controller

import (
	m "arc/model"
	"os/exec"
)

type folder struct {
	entries       []*m.File
	selectedIdx   int
	offsetIdx     int
	sortColumn    m.SortColumn
	sortAscending []bool
}

func (f *folder) addEntry(entry *m.File) {
	f.entries = append(f.entries, entry)
}

func (f *folder) selected() *m.File {
	if f.selectedIdx < len(f.entries) {
		return f.entries[f.selectedIdx]
	}
	return nil
}

func (f *folder) entry(base m.Base) *m.File {
	for _, entry := range f.entries {
		if entry.Base == base {
			return entry
		}
	}
	return nil
}

func (f *folder) selectFirst() {
	f.selectedIdx = 0
}

func (f *folder) selectLast() {
	f.selectedIdx = len(f.entries) - 1
}

func (f *folder) moveSelection(lines int) {
	f.selectedIdx += lines

	if f.selectedIdx >= len(f.entries) {
		f.selectedIdx = len(f.entries) - 1
	}
	if f.selectedIdx < 0 {
		f.selectedIdx = 0
	}
}

func (f *folder) moveOffset(lines, fileTreeLines int) {
	f.offsetIdx += lines

	if f.offsetIdx >= len(f.entries)+1-fileTreeLines {
		f.offsetIdx = len(f.entries) - fileTreeLines
	}
	if f.offsetIdx < 0 {
		f.offsetIdx = 0
	}
}

func (f *folder) open() {
	exec.Command("open", f.selected().Id.String()).Start()
}
