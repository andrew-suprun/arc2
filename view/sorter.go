package view

import (
	"cmp"
	"slices"
	"strings"
)

func (v *View) Sort(sortColumn SortColumn, sortAscending bool) {
	var cmp func(a, b Entry) int
	switch sortColumn {
	case SortByName:
		cmp = cmpByAscendingName
	case SortByTime:
		cmp = cmpByAscendingTime
	case SortBySize:
		cmp = cmpByAscendingSize
	}
	slices.SortFunc(v.Entries, cmp)
	if !sortAscending {
		slices.Reverse(v.Entries)
	}
}

func cmpByName(a, b Entry) int {
	return cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
}

func cmpBySize(a, b Entry) int {
	return cmp.Compare(a.Size, b.Size)
}

func cmpByTime(a, b Entry) int {
	if a.ModTime.Before(b.ModTime) {
		return -1
	} else if b.ModTime.Before(a.ModTime) {
		return 1
	}
	return 0
}

func cmpByAscendingName(a, b Entry) int {
	result := cmpByName(a, b)
	if result != 0 {
		return result
	}

	result = cmpBySize(a, b)
	if result != 0 {
		return result
	}
	return cmpByTime(a, b)
}

func cmpByDescendingName(a, b Entry) int {
	return cmpByAscendingName(b, a)
}

func cmpByAscendingTime(a, b Entry) int {
	result := cmpByTime(a, b)
	if result != 0 {
		return result
	}

	result = cmpByName(a, b)
	if result != 0 {
		return result
	}

	return cmpBySize(a, b)
}

func cmpByDescendingTime(a, b Entry) int {
	return cmpByAscendingTime(b, a)
}

func cmpByAscendingSize(a, b Entry) int {
	result := cmpBySize(a, b)
	if result != 0 {
		return result
	}

	result = cmpByName(a, b)
	if result != 0 {
		return result
	}

	return cmpByTime(a, b)
}

func cmpByDescendingSize(a, b Entry) int {
	return cmpByAscendingSize(b, a)
}
