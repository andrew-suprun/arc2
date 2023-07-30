package controller

import (
	m "arc/model"
	"log"
	"path/filepath"
	"strings"
)

func (a *archive) enter() {
	file := a.currentFolder().selected()
	if file != nil && file.Kind == m.FileFolder {
		a.currentPath = m.Path(file.Name.String())
	}
}

func (a *archive) exit() {
	if a.currentPath == "" {
		return
	}
	parts := strings.Split(a.currentPath.String(), "/")
	if len(parts) == 1 {
		a.currentPath = ""
	}
	a.currentPath = m.Path(filepath.Join(parts[:len(parts)-1]...))
}

func (a *archive) mouseTarget(cmd any) {
	switch cmd := cmd.(type) {
	case m.SelectFile:
		a.currentFolder().selectFile(cmd)

	case m.SelectFolder:
		a.currentPath = m.Path(cmd)

	case m.SortColumn:
		folder := a.currentFolder()
		folder.selectSortColumn(cmd)
		folder.sort()
		folder.makeSelectedVisible(a.fileTreeLines)
	}
}

func (a *archive) keepFile(file *m.File) {
	folder := a.folders[file.Path]
	var sameName *m.File
	for _, entry := range folder.entries {
		if entry.Name == file.Name {
			sameName = entry
			break
		}
	}
	if sameName.Hash != file.Hash {
		newBase := folder.uniqueName(file.Base)
		newId := m.Id{Root: file.Root, Name: m.Name{Path: file.Path, Base: newBase}}
		log.Printf("keepFile: rename from: %q to %q", file.Id, newId)
		// TODO Finish

	}

	sameHash := []*m.File{}
	for _, folder := range a.folders {
		for _, entry := range folder.entries {
			if entry.Hash == file.Hash {
				sameHash = append(sameHash, entry)
			}
		}
	}
	var keep *m.File
	for _, entry := range sameHash {
		if entry.Name == file.Name {
			keep = entry
		}
	}
	if keep == nil {
		keep = sameHash[0]
	}

}
