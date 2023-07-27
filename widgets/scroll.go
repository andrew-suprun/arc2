package widgets

import (
	m "arc/model"
	"fmt"
	"strings"
)

type scroll struct {
	command    any
	constraint Constraint
	widget     func(size Size) Widget
}

// TODO: Separate Scroll into Scroll and Sized
func Scroll(event m.Scroll, constraint Constraint, widget func(size Size) Widget) Widget {
	return scroll{command: event.Command, constraint: constraint, widget: widget}
}

func (s scroll) Constraint() Constraint {
	return s.constraint
}

func (s scroll) Render(screen *Screen, pos Position, size Size) {
	screen.ScrollAreas = append(screen.ScrollAreas, ScrollArea{Command: s.command, Position: pos, Size: size})
	widget := s.widget(size)
	widget.Render(screen, pos, size)
}

func (s scroll) String() string { return toString(s) }

func (s scroll) ToString(buf *strings.Builder, offset string) {
	fmt.Fprintf(buf, "%sScroll(%s\n", offset, s.command)
	fmt.Fprintf(buf, "%s| %s\n", offset, s.constraint)
	widget := s.widget(Size{80, 3})
	widget.ToString(buf, offset+"| ")
}
