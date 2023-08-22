package controller

import (
	m "arc/model"
	v "arc/view"
	"log"
)

func (c *controller) renameFile(meta *file) {
	folder := meta.parent
	oldFullName := meta.fullName()
	delete(folder.files, meta.name)
	meta.name = folder.uniqueName(meta.name)
	newFulleName := meta.fullName()
	folder.files[meta.name] = meta
	meta.state = v.Resolved

	log.Printf("renameEntry: >>> from: %q, to: %q", oldFullName, newFulleName)
	c.fs.Send(m.RenameFile{
		Root: meta.root,
		From: oldFullName,
		To:   newFulleName,
	})

}

func (c *controller) renameFolder(archive *archive, folder *folder) {
	// TODO
}
