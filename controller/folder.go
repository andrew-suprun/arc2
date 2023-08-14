package controller

import (
	m "arc/model"
	"fmt"
	"strings"
	"time"
)

type folder struct {
	path               m.Path
	files              map[m.Base]*m.File
	selectedBase       m.Base
	entries            int
	selectedIdx        int
	offsetIdx          int
	sortColumn         m.SortColumn
	cmpFunc            func(m.Entry, m.Entry) int
	sortAscending      []bool
	lastMouseEventTime time.Time
}

func newFolder(path m.Path) *folder {
	return &folder{
		path:          path,
		files:         map[m.Base]*m.File{},
		sortAscending: []bool{true, false, false},
		cmpFunc:       cmpByAscendingName,
	}
}

// func (f *folder) insertEntry(entry m.Entry) {
// idx, _ := slices.BinarySearchFunc(f.entries, entry, f.cmpFunc)
// f.entries = slices.Insert(f.entries, idx, entry)
// }

// func (f *folder) deleteEntry(base m.Base) m.Entry {
// if idx := slices.IndexFunc(f.entries, func(entry m.Entry) bool { return entry.Meta().Base == base }); idx >= 0 {
// 	result := f.entries[idx]
// 	f.entries = slices.Delete(f.entries, idx, idx+1)
// 	return result
// }
// log.Fatalf("Deleting non-existing entry: folder: %q, base: %q", f.path, base)
// return nil
// }

func (f *folder) printTo(buf *strings.Builder) {
	fmt.Fprintf(buf, "    Folder: %q\n", f.path)
	fmt.Fprintf(buf, "      Selected Idx: %d\n", f.selectedIdx)
	fmt.Fprintf(buf, "      Offset Idx: %d\n", f.offsetIdx)
	fmt.Fprintln(buf, "      Entries:")
	for _, entry := range f.files {
		fmt.Fprintf(buf, "        %s,\n", entry)
	}
}

// func (f *folder) selected() m.Entry {
// 	if len(f.entries) == 0 {
// 		return nil
// 	}
// 	if f.selectedIdx >= len(f.entries) {
// 		f.selectedIdx = len(f.entries) - 1
// 	}
// 	return f.entries[f.selectedIdx]
// }
