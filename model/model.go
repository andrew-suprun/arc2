package model

import (
	"fmt"
	"path/filepath"
	"time"
)

type File struct {
	Meta
	State
	Hash
	Counts   []int
	Progress uint64
}

func NewFile(meta Meta, state State) *File {
	return &File{Meta: meta, State: state}
}

func (f *File) String() string {
	return fmt.Sprintf("File{FileId: %q, Size: %d, Hash: %q, State: %s, Hashed: %d}",
		f.Id, f.Size, f.Hash, f.State, f.Progress)
}

type Root string

func (root Root) String() string {
	return string(root)
}

type Path string

func (path Path) String() string {
	return string(path)
}

func (p Path) ParentName() Name {
	strPath := p.String()
	path := Path(filepath.Dir(strPath))
	if path == "." {
		path = ""
	}
	return Name{Path: path, Base: Base(filepath.Base(strPath))}
}

type Base string

func (name Base) String() string {
	return string(name)
}

type Name struct {
	Path
	Base
}

func (name Name) String() string {
	return filepath.Join(name.Path.String(), name.Base.String())
}

func (n Name) ChildPath() Path {
	return Path(filepath.Join(n.Path.String(), n.Base.String()))
}

type Id struct {
	Root
	Name
}

func (id Id) String() string {
	return filepath.Join(id.Root.String(), id.Path.String(), id.Base.String())
}

type Hash string

func (hash Hash) String() string {
	return string(hash)
}

type Meta struct {
	Id
	Size    uint64
	ModTime time.Time
}

func (m *Meta) String() string {
	return fmt.Sprintf("Meta{Root: %q, Path: %q Name: %q, Size: %d, ModTime: %s}",
		m.Root, m.Path, m.Base, m.Size, m.ModTime.Format(time.DateTime))
}

type State int

const (
	Resolved State = iota
	Scanned
	Hashing
	Hashed
	Pending
	Copying
	Divergent
)

func (s State) String() string {
	switch s {
	case Resolved:
		return "Resolved"
	case Scanned:
		return "Scanned"
	case Hashing:
		return "Hashing"
	case Hashed:
		return "Hashed"
	case Pending:
		return "Pending"
	case Copying:
		return "Copying"
	case Divergent:
		return "Divergent"
	}
	return "UNKNOWN FILE STATE"
}

func (s State) Merge(other State) State {
	if other > s {
		return other
	}
	return s
}

type SortColumn int

const (
	SortByName SortColumn = iota
	SortByTime
	SortBySize
)

func (c SortColumn) String() string {
	switch c {
	case SortByName:
		return "SortByName"
	case SortByTime:
		return "SortByTime"
	case SortBySize:
		return "SortBySize"
	}
	return "Illegal Sort Solumn"
}
