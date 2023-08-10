package controller

import (
	m "arc/model"
	"slices"
	"strings"
)

func cmpByName(one, two m.Entry) int {
	m1 := one.Meta()
	m2 := two.Meta()
	oneName := strings.ToLower(m1.Base.String())
	twoName := strings.ToLower(m2.Base.String())
	if oneName < twoName {
		return -1
	} else if oneName > twoName {
		return 1
	}
	return 0
}

func cmpBySize(one, two m.Entry) int {
	m1 := one.Meta()
	m2 := two.Meta()
	if m1.Size < m2.Size {
		return -1
	} else if m1.Size > m2.Size {
		return 1
	}
	return 0
}

func cmpByTime(one, two m.Entry) int {
	m1 := one.Meta()
	m2 := two.Meta()
	if m1.ModTime.Before(m1.ModTime) {
		return -1
	} else if m2.ModTime.Before(m1.ModTime) {
		return 1
	}
	return 0
}

func cmpByAscendingName(one, two m.Entry) int {
	result := cmpByName(one, two)
	if result != 0 {
		return result
	}

	result = cmpBySize(one, two)
	if result != 0 {
		return result
	}
	return cmpByTime(one, two)
}

func cmpByDescendingName(one, two m.Entry) int {
	return cmpByAscendingName(two, one)
}

func cmpByAscendingTime(one, two m.Entry) int {
	result := cmpByTime(one, two)
	if result != 0 {
		return result
	}

	result = cmpByName(one, two)
	if result != 0 {
		return result
	}

	return cmpBySize(one, two)
}

func cmpByDescendingTime(one, two m.Entry) int {
	return cmpByAscendingTime(two, one)
}

func cmpByAscendingSize(one, two m.Entry) int {
	result := cmpBySize(one, two)
	if result != 0 {
		return result
	}

	result = cmpByName(one, two)
	if result != 0 {
		return result
	}

	return cmpByTime(one, two)
}

func cmpByDescendingSize(one, two m.Entry) int {
	return cmpByAscendingSize(two, one)
}

func (f *folder) sort() {
	slices.SortFunc(f.entries, f.cmpFunc)
}
