package controller

import (
	m "arc/model"
	"arc/stream"
	w "arc/widgets"
	"fmt"
	"runtime/debug"
	"strings"
	"time"
)

type controller struct {
	fs         m.FS
	roots      []m.Root
	currRoot   m.Root
	archives   map[m.Root]*archive
	byHash     map[m.Hash][]*file
	screenSize w.Size
	errors     []m.Error
	quit       bool
}

type archive struct {
	root          m.Root
	idx           int
	currentPath   m.Path
	folders       map[m.Path]*folder
	state         archiveState
	fileTreeLines int
}

type folder struct {
	path               m.Path
	files              map[m.Base]*file
	selectedId         m.Id
	entries            int
	selectedIdx        int
	offsetIdx          int
	sortColumn         m.SortColumn
	sortAscending      []bool
	lastMouseEventTime time.Time
}

type file struct {
	m.Meta
	m.State
	m.Hash
	Counts       []int
	progressSize uint64
	progressDone uint64
}

type archiveState int

const (
	scanning archiveState = iota
	hashing
	ready
	copying
)

func newArchive(root m.Root, idx int) *archive {
	return &archive{
		root:    root,
		idx:     idx,
		folders: map[m.Path]*folder{},
	}
}

func (a *archive) printTo(buf *strings.Builder) {
	fmt.Fprintf(buf, "  Archive: %q\n", a.root)
	fmt.Fprintf(buf, "    Current Path: %q\n", a.currentPath)
	for _, folder := range a.folders {
		folder.printTo(buf)
	}
}

func (a *archive) currFolder() *folder {
	return a.getFolder(a.currentPath)
}

func (a *archive) getFolder(path m.Path) *folder {
	pathFolder, ok := a.folders[path]
	if !ok {
		pathFolder = newFolder(path)
		a.folders[path] = pathFolder
	}
	return pathFolder
}

func newFolder(path m.Path) *folder {
	return &folder{
		path:          path,
		files:         map[m.Base]*file{},
		sortAscending: []bool{true, false, false},
	}
}

func (f *folder) printTo(buf *strings.Builder) {
	fmt.Fprintf(buf, "    Folder: %q\n", f.path)
	fmt.Fprintf(buf, "      Selected Idx: %d\n", f.selectedIdx)
	fmt.Fprintf(buf, "      Offset Idx: %d\n", f.offsetIdx)
	fmt.Fprintln(buf, "      Entries:")
	for _, entry := range f.files {
		fmt.Fprintf(buf, "        %s,\n", entry)
	}
}

func (f *file) String() string {
	return fmt.Sprintf("File{FileId: %q, Size: %d, Hash: %q, State: %s, Hashed: %d}",
		f.Id, f.Size, f.Hash, f.State, f.progressDone)
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
		currRoot: roots[0],
		archives: map[m.Root]*archive{},
		byHash:   map[m.Hash][]*file{},
		fs:       fs,
	}

	for idx, root := range roots {
		c.archives[root] = newArchive(root, idx)
		c.fs.Scan(root)
	}

	for !c.quit {
		events, _ := events.Pull()
		for _, event := range events {
			c.handleEvent(event)
		}

		screen := w.NewScreen(c.screenSize)
		view := c.view()
		rootWidget := view.RootWidget()
		rootWidget.Render(screen, w.Position{X: 0, Y: 0}, c.screenSize)
		renderer.Push(screen)
	}
}

func (c *controller) archive(id m.Id) *archive {
	return c.archives[id.Root]
}

func (c *controller) currArchive() *archive {
	return c.archives[c.currRoot]
}

func (c *controller) folder(id m.Id) *folder {
	return c.archive(id).folders[id.Path]
}

func (c *controller) currFolder() *folder {
	archive := c.currArchive()
	return archive.folders[archive.currentPath]
}

func (c *controller) file(id m.Id) *file {
	return c.folder(id).files[id.Base]
}

func (c *controller) selectedFile() *file {
	folder := c.currFolder()
	return folder.files[folder.selectedId.Base]
}

func (c *controller) selectedId() m.Id {
	return c.currFolder().selectedId
}
