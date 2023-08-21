package controller

import (
	v "arc/view"
	"fmt"
	"strings"
)

func newFolder(name string) *folder {
	return &folder{
		name:     name,
		children: map[string]*folder{},
		files:    map[string]*file{},
	}
}

func (f *folder) file(name string) *file {
	return f.files[name]
}

func (f *folder) child(name string) *folder {
	child := f.children[name]
	if child == nil {
		child = newFolder(name)
		f.children[name] = child
	}
	return child
}

func (f *folder) mergeState(state v.State) {
	if f != nil && f.state < state {
		f.state = state
		f.parent.mergeState(state)
	}
}

func (f *folder) updateState() {
	if f == nil {
		return
	}
	curState := f.state
	newState := v.Resolved
	for _, folder := range f.children {
		if newState < folder.state {
			newState = folder.state
		}
	}
	for _, file := range f.files {
		if newState < file.state {
			newState = file.state
		}
	}
	if f != nil && f.state < newState {
		f.state = newState
		f.parent.mergeState(newState)
	}
	if curState != newState {
		f.state = newState
		f.parent.updateState()
	}
}

func (f *folder) printTo(buf *strings.Builder) {
	fmt.Fprintf(buf, "    Folder:\n")
	fmt.Fprintf(buf, "      Selected Idx: %d\n", f.selectedIdx)
	fmt.Fprintf(buf, "      Offset Idx: %d\n", f.offsetIdx)
	fmt.Fprintln(buf, "      Folders:")
	for _, child := range f.children {
		fmt.Fprintf(buf, "        %s,\n", child)
	}
	fmt.Fprintln(buf, "      Files:")
	for _, file := range f.files {
		fmt.Fprintf(buf, "        %s,\n", file)
	}
}
