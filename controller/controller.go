package controller

import (
	m "arc/model"
	"arc/stream"
	w "arc/widgets"
	"log"
	"time"
)

type controller struct {
	roots    []m.Root
	archives map[m.Root]*archive
	scanners map[m.Root]m.ArchiveScanner
	archive  *archive

	screenSize w.Size
	frames     int
	prevTick   time.Time

	copySize           uint64
	totalCopiedSize    uint64
	fileCopiedSize     uint64
	prevCopied         uint64
	copySpeed          float64
	timeRemaining      time.Duration
	lastMouseEventTime time.Time

	quit bool
}

func Run(fs m.FS, renderer w.Renderer, events *stream.Stream[m.Event], roots []m.Root) {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("PANIC: %#v", err)
		}
	}()
	run(fs, renderer, events, roots)
}

func run(fs m.FS, renderer w.Renderer, events *stream.Stream[m.Event], roots []m.Root) {
	c := newController(roots)

	go ticker(events)

	for _, root := range roots {
		scanner := fs.NewArchiveScanner(root)
		c.archives[root] = newArchive(root)
		c.scanners[root] = scanner
		scanner.Send(m.ScanArchive{})
	}
	c.archive = c.archives[roots[0]]

	for !c.quit {
		for _, event := range events.Pull() {
			c.handleEvent(event)
		}

		c.frames++
		screen := w.NewScreen(c.screenSize)
		c.archive.rootWidget().Render(screen, w.Position{X: 0, Y: 0}, c.screenSize)
		renderer.Push(screen)
	}
}

func newController(roots []m.Root) *controller {
	c := &controller{
		roots:    roots,
		archives: map[m.Root]*archive{},
		scanners: map[m.Root]m.ArchiveScanner{},
	}
	return c
}
