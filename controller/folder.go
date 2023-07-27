package controller

import m "arc/model"

type folder struct {
	entries       map[m.Base]*m.File
	selectedEntry *m.File
	offsetIdx     int
	sortColumn    m.SortColumn
	sortAscending []bool
}
