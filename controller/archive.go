package controller

import (
	"strings"
)

func (a *archive) printTo(buf *strings.Builder) {
	// TODO
}

func (a *archive) folder(path path) *folder {
	currFolder := a.rootFolder
	for _, name := range path[:len(path)-1] {
		currFolder = currFolder.children[name]
	}
	return currFolder
}

func (a *archive) curFolder() *folder {
	return a.folder(a.currentPath)
}
