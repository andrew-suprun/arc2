package model

import "fmt"

type Entry interface {
	fmt.Stringer
	Meta() *Meta
	State() State
	SetState(State)
}

type common struct {
	Meta
	State
}

type Folder struct {
	common
	Hashed      uint64
	TotalHashed uint64
}

func NewFolder(meta Meta, state State) *Folder {
	return &Folder{common: common{Meta: meta, State: state}}
}

func (f *Folder) String() string {
	return fmt.Sprintf("Folder{FileId: %q, Size: %d, State: %s, Hashed: %d, TotalHashed: %d}",
		f.Id, f.Size, f.State(), f.Hashed, f.TotalHashed)
}

func (f *Folder) Meta() *Meta {
	return &f.common.Meta
}

func (f *Folder) State() State {
	return f.common.State
}

func (f *Folder) SetState(state State) {
	f.common.State = state
}

type File struct {
	common
	Hash
	Hashed uint64
	Counts []int
}

func NewFile(meta Meta, state State) *File {
	return &File{common: common{Meta: meta, State: state}}
}

func (f *File) String() string {
	return fmt.Sprintf("File{FileId: %q, Size: %d, Hash: %q, State: %s, Hashed: %d}",
		f.Id, f.Size, f.Hash, f.State(), f.Hashed)
}

func (f *File) Meta() *Meta {
	return &f.common.Meta
}

func (f *File) State() State {
	return f.common.State
}
func (f *File) SetState(state State) {
	f.common.State = state
}
