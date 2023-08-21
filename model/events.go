package model

import (
	"fmt"
	"time"
)

type FileScanned struct {
	Root    string
	Path    string
	Size    uint64
	ModTime time.Time
}

type ArchiveScanned string

type ArchiveHashed string

type FileHashed struct {
	Root string
	Path string
	Hash string
}

func (f FileHashed) String() string {
	return fmt.Sprintf("FileHashed: path: %q, hash: %q", f.Path, f.Hash)
}

type FileDeleted DeleteFile

func (h FileDeleted) String() string {
	return DeleteFile(h).String()
}

type FileRenamed RenameFile

func (h FileRenamed) String() string {
	return RenameFile(h).String()
}

type FileCopied CopyFile

func (h FileCopied) String() string {
	return CopyFile(h).String()
}

type ProgressState int // TODO move elsewhere

const (
	ProgressInitial ProgressState = iota
	ProgressScanned
	ProgressHashed
)

type HashingProgress struct {
	Root   string
	Path   string
	Hashed uint64
}

type CopyingProgress struct {
	Root   string
	Path   string
	Copied uint64
}

type Error struct {
	Path  string
	Error error
}

type ScreenSize struct {
	Width, Height int
}

type SelectArchive struct {
	Idx int
}

type Open struct{}

type Enter struct{}

type Exit struct{}

type RevealInFinder struct{}

type SelectFirst struct{}

type SelectLast struct{}

type MoveSelection struct{ Lines int }

type ResolveOne struct{}

type ResolveAll struct{}

type KeepAll struct{}

type Tab struct{}

type Delete struct{}

type Scroll struct {
	Command any
	Lines   int
}

func (s Scroll) String() string {
	return fmt.Sprintf("Scroll(%#v)", s.Lines)
}

type MouseTarget struct{ Command any }

func (t MouseTarget) String() string {
	return fmt.Sprintf("MouseTarget(%q)", t.Command)
}

type PgUp struct{}

type PgDn struct{}

type DebugPrintState struct{}

type DebugPrintRootWidget struct{}

type Quit struct{}
