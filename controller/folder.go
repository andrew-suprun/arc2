package controller

import (
	v "arc/view"
	"fmt"
	"strings"
	"time"
)

func newFolder(root, name string) *folder {
	return &folder{
		meta: meta{
			root: root,
			name: name,
		},
		children:      map[string]*folder{},
		files:         map[string]*file{},
		sortAscending: []bool{true, true, true},
	}
}

func (f *folder) path() []string {
	return f.pathInternal([]string{})
}

func (f *folder) pathInternal(slice []string) []string {
	if f.parent != nil {
		slice = f.parent.pathInternal(slice)
	}
	return append(slice, f.name)
}

func (f *folder) file(name string) *file {
	return f.files[name]
}

func (f *folder) child(name string) *folder {
	child := f.children[name]
	if child == nil {
		child = newFolder(f.root, name)
		f.children[name] = child
		child.parent = f
	}
	return child
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

func (f *folder) updateMetas() {
	f.size = 0
	f.modTime = time.Time{}
	f.state = v.Resolved
	f.progress = v.Progress{}

	for _, sub := range f.children {
		sub.updateMetas()
		f.updateMeta(&sub.meta)
	}
	for _, file := range f.files {
		f.updateMeta(&file.meta)
	}
}

func (f *folder) updateMeta(meta *meta) {
	f.progress.Size = meta.progress.Size
	f.progress.Done = meta.progress.Done
	f.size += meta.size
	if f.modTime.Before(meta.modTime) {
		f.modTime = meta.modTime
	}
	f.state = max(f.state, meta.state)
}

func (folder *folder) uniqueName(name string) string {
	for i := 1; ; i++ {
		name = newSuffix(name, i)
		if _, ok := folder.files[name]; !ok {
			break
		}
	}
	return name
}

func newSuffix(name string, idx int) string {
	parts := strings.Split(name, ".")

	var part string
	if len(parts) == 1 {
		part = stripIdx(parts[0])
	} else {
		part = stripIdx(parts[len(parts)-2])
	}
	var newName string
	if len(parts) == 1 {
		newName = fmt.Sprintf("%s%c%d", part, '`', idx)
	} else {
		parts[len(parts)-2] = fmt.Sprintf("%s%c%d", part, '`', idx)
		newName = strings.Join(parts, ".")
	}
	return newName
}

type stripIdxState int

const (
	expectDigit stripIdxState = iota
	expectDigitOrBacktick
)

func stripIdx(name string) string {
	state := expectDigit
	i := len(name) - 1
	for ; i >= 0; i-- {
		ch := name[i]
		if ch >= '0' && ch <= '9' && (state == expectDigit || state == expectDigitOrBacktick) {
			state = expectDigitOrBacktick
		} else if ch == '`' && state == expectDigitOrBacktick {
			return name[:i]
		} else {
			return name
		}
	}
	return name
}
