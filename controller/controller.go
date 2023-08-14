package controller

import (
	m "arc/model"
	"arc/stream"
	w "arc/widgets"
	"runtime/debug"
	"time"
)

type controller struct {
	roots    []m.Root
	archives map[m.Root]*archive
	byHash   map[m.Hash][]*m.File
	archive  *archive
	hashed   int

	*shared

	screenSize w.Size
	frames     int
	prevTick   time.Time

	copySize        uint64
	totalCopiedSize uint64
	fileCopiedSize  uint64
	prevCopied      uint64
	copySpeed       float64
	timeRemaining   time.Duration

	errors []m.Error

	quit bool
}

type shared struct {
	fps int
	fs  m.FS
}

func Run(fs m.FS, renderer w.Renderer, events *stream.Stream[m.Event], roots []m.Root) (err any, stack []byte) {
	defer func() {
		err = recover()
		stack = debug.Stack()
	}()
	run(fs, renderer, events, roots)
	return nil, nil
}

func run(fs m.FS, renderer w.Renderer, events *stream.Stream[m.Event], roots []m.Root) {
	c := &controller{
		roots:    roots,
		archives: map[m.Root]*archive{},
		byHash:   map[m.Hash][]*m.File{},
		shared:   &shared{},
	}
	c.shared.fs = fs

	for idx, root := range roots {
		c.archives[root] = newArchive(root, idx, c.shared)
		c.shared.fs.Scan(root)
	}

	c.archive = c.archives[roots[0]]

	go ticker(events)

	for !c.quit {
		events, _ := events.Pull()
		for _, event := range events {
			c.handleEvent(event)
		}

		c.analyzeDiscrepancies()

		c.frames++
		screen := w.NewScreen(c.screenSize)
		view := c.view()
		rootWidget := view.RootWidget()
		rootWidget.Render(screen, w.Position{X: 0, Y: 0}, c.screenSize)
		renderer.Push(screen)
	}
}
