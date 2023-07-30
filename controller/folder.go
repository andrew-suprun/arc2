package controller

import (
	m "arc/model"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type folder struct {
	path               m.Path
	entries            []*m.File
	selectedIdx        int
	offsetIdx          int
	sortColumn         m.SortColumn
	sortAscending      []bool
	lastMouseEventTime time.Time
	needsSorting       bool
}

func (f *folder) printTo(buf *strings.Builder) {
	fmt.Fprintf(buf, "    Folder: %q\n", f.path)
	fmt.Fprintln(buf, "      Entries:")
	for _, entry := range f.entries {
		fmt.Fprintf(buf, "        %s,\n", entry)
	}
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

func (f *folder) revealInFinder() {
	exec.Command("open", "-R", f.selected().Id.String()).Start()
}

func (f *folder) selectFile(cmd m.SelectFile) {
	selected := f.selected()
	if selected.Id == m.Id(cmd) && time.Since(f.lastMouseEventTime).Seconds() < 0.5 {
		f.open()
	} else {
		for idx := range f.entries {
			if f.entries[idx].Base == cmd.Base {
				f.selectedIdx = idx
				break
			}
		}
	}
	f.lastMouseEventTime = time.Now()
}

func (f *folder) selectSortColumn(cmd m.SortColumn) {
	if cmd == f.sortColumn {
		f.sortAscending[f.sortColumn] = !f.sortAscending[f.sortColumn]
	} else {
		f.sortColumn = cmd
	}
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
			if entry.Base == newBase {
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
