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
	byHash   map[string][]*file

	fs         m.FS
	screenSize w.Size

	hashed              int
	copySize            uint64
	totalCopiedSize     uint64
	fileCopiedSize      uint64
	prevCopied          uint64
	copySpeed           float64
	timeRemaining       time.Duration
	lastMouseEventTime  time.Time
	fileTreeLines       int
	makeSelectedVisible bool
	errors              []m.Error

	quit bool
}

type archive struct {
	idx           int
	rootFolder    *folder
	currentPath   []string
	state         archiveState
	totalSize     uint64
	timeRemaining time.Duration
	scanned       bool
	progress      v.Progress
}

type meta struct {
	root     string
	name     string
	parent   *folder
	size     uint64
	modTime  time.Time
	state    v.State
	progress v.Progress
}

type folder struct {
	meta
	children      map[string]*folder
	files         map[string]*file
	selectedName  string
	selectedIdx   int
	offsetIdx     int
	sortColumn    v.SortColumn
	sortAscending []bool
}

type file struct {
	meta
	hash   string
	counts []int
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

func parsePath(strPath string) []string {
	return strings.Split(string(strPath), "/")
}

func parseName(strPath string) ([]string, string) {
	path := parsePath(strPath)
	base := path[len(path)-1]
	return path[:len(path)-1], base
}
