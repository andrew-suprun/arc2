package controller

import (
	m "arc/model"
	"fmt"
	"log"
	"strings"
)

func (c *controller) renameFile(meta *file) {
	folder := c.archives[meta.Root].getFolder(meta.Path)
	newName := newName(folder, meta.Name)
	log.Printf("renameEntry: >>> from: %q, to: %q", meta.Id, newName)
	c.fs.Send(m.RenameFile{
		Hash: meta.Hash,
		From: meta.Id,
		To:   newName,
	})

	delete(folder.files, meta.Base)
	meta.Name = newName
	meta.State = m.Resolved
	folder.files[meta.Base] = meta
}

func (c *controller) renameFolder(archive *archive, path m.Path, base m.Base) {
}

func newName(folder *folder, name m.Name) m.Name {
	for i := 1; ; i++ {
		name = newSuffix(name, i)
		if _, ok := folder.files[name.Base]; !ok {
			break
		}
	}
	return name
}

func newSuffix(name m.Name, idx int) m.Name {
	parts := strings.Split(name.Base.String(), ".")

	var part string
	if len(parts) == 1 {
		part = stripIdx(parts[0])
	} else {
		part = stripIdx(parts[len(parts)-2])
	}
	var newBase m.Base
	if len(parts) == 1 {
		newBase = m.Base(fmt.Sprintf("%s%c%d", part, '`', idx))
	} else {
		parts[len(parts)-2] = fmt.Sprintf("%s%c%d", part, '`', idx)
		newBase = m.Base(strings.Join(parts, "."))
	}
	return m.Name{Path: name.Path, Base: newBase}
}

type stripIdxState int

const (
	expectDigit stripIdxState = iota
	expectDigitOrBacktick
)

func stripIdx(name string) string {
	state := expectDigit
	i := len(name) - 1
	for ; i >= 0; i-- {
		ch := name[i]
		if ch >= '0' && ch <= '9' && (state == expectDigit || state == expectDigitOrBacktick) {
			state = expectDigitOrBacktick
		} else if ch == '`' && state == expectDigitOrBacktick {
			return name[:i]
		} else {
			return name
		}
	}
	return name
}
