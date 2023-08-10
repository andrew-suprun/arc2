package controller

import (
	m "arc/model"
	"fmt"
	"log"
	"os/exec"
	"slices"
	"strings"
	"time"
)

type folder struct {
	path               m.Path
	entries            []m.Entry
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
		sortAscending: []bool{true, false, false},
		cmpFunc:       cmpByAscendingName,
	}
}

func (f *folder) insertEntry(entry m.Entry) {
	idx, _ := slices.BinarySearchFunc(f.entries, entry, f.cmpFunc)
	f.entries = slices.Insert(f.entries, idx, entry)
}

func (f *folder) deleteEntry(base m.Base) m.Entry {
	if idx := slices.IndexFunc(f.entries, func(entry m.Entry) bool { return entry.Meta().Base == base }); idx >= 0 {
		result := f.entries[idx]
		f.entries = slices.Delete(f.entries, idx, idx+1)
		return result
	}
	log.Fatalf("Deleting non-existing entry: folder: %q, base: %q", f.path, base)
	return nil
}

func (f *folder) printTo(buf *strings.Builder) {
	fmt.Fprintf(buf, "    Folder: %q\n", f.path)
	fmt.Fprintf(buf, "      Selected Idx: %d\n", f.selectedIdx)
	fmt.Fprintf(buf, "      Offset Idx: %d\n", f.offsetIdx)
	fmt.Fprintln(buf, "      Entries:")
	for _, entry := range f.entries {
		fmt.Fprintf(buf, "        %s,\n", entry)
	}
}

func (f *folder) selected() m.Entry {
	if len(f.entries) == 0 {
		return nil
	}
	if f.selectedIdx >= len(f.entries) {
		f.selectedIdx = len(f.entries) - 1
	}
	return f.entries[f.selectedIdx]
}

func (f *folder) entry(base m.Base) m.Entry {
	idx := slices.IndexFunc(f.entries, func(entry m.Entry) bool { return entry.Meta().Base == base })
	if idx >= 0 {
		return f.entries[idx]
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
	exec.Command("open", f.selected().Meta().Id.String()).Start()
}

func (f *folder) revealInFinder() {
	exec.Command("open", "-R", f.selected().Meta().Id.String()).Start()
}

func (f *folder) selectFile(cmd m.SelectFile) {
	selected := f.selected()
	if selected.Meta().Id == m.Id(cmd) && time.Since(f.lastMouseEventTime).Seconds() < 0.5 {
		f.open()
	} else {
		f.selectedIdx = slices.IndexFunc(f.entries, func(e m.Entry) bool { return e.Meta().Base == cmd.Base })
	}
	f.lastMouseEventTime = time.Now()
}

func (f *folder) selectSortColumn(cmd m.SortColumn) {
	if cmd == f.sortColumn {
		f.sortAscending[f.sortColumn] = !f.sortAscending[f.sortColumn]
	} else {
		f.sortColumn = cmd
	}
	if f.sortAscending[f.sortColumn] {
		switch f.sortColumn {
		case m.SortByName:
			f.cmpFunc = cmpByAscendingName
		case m.SortByTime:
			f.cmpFunc = cmpByAscendingTime
		case m.SortBySize:
			f.cmpFunc = cmpByAscendingSize
		}
	} else {
		switch f.sortColumn {
		case m.SortByName:
			f.cmpFunc = cmpByDescendingName
		case m.SortByTime:
			f.cmpFunc = cmpByDescendingTime
		case m.SortBySize:
			f.cmpFunc = cmpByDescendingSize
		}
	}
	f.sort()
}

func (f *folder) makeSelectedVisible(fileTreeLines int) {
	if f.offsetIdx > f.selectedIdx {
		f.offsetIdx = f.selectedIdx
	}
	if f.offsetIdx < f.selectedIdx+1-fileTreeLines {
		f.offsetIdx = f.selectedIdx + 1 - fileTreeLines
	}
}

func (f *folder) uniqueName(base m.Base) m.Base {
	parts := strings.Split(base.String(), ".")

	var part string
	if len(parts) == 1 {
		part = stripIdx(parts[0])
	} else {
		part = stripIdx(parts[len(parts)-2])
	}
outer:
	for idx := 1; ; idx++ {
		var newBase m.Base
		if len(parts) == 1 {
			newBase = m.Base(fmt.Sprintf("%s%c%d", part, '\x60', idx))
		} else {
			parts[len(parts)-2] = fmt.Sprintf("%s%c%d", part, '\x60', idx)
			newBase = m.Base(strings.Join(parts, "."))
		}
		for _, entry := range f.entries {
			if entry.Meta().Base == newBase {
				continue outer
			}
		}
		return newBase
	}
}

type stripIdxState int

const (
	expectDigit stripIdxState = iota
	expectDigitOrBacktick
)

func stripIdx(name string) string {
	state := expectDigit
	i := len(name) - 1
	for ; i >= 0; i-- {
		ch := name[i]
		if ch >= '0' && ch <= '9' && (state == expectDigit || state == expectDigitOrBacktick) {
			state = expectDigitOrBacktick
		} else if ch == '\x60' && state == expectDigitOrBacktick {
			return name[:i]
		} else {
			return name
		}
	}
	return name
}
