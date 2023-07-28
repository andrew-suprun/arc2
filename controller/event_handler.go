package controller

import (
	m "arc/model"
	w "arc/widgets"
	"log"
)

func (c *controller) handleEvent(event any) {
	if event == nil {
		return
	}
	switch event := event.(type) {
	case m.ArchiveScanned:
		c.archiveScanned(event)

	case m.FileHashed:
		c.fileHashed(event)

	case m.ArchiveHashed:
		c.archiveHashed(event)

	case m.FileDeleted:
		// c.fileDeleted(event)

	case m.FileRenamed:
		// c.fileRenamed(event)

	case m.FileCopied:
		// c.fileCopied(event)

	case m.HashingProgress:
		c.handleHashingProgress(event)

	case m.CopyingProgress:
		c.handleCopyingProgress(event)

	case m.Tick:
		c.handleTick(event)

	case m.ScreenSize:
		c.screenSize = w.Size{Width: event.Width, Height: event.Height}

	case m.SelectArchive:
		if event.Idx < len(c.roots) {
			c.archive = c.archives[c.roots[event.Idx]]
		}

	case m.Enter:
		// c.enter()

	case m.Open:
		// c.open()

	case m.Exit:
		// c.exit()

	case m.RevealInFinder:
		// c.revealInFinder()

	case m.MoveSelection:
		c.archive.currentFolder().moveSelection(event.Lines)

	case m.SelectFirst:
		c.archive.currentFolder().selectFirst()

	case m.SelectLast:
		c.archive.currentFolder().selectLast()

	case m.Scroll:
		// c.shiftOffset(event.Lines)

	case m.MouseTarget:
		// c.mouseTarget(event.Command)

	case m.PgUp:
		folder := c.archive.currentFolder()
		folder.moveSelection(-c.archive.fileTreeLines)
		folder.makeSelectionVisible()

	case m.PgDn:
		folder := c.archive.currentFolder()
		folder.moveSelection(c.archive.fileTreeLines)
		folder.makeSelectionVisible()

	case m.Tab:
		// c.tab()

	case m.KeepOne:
		// c.keepSelected()

	case m.KeepAll:
		// TODO: Implement, maybe?

	case m.Delete:
		// folder := c.currentFolder()
		// c.deleteFile(folder.selectedEntry)

	case m.Error:
		// log.Printf("### Error: %s", event)
		// c.Errors = append(c.Errors, event)

	case m.Quit:
		c.quit = true

	case m.Debug:
		// log.Println(c.screenString())

	default:
		log.Panicf("### unhandled event: %#v", event)
	}
}

func (c *controller) archiveScanned(event m.ArchiveScanned) {
	archive := c.archives[event.Root]
	archive.addFiles(event)

	for _, file := range event.Files {
		archive.totalSize += file.Size
	}
}

func (c *controller) fileHashed(event m.FileHashed) {
	archive := c.archives[event.Root]
	folder := archive.getFolder(event.Path)
	file := folder.getEntry(event.Base)
	file.Hash = event.Hash
	file.State = m.Hashed

	for _, folder := range archive.folders {
		for _, entry := range folder.entries {
			if entry != file && entry.Hash == file.Hash {
				file.State = m.Duplicate
				archive.parents(file, func(parent *m.File) {
					parent.State = m.Duplicate
				})
				entry.State = m.Duplicate
				archive.parents(entry, func(parent *m.File) {
					parent.State = m.Duplicate
				})
			}
		}
	}

	archive.parents(file, func(parent *m.File) {
		parent.State = file.State
		parent.Hashed = 0
		parent.TotalHashed += file.Size
	})

	archive.totalHashed += file.Size
	archive.fileHashed = 0
}

func (c *controller) archiveHashed(event m.ArchiveHashed) {
	archive := c.archives[event.Root]
	archive.progressInfo = nil

	for _, archive := range c.archives {
		if archive.progressInfo != nil {
			return
		}
	}

}

func (c *controller) handleHashingProgress(event m.HashingProgress) {
	archive := c.archives[event.Root]
	archive.fileHashed = event.Hashed
	folder := archive.folders[event.Path]
	file := folder.getEntry(event.Base)
	file.State = m.Hashing
	file.Hashed = event.Hashed

	c.archives[event.Root].progressInfo = &progressInfo{
		tab:           " Hashing",
		value:         float64(archive.totalHashed+uint64(archive.fileHashed)) / float64(archive.totalSize),
		speed:         archive.speed,
		timeRemaining: archive.timeRemaining,
	}

	archive.parents(file, func(file *m.File) {
		file.State = m.Hashing
		file.Hashed = event.Hashed
	})

}

func (c *controller) handleCopyingProgress(event m.CopyingProgress) {
	c.fileCopiedSize = uint64(event)
	info := &progressInfo{
		tab:           " Copying",
		value:         float64(c.totalCopiedSize+uint64(c.fileCopiedSize)) / float64(c.copySize),
		speed:         c.copySpeed,
		timeRemaining: c.timeRemaining,
	}
	for _, archive := range c.archives {
		archive.progressInfo = info
	}
}

func (c *controller) analyzeAbsentFiles() {
	log.Panic("Implement: controller.analyzeAbsentFiles()")
}
