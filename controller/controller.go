package controller

import (
	m "arc/model"
	"arc/stream"
	w "arc/widgets"
	"log"
	"runtime/debug"
	"sort"
	"strings"
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

	copySize        uint64
	totalCopiedSize uint64
	fileCopiedSize  uint64
	prevCopied      uint64
	copySpeed       float64
	timeRemaining   time.Duration

	quit bool
}

func Run(fs m.FS, renderer w.Renderer, events *stream.Stream[m.Event], roots []m.Root) {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("PANIC: %#v", err)
			debug.PrintStack()
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
		c.archive.rootWidget(c.archive.fileTreeLines).Render(screen, w.Position{X: 0, Y: 0}, c.screenSize)
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

// Events

func (c *controller) tab() {
	selected := c.archive.currentFolder().selected()

	if selected == nil || selected.Kind != m.FileRegular {
		return
	}
	sameHash := []*m.File{}
	for _, archive := range c.archives {
		for _, folder := range archive.folders {
			for _, entry := range folder.entries {
				if entry.Hash == selected.Hash {
					sameHash = append(sameHash, entry)
				}
			}
		}
	}

	sort.Slice(sameHash, func(i, j int) bool {
		return strings.ToLower(sameHash[i].Name.String()) < strings.ToLower(sameHash[j].Name.String())
	})

	idx := 0
	for idx := range sameHash {
		if sameHash[idx] == selected {
			break
		}
	}
	newSelected := sameHash[(idx+1)%len(sameHash)]
	c.archive = c.archives[newSelected.Root]
	c.archive.currentPath = newSelected.Path
	for idx, entry := range c.archive.currentFolder().entries {
		if newSelected == entry {
			c.archive.currentFolder().selectedIdx = idx
			break
		}
	}
	c.archive.currentFolder().makeSelectedVisible(c.archive.fileTreeLines)
}
