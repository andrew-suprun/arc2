package widgets

import (
	"fmt"
	"strings"
)

type NullWidget struct{}

func (w NullWidget) Constraint() Constraint {
	return Constraint{Size: Size{Width: 0, Height: 0}, Flex: Flex{X: 0, Y: 0}}
}

func (w NullWidget) Render(screen *Screen, pos Position, size Size) {
}

func (s NullWidget) String() string { return toString(s) }

func (s NullWidget) ToString(buf *strings.Builder, offset string) {
	fmt.Fprintf(buf, offset+"NullWidget{}\n")
}
