package controller

import (
	m "arc/model"
	"log"
)

func (c *controller) renameFile(meta *file) {
	folder := c.archives[meta.Root].getFolder(meta.Path)
	newName := uniqueName(folder, meta.Name)
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

func uniqueName(folder *folder, name m.Path) m.Path {
	for i := 1; ; i++ {
		name = newSuffix(name, i)
		if _, ok := folder.files[name.Base]; !ok {
			break
		}
	}
	return name
}
