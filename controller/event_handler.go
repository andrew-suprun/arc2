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
	case m.FileScanned:
		c.archives[event.Root].fileScanned(event)

	case m.ArchiveScanned:
		c.archives[event.Root].archiveScanned()

	case m.FileHashed:
		c.archives[event.Root].fileHashedEvent(event)

	case m.ArchiveHashed:
		c.archiveHashed(event)

	case m.FileDeleted:
		c.fileDeleted(event)

	case m.FileRenamed:
		c.fileRenamed(event)

	case m.FileCopied:
		c.fileCopied(event)

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
		c.archive.enter()

	case m.Exit:
		c.archive.exit()

	case m.Open:
		c.archive.currentFolder().open()

	case m.RevealInFinder:
		c.archive.currentFolder().revealInFinder()

	case m.MoveSelection:
		folder := c.archive.currentFolder()
		folder.moveSelection(event.Lines)
		folder.makeSelectedVisible(c.archive.fileTreeLines)

	case m.SelectFirst:
		c.archive.currentFolder().selectFirst()

	case m.SelectLast:
		c.archive.currentFolder().selectLast()

	case m.Scroll:
		c.archive.currentFolder().moveOffset(event.Lines, c.archive.fileTreeLines)

	case m.MouseTarget:
		c.archive.mouseTarget(event.Command)

	case m.PgUp:
		folder := c.archive.currentFolder()
		folder.moveOffset(-c.archive.fileTreeLines, c.archive.fileTreeLines)
		folder.moveSelection(-c.archive.fileTreeLines)

	case m.PgDn:
		folder := c.archive.currentFolder()
		folder.moveOffset(c.archive.fileTreeLines, c.archive.fileTreeLines)
		folder.moveSelection(c.archive.fileTreeLines)

	case m.Tab:
		c.tab()
		c.archive.currentFolder().makeSelectedVisible(c.archive.fileTreeLines)

	case m.KeepOne:
		c.keepSelected()

	case m.KeepAll:
		// TODO: Implement, maybe?

	case m.Delete:
		// folder := c.currentFolder()
		// c.deleteFile(folder.selectedEntry)

	case m.Error:
		log.Printf("### Error: %s", event)
		c.errors = append(c.errors, event)

	case m.Quit:
		c.quit = true

	case m.DebugPrintState:
		log.Println(c.String())

	case m.DebugPrintRootWidget:
		log.Println(c.archive.rootWidget())

	default:
		log.Panicf("### unhandled event: %#v", event)
	}
}
