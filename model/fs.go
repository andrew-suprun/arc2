package model

import (
	"fmt"
)

type FS interface {
	Scan(root string)
	Send(cmd FileCommand)
}

type FileCommand interface {
	cmd()
}

type DeleteFile struct {
	Root string
	Path string
}

func (DeleteFile) cmd() {}

func (d DeleteFile) String() string {
	return fmt.Sprintf("DeleteFile: root: %q, name: %q", d.Root, d.Path)
}

type RenameFile struct {
	Root string
	From string
	To   string
}

func (RenameFile) cmd() {}

func (r RenameFile) String() string {
	return fmt.Sprintf("RenameFile: root: %q, from: %q, to: %q", r.Root, r.From, r.To)
}

type CopyFile struct {
	Root string
	From string
	To   []string
}

func (CopyFile) cmd() {}

func (c CopyFile) String() string {
	return fmt.Sprintf("CopyFile: from: %q, to: %v", c.From, c.To)
}
