package controller

import (
	m "arc/model"
	"cmp"
	"slices"
	"strings"
)

func cmpByName(a, b m.Entry) int {
	return cmp.Compare(strings.ToLower(a.Meta().Base.String()), strings.ToLower(b.Meta().Base.String()))
}

func cmpBySize(a, b m.Entry) int {
	return cmp.Compare(a.Meta().Size, b.Meta().Size)
}

func cmpByTime(a, b m.Entry) int {
	m1 := a.Meta()
	m2 := b.Meta()
	if m1.ModTime.Before(m1.ModTime) {
		return -1
	} else if m2.ModTime.Before(m1.ModTime) {
		return 1
	}
	return 0
}

func cmpByAscendingName(a, b m.Entry) int {
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

func cmpByDescendingName(a, b m.Entry) int {
	return cmpByAscendingName(b, a)
}

func cmpByAscendingTime(a, b m.Entry) int {
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

func cmpByDescendingTime(a, b m.Entry) int {
	return cmpByAscendingTime(b, a)
}

func cmpByAscendingSize(a, b m.Entry) int {
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

func cmpByDescendingSize(a, b m.Entry) int {
	return cmpByAscendingSize(b, a)
}

func (f *folder) sort() {
	slices.SortFunc(f.entries, f.cmpFunc)
}
