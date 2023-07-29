package controller

import (
	m "arc/model"
	"log"
	"sort"
	"strings"
)

func (f *folder) sort() {
	selected := f.selected()
	log.Printf("sort:1: idx: %d, selected: %s", f.selectedIdx, selected)
	files := sliceBy(f.entries)
	var slice sort.Interface
	switch f.sortColumn {
	case m.SortByName:
		slice = sliceByName{sliceBy: files}
	case m.SortByTime:
		slice = sliceByTime{sliceBy: files}
	case m.SortBySize:
		slice = sliceBySize{sliceBy: files}
	}
	if !f.sortAscending[f.sortColumn] {
		slice = sort.Reverse(slice)
	}
	sort.Sort(slice)
	for idx, entry := range f.entries {
		if entry == selected {
			f.selectedIdx = idx
			break
		}
	}
	log.Printf("sort:2: idx: %d, selected: %s", f.selectedIdx, selected)
}

type sliceBy []*m.File

func (s sliceBy) Len() int {
	return len(s)
}

func (s sliceBy) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type sliceByName struct {
	sliceBy
}

func (s sliceByName) Less(i, j int) bool {
	iName := strings.ToLower(s.sliceBy[i].Base.String())
	jName := strings.ToLower(s.sliceBy[j].Base.String())
	if iName != jName {
		return iName < jName
	}

	iSize := s.sliceBy[i].Size
	jSize := s.sliceBy[j].Size
	if iSize != jSize {
		return iSize < jSize
	}

	return s.sliceBy[i].ModTime.Before(s.sliceBy[j].ModTime)
}

type sliceByTime struct {
	sliceBy
}

func (s sliceByTime) Less(i, j int) bool {
	iModTime := s.sliceBy[i].ModTime
	jModTime := s.sliceBy[j].ModTime
	if iModTime.Before(jModTime) {
		return true
	} else if iModTime.After(jModTime) {
		return false
	}

	iName := strings.ToLower(s.sliceBy[i].Base.String())
	jName := strings.ToLower(s.sliceBy[j].Base.String())
	if iName != jName {
		return iName < jName
	}

	return s.sliceBy[i].Size < s.sliceBy[j].Size
}

type sliceBySize struct {
	sliceBy
}

func (s sliceBySize) Less(i, j int) bool {
	iSize := s.sliceBy[i].Size
	jSize := s.sliceBy[j].Size
	if iSize != jSize {
		return iSize < jSize
	}

	iName := strings.ToLower(s.sliceBy[i].Base.String())
	jName := strings.ToLower(s.sliceBy[j].Base.String())
	if iName != jName {
		return iName < jName
	}

	return s.sliceBy[i].ModTime.Before(s.sliceBy[j].ModTime)
}
