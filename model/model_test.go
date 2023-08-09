package model

import (
	"fmt"
	"testing"
)

func TestPath(t *testing.T) {
	path := Path("a/b/c")
	name := path.ParentName()
	fmt.Printf("name.path: %q, name.base: %q\n", name.Path, name.Base)
	name = name.Path.ParentName()
	fmt.Printf("name.path: %q, name.base: %q\n", name.Path, name.Base)
	name = name.Path.ParentName()
	fmt.Printf("name.path: %q, name.base: %q\n", name.Path, name.Base)
	name = name.Path.ParentName()
	fmt.Printf("name.path: %q, name.base: %q\n", name.Path, name.Base)
}
