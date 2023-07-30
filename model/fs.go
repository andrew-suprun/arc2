package model

import (
	"fmt"
)

type FS interface {
	Scan(root Root)
	Send(cmd FileCommand)
}

type FileCommand interface {
	cmd()
}

type DeleteFile struct {
	Hash Hash // TODO Need it?
	Id   Id
}

func (DeleteFile) cmd() {}

func (d DeleteFile) String() string {
	return fmt.Sprintf("DeleteFile: Id: %q, hash: %q", d.Id, d.Hash)
}

type RenameFile struct {
	Hash Hash // TODO Need it?
	From Id
	To   Id
}

func (RenameFile) cmd() {}

func (r RenameFile) String() string {
	return fmt.Sprintf("RenameFile: From: %q, To: %q, hash: %q", r.From, r.To, r.Hash)
}

type CopyFile struct {
	Hash Hash // TODO Need it?
	From Id
	To   []Id
}

func (CopyFile) cmd() {}

func (c CopyFile) String() string {
	return fmt.Sprintf("CopyFile: From: %q, To: %v, hash: %q", c.From, c.To, c.Hash)
}
