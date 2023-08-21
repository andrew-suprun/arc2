package controller

import (
	m "arc/model"
	v "arc/view"
	w "arc/widgets"
	"strings"
	"time"
)

type controller struct {
	roots    []string
	root     string
	archives map[string]*archive

	fs         m.FS
	screenSize w.Size
	frames     int
	prevTick   time.Time

	hashed             int
	copySize           uint64
	totalCopiedSize    uint64
	fileCopiedSize     uint64
	prevCopied         uint64
	copySpeed          float64
	timeRemaining      time.Duration
	lastMouseEventTime time.Time
	fileTreeLines      int
	errors             []m.Error

	quit bool
}

type archive struct {
	rootFolder    *folder
	currentPath   []string
	state         archiveState
	totalSize     uint64
	timeRemaining time.Duration
	progressInfo  *progressInfo
	scanned       bool
}

type archiveState int

const (
	scanning archiveState = iota
	hashing
	ready
	copying
)

type progressInfo struct {
	tab           string
	value         float64
	speed         float64
	timeRemaining time.Duration
}

type file struct {
	name     string
	folder   *folder
	size     uint64
	modTime  time.Time
	hash     string
	state    v.State
	progress uint64
}

type folder struct {
	name          string
	parent        *folder
	children      map[string]*folder
	files         map[string]*file
	size          uint64
	modTime       time.Time
	state         v.State
	selectedName  string
	selectedIdx   int
	offsetIdx     int
	sortColumn    v.SortColumn
	sortAscending []bool
}

func parsePath(strPath string) []string {
	return strings.Split(string(strPath), "/")
}

func parseName(strPath string) ([]string, string) {
	path := parsePath(strPath)
	base := path[len(path)-1]
	return path[:len(path)-1], base
}
