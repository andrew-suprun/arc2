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
	sortAscending      []bool
	lastMouseEventTime time.Time
}

func newFolder(path m.Path) *folder {
	return &folder{
		path:          path,
		files:         map[m.Base]*m.File{},
		sortAscending: []bool{true, false, false},
	}
}

func (f *folder) printTo(buf *strings.Builder) {
	fmt.Fprintf(buf, "    Folder: %q\n", f.path)
	fmt.Fprintf(buf, "      Selected Idx: %d\n", f.selectedIdx)
	fmt.Fprintf(buf, "      Offset Idx: %d\n", f.offsetIdx)
	fmt.Fprintln(buf, "      Entries:")
	for _, entry := range f.files {
		fmt.Fprintf(buf, "        %s,\n", entry)
	}
}
