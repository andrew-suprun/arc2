package controller

import (
	m "arc/model"
	"log"
)

func (c *controller) renameFile(meta *file, newName m.Name) {
	log.Printf("renameEntry: >>> from: %q, to: %q", meta.Id, newName)
	c.fs.Send(m.RenameFile{
		Hash: meta.Hash,
		From: meta.Id,
		To:   newName,
	})

	delete(c.archives[meta.Root].folders[meta.Path].files, meta.Base)
	meta.Name = newName
	meta.State = m.Resolved
	c.archives[meta.Root].getFolder(meta.Path).files[meta.Base] = meta
}
