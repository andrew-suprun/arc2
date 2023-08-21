package controller

import (
	m "arc/model"
	"arc/stream"
	w "arc/widgets"
	"runtime/debug"
)

func Run(fs m.FS, renderer w.Renderer, events *stream.Stream[m.Event], roots []string) (err any, stack []byte) {
	defer func() {
		err = recover()
		stack = debug.Stack()
	}()
	run(fs, renderer, events, roots)
	return nil, nil
}

func run(fs m.FS, renderer w.Renderer, events *stream.Stream[m.Event], roots []string) {
	c := &controller{
		roots:    make([]string, len(roots)),
		archives: map[string]*archive{},
		fs:       fs,
	}

	for i, root := range roots {
		c.roots[i] = root
		c.archives[root] = &archive{
			rootFolder: newFolder(root),
		}
		c.fs.Scan(root)
	}

	c.root = roots[0]

	for !c.quit {
		events, _ := events.Pull()
		for _, event := range events {
			c.handleEvent(event)
		}

		c.frames++
		screen := w.NewScreen(c.screenSize)
		view := c.view()
		rootWidget := view.RootWidget()
		rootWidget.Render(screen, w.Position{X: 0, Y: 0}, c.screenSize)
		renderer.Push(screen)
	}
}

func (c *controller) archive(root string) *archive {
	return c.archives[root]
}

func (c *controller) curArchive() *archive {
	return c.archives[c.root]
}
