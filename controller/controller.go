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
	hashed   int

	screenSize w.Size
	frames     int
	prevTick   time.Time

	copySize        uint64
	totalCopiedSize uint64
	fileCopiedSize  uint64
	prevCopied      uint64
	copySpeed       float64
	timeRemaining   time.Duration

	errors []any

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

func (c *controller) archiveScanned(event m.ArchiveScanned) {
	archive := c.archives[event.Root]
	archive.addFiles(event)
	archive.sort()

	for _, file := range event.Files {
		archive.totalSize += file.Size
	}
}

func (c *controller) fileHashed(event m.FileHashed) {
	archive := c.archives[event.Root]
	folder := archive.getFolder(event.Path)
	file := folder.entry(event.Base)
	file.Hash = event.Hash
	file.State = m.Resolved

	archive.markDuplicates()
	archive.updateFolderStates("")

	archive.parents(file, func(parent *m.File) {
		parent.Hashed = 0
		parent.TotalHashed += file.Size
	})

	archive.totalHashed += file.Size
	archive.fileHashed = 0
}

func (c *controller) archiveHashed(event m.ArchiveHashed) {
	archive := c.archives[event.Root]
	archive.progressInfo = nil
	c.hashed++

	if c.hashed == len(c.roots) {
		c.analyzeDiscrepancies()
	}
}

func (c *controller) handleHashingProgress(event m.HashingProgress) {
	archive := c.archives[event.Root]
	archive.fileHashed = event.Hashed
	folder := archive.folders[event.Path]
	file := folder.entry(event.Base)
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

func (c *controller) analyzeDiscrepancies() {
	allNames := map[m.Name]m.Hash{}
	archiveNames := map[m.Root]map[m.Name]m.Hash{}
	for _, root := range c.roots {
		archiveNames[root] = map[m.Name]m.Hash{}
	}

	for _, archive := range c.archives {
		for _, folder := range archive.folders {
			for _, entry := range folder.entries {
				allNames[entry.Name] = entry.Hash
				archiveNames[entry.Root][entry.Name] = entry.Hash
			}
		}
	}

	divergency := map[m.Name]struct{}{}
	for name, hash := range allNames {
		for _, archNames := range archiveNames {
			if archNames[name] != hash {
				divergency[name] = struct{}{}
			}
		}
	}

	for _, archive := range c.archives {
		for _, folder := range archive.folders {
			for _, entry := range folder.entries {
				if _, ok := divergency[entry.Name]; ok && entry.State != m.Duplicate {
					entry.State = m.Divergent
				}
			}
		}
		archive.updateFolderStates("")
	}
}
