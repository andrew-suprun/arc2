package controller

import (
	m "arc/model"
	w "arc/widgets"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type archive struct {
	root        m.Root
	folders     map[m.Path]*folder
	currentPath m.Path

	totalSize      uint64
	totalHashed    uint64
	fileHashed     uint64
	prevHashed     uint64
	speed          float64
	timeRemaining  time.Duration
	progressInfo   *progressInfo
	pendingFiles   int
	duplicateFiles int
	absentFiles    int
	fileTreeLines  int
	fps            int
}

type progressInfo struct {
	tab           string
	value         float64
	speed         float64
	timeRemaining time.Duration
}

func newArchive(root m.Root) *archive {
	return &archive{
		root:    root,
		folders: map[m.Path]*folder{},
	}
}

func (a *archive) addFiles(event m.ArchiveScanned) {
	for _, file := range event.Files {
		a.addFile(&m.File{Meta: *file, State: m.Scanned})
	}
	a.currentPath = ""
}

func (a *archive) addFile(file *m.File) {
	folder := a.getFolder(file.Path)
	folder.addEntry(file)
	name := file.ParentName()
	for name.Base != "." {
		parentFolder := a.getFolder(name.Path)
		folderEntry := parentFolder.entry(name.Base)
		if folderEntry != nil {
			folderEntry.Size += file.Size
			if folderEntry.ModTime.Before(file.ModTime) {
				folderEntry.ModTime = file.ModTime
			}
		} else {
			folderEntry := &m.File{
				Meta: m.Meta{
					Id: m.Id{
						Root: file.Root,
						Name: name,
					},
					Size:    file.Size,
					ModTime: file.ModTime,
				},
				Kind:  m.FileFolder,
				State: m.Scanned,
			}
			parentFolder.addEntry(folderEntry)

		}

		name = name.Path.ParentName()
	}
}

func (a *archive) sort() {
	for _, folder := range a.folders {
		folder.sort()
	}
}

func (a *archive) currentFolder() *folder {
	return a.getFolder(a.currentPath)
}

func (a *archive) getFolder(path m.Path) *folder {
	pathFolder, ok := a.folders[path]
	if !ok {
		pathFolder = &folder{
			sortAscending: []bool{true, false, false},
		}
		a.folders[path] = pathFolder
	}
	return pathFolder
}

func (a *archive) parents(file *m.File, proc func(parent *m.File)) {
	name := file.ParentName()
	for name.Base != "." {
		proc(a.getFolder(name.Path).entry(name.Base))
		name = name.Path.ParentName()
	}
}

func (a *archive) markDuplicates() {
	hashes := map[m.Hash]int{}
	for _, folder := range a.folders {
		for _, entry := range folder.entries {
			if entry.Hash != "" {
				hashes[entry.Hash]++
			}
		}
	}

	for _, folder := range a.folders {
		for _, entry := range folder.entries {
			if hashes[entry.Hash] > 1 {
				entry.State = m.Duplicate
			}
		}
	}

	a.duplicateFiles = 0
	for _, count := range hashes {
		if count > 1 {
			a.duplicateFiles++
		}
	}
}

func (a *archive) updateFolderStates(path m.Path) m.State {
	state := m.Resolved
	folder := a.getFolder(path)
	for _, entry := range folder.entries {
		if entry.Kind == m.FileFolder {
			entry.State = a.updateFolderStates(m.Path(entry.Name.String()))
			state = state.Merge(entry.State)
		} else {
			state = state.Merge(entry.State)
		}
	}
	return state
}

// Events

func (a *archive) enter() {
	file := a.currentFolder().selected()
	if file != nil && file.Kind == m.FileFolder {
		a.currentPath = m.Path(file.Name.String())
	}
}

func (a *archive) exit() {
	if a.currentPath == "" {
		return
	}
	parts := strings.Split(a.currentPath.String(), "/")
	if len(parts) == 1 {
		a.currentPath = ""
	}
	a.currentPath = m.Path(filepath.Join(parts[:len(parts)-1]...))
}

func (a *archive) mouseTarget(cmd any) {
	switch cmd := cmd.(type) {
	case m.SelectFile:
		a.currentFolder().selectFile(cmd)

	case m.SelectFolder:
		a.currentPath = m.Path(cmd)

	case m.SortColumn:
		folder := a.currentFolder()
		folder.selectSortColumn(cmd)
		folder.sort()
		folder.makeSelectedVisible(a.fileTreeLines)
	}
}

// Widgets

func (a *archive) rootWidget(fileTreeLines int) w.Widget {
	return w.Styled(styleDefault,
		w.Column(colConstraint,
			a.title(),
			a.folderWidget(fileTreeLines),
			a.progressWidget(),
			a.fileStats(),
		),
	)
}

func (a *archive) title() w.Widget {
	return w.Row(rowConstraint,
		w.Styled(styleAppTitle, w.Text(" Archive")), w.Text(" "),
		w.Styled(styleArchive, w.Text(a.root.String()).Flex(1)),
	)
}

func (a *archive) folderWidget(fileTreeLines int) w.Widget {
	return w.Column(colConstraint,
		a.breadcrumbs(),
		w.Styled(styleArchiveHeader,
			w.Row(rowConstraint,
				w.Text(" Status").Width(13),
				w.MouseTarget(m.SortByName, w.Text(" Document"+a.sortIndicator(m.SortByName)).Width(20).Flex(1)),
				w.MouseTarget(m.SortByTime, w.Text("  Date Modified"+a.sortIndicator(m.SortByTime)).Width(19)),
				w.MouseTarget(m.SortBySize, w.Text(fmt.Sprintf("%22s", "Size"+a.sortIndicator(m.SortBySize)+" "))),
			),
		),
		w.Scroll(m.Scroll{}, w.Constraint{Size: w.Size{Width: 0, Height: 0}, Flex: w.Flex{X: 1, Y: 1}},
			func(size w.Size) w.Widget {
				folder := a.currentFolder()
				a.fileTreeLines = size.Height
				if folder.offsetIdx > len(folder.entries)+1-size.Height {
					folder.offsetIdx = len(folder.entries) + 1 - size.Height
				}
				if folder.offsetIdx < 0 {
					folder.offsetIdx = 0
				}
				rows := []w.Widget{}
				selected := folder.selected()
				for i, file := range folder.entries[folder.offsetIdx:] {
					if i >= size.Height {
						break
					}
					rows = append(rows, w.Styled(a.styleFile(file, selected == file),
						w.MouseTarget(m.SelectFile(file.Id), w.Row(rowConstraint,
							a.fileRow(file)...,
						)),
					))
				}
				rows = append(rows, w.Spacer{})
				return w.Column(colConstraint, rows...)
			},
		),
	)
}

func (a *archive) fileRow(file *m.File) []w.Widget {
	result := []w.Widget{w.Text(" "), state(file)}

	if file.Kind == m.FileRegular {
		result = append(result, w.Text("   "))
	} else {
		result = append(result, w.Text(" ▶ "))
	}
	result = append(result, w.Text(file.Base.String()).Width(20).Flex(1))
	result = append(result, w.Text("  "))
	result = append(result, w.Text(file.ModTime.Format(time.DateTime)))
	result = append(result, w.Text("  "))
	result = append(result, w.Text(formatSize(file.Size)).Width(18))
	return result
}

func state(file *m.File) w.Widget {
	totalHashed := file.TotalHashed + file.Hashed
	if totalHashed > 0 && file.TotalHashed+file.Hashed < file.Size {
		value := float64(file.TotalHashed+file.Hashed) / float64(file.Size)
		return w.Styled(styleProgressBar, w.ProgressBar(value).Width(10).Flex(0))
	}
	text := ""
	switch file.State {
	case m.Pending:
		text = "Pending"
	case m.Duplicate:
		text = "Duplicate"
	case m.Absent:
		text = "Absent"
	}
	return w.Text(text).Width(10)
}

func (a *archive) sortIndicator(column m.SortColumn) string {
	folder := a.currentFolder()
	if column == folder.sortColumn {
		if folder.sortAscending[column] {
			return " ▲"
		}
		return " ▼"
	}
	return ""
}

func (a *archive) breadcrumbs() w.Widget {
	names := strings.Split(a.currentPath.String(), "/")
	widgets := make([]w.Widget, 0, len(names)*2+2)
	widgets = append(widgets, w.MouseTarget(m.SelectFolder(""),
		w.Styled(styleBreadcrumbs, w.Text(" Root")),
	))
	for i := range names {
		widgets = append(widgets, w.Text(" / "))
		widgets = append(widgets,
			w.MouseTarget(m.SelectFolder(m.Path(filepath.Join(names[:i+1]...))),
				w.Styled(styleBreadcrumbs, w.Text(names[i])),
			),
		)
	}
	widgets = append(widgets, w.Spacer{})
	return w.Row(rowConstraint, widgets...)
}

func (a *archive) progressWidget() w.Widget {
	if a.progressInfo == nil {
		return w.NullWidget{}
	}
	return w.Styled(styleStatusLine, w.Row(w.Constraint{Size: w.Size{Width: 0, Height: 1}, Flex: w.Flex{X: 1, Y: 0}},
		w.Text(a.progressInfo.tab), w.Text(" "),
		w.Text(fmt.Sprintf(" %6.2f%%", a.progressInfo.value*100)),
		w.Text(fmt.Sprintf(" %5.1f Mb/S", a.progressInfo.speed)),
		w.Text(fmt.Sprintf(" ETA %6s", a.progressInfo.timeRemaining.Truncate(time.Second))), w.Text(" "),
		w.Styled(styleProgressBar, w.ProgressBar(a.progressInfo.value)),
		w.Text(" "),
	))
}

func (a *archive) fileStats() w.Widget {
	stats := []w.Widget{}
	if a.duplicateFiles == 0 && a.absentFiles == 0 && a.pendingFiles == 0 {
		stats = append(stats, w.Text(" All Clear").Flex(1))
	} else {
		stats = append(stats, w.Text(" Stats:"))
		if a.duplicateFiles > 0 {
			stats = append(stats, w.Text(fmt.Sprintf(" Duplicates: %d", a.duplicateFiles)))
		}
		if a.absentFiles > 0 {
			stats = append(stats, w.Text(fmt.Sprintf(" Absent: %d", a.absentFiles)))
		}
		if a.pendingFiles > 0 {
			stats = append(stats, w.Text(fmt.Sprintf(" Pending: %d", a.pendingFiles)))
		}
		stats = append(stats, w.Text("").Flex(1))
	}
	stats = append(stats, w.Text(fmt.Sprintf(" FPS: %d ", a.fps)))
	return w.Styled(
		styleAppTitle,
		w.Row(w.Constraint{Size: w.Size{Width: 0, Height: 1}, Flex: w.Flex{X: 1, Y: 0}}, stats...),
	)

}

func formatSize(size uint64) string {
	str := fmt.Sprintf("%13d ", size)
	slice := []string{str[:1], str[1:4], str[4:7], str[7:10]}
	b := strings.Builder{}
	for _, s := range slice {
		b.WriteString(s)
		if s == " " || s == "   " {
			b.WriteString(" ")
		} else {
			b.WriteString(",")
		}
	}
	b.WriteString(str[10:])
	return b.String()
}

func (a *archive) styleFile(file *m.File, selected bool) w.Style {
	bg, flags := byte(17), w.Flags(0)
	if file.Kind == m.FileFolder {
		bg = byte(18)
	}
	result := w.Style{FG: a.statusColor(file), BG: bg, Flags: flags}
	if selected {
		result.Flags |= w.Reverse
	}
	return result
}

var styleBreadcrumbs = w.Style{FG: 250, BG: 17, Flags: w.Bold + w.Italic}

func (a *archive) statusColor(file *m.File) byte {
	switch file.State {
	case m.Scanned:
		return 248
	case m.Hashing:
		return 248
	case m.Resolved:
		return 195
	case m.Pending:
		return 214
	case m.Duplicate, m.Absent:
		return 196
	}
	return 231
}
