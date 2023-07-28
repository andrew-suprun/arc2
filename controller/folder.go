package controller

import (
	m "arc/model"
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

func (f *folder) getSelected() *m.File {
	if f.selectedIdx < len(f.entries) {
		return f.entries[f.selectedIdx]
	}
	return nil
}

func (f *folder) getEntry(base m.Base) *m.File {
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

func (f *folder) makeSelectionVisible() {
	// TODO
}
