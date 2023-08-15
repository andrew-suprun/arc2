package view

import (
	m "arc/model"
	w "arc/widgets"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"
)

type View struct {
	Archive       m.Root
	Path          m.Path
	Entries       []Entry
	SelectedBase  m.Base
	FileTreeLines int
	OffsetIdx     int
	Progress      *Progress
	SortColumn    m.SortColumn
	SortAscending bool
}

type Entry struct {
	*m.File
	Kind
}

type Kind int

const (
	Regular Kind = iota
	Folder
)

type Progress struct {
	Tab           string
	Value         float64
	TimeRemaining time.Duration // TODO Implement
}

var (
	styleDefault       = w.Style{FG: 226, BG: 17}
	styleAppTitle      = w.Style{FG: 226, BG: 0, Flags: w.Bold + w.Italic}
	styleStatusLine    = w.Style{FG: 230, BG: 0, Flags: w.Italic}
	styleArchive       = w.Style{FG: 226, BG: 0, Flags: w.Bold}
	styleProgressBar   = w.Style{FG: 231, BG: 19}
	styleArchiveHeader = w.Style{FG: 231, BG: 8, Flags: w.Bold}
)

var (
	rowConstraint = w.Constraint{Size: w.Size{Width: 0, Height: 1}, Flex: w.Flex{X: 1, Y: 0}}
	colConstraint = w.Constraint{Size: w.Size{Width: 0, Height: 0}, Flex: w.Flex{X: 1, Y: 1}}
)

func (a *View) RootWidget() w.Widget {
	return w.Styled(styleDefault,
		w.Column(colConstraint,
			a.title(),
			a.folderWidget(),
			a.progressWidget(),
		),
	)
}

func (a *View) title() w.Widget {
	return w.Row(rowConstraint,
		w.Styled(styleAppTitle, w.Text(" Archive")), w.Text(" "),
		w.Styled(styleArchive, w.Text(a.Archive.String()).Flex(1)),
	)
}

func (v *View) folderWidget() w.Widget {
	return w.Column(colConstraint,
		v.breadcrumbs(),
		w.Styled(styleArchiveHeader,
			w.Row(rowConstraint,
				w.Text(" Status").Width(13),
				w.MouseTarget(m.SortByName, w.Text(" Document"+v.sortIndicator(m.SortByName)).Width(20).Flex(1)),
				w.MouseTarget(m.SortByTime, w.Text("  Date Modified"+v.sortIndicator(m.SortByTime)).Width(19)),
				w.MouseTarget(m.SortBySize, w.Text(fmt.Sprintf("%22s", "Size"+v.sortIndicator(m.SortBySize)+" "))),
			),
		),
		w.Scroll(m.Scroll{}, w.Constraint{Size: w.Size{Width: 0, Height: 0}, Flex: w.Flex{X: 1, Y: 1}},
			func(size w.Size) w.Widget {
				v.FileTreeLines = size.Height
				if v.OffsetIdx > len(v.Entries)+1-size.Height {
					v.OffsetIdx = len(v.Entries) + 1 - size.Height
				}
				if v.OffsetIdx < 0 {
					v.OffsetIdx = 0
				}
				rows := []w.Widget{}
				for i, file := range v.Entries[v.OffsetIdx:] {
					if i >= size.Height {
						break
					}
					rows = append(rows, w.Styled(v.styleFile(file, v.SelectedBase == file.Base),
						w.MouseTarget(m.SelectFile(file.Id), w.Row(rowConstraint,
							v.fileRow(file)...,
						)),
					))
				}
				rows = append(rows, w.Spacer{})
				return w.Column(colConstraint, rows...)
			},
		),
	)
}

func (a *View) fileRow(file Entry) []w.Widget {
	result := []w.Widget{w.Text(" "), state(file)}

	switch file.Kind {
	case Regular:
		result = append(result, w.Text("   "))
	case Folder:
		result = append(result, w.Text(" ▶ "))
	}

	result = append(result, w.Text(file.Id.Base.String()).Width(20).Flex(1))
	result = append(result, w.Text("  "))
	result = append(result, w.Text(file.ModTime.Format(time.DateTime)))
	result = append(result, w.Text("  "))
	result = append(result, w.Text(formatSize(file.Size)).Width(18))
	return result
}

func state(entry Entry) w.Widget {
	switch entry.State {
	case m.Hashing, m.Copying:
		value := float64(entry.Progress) / float64(entry.Size)
		return w.Styled(styleProgressBar, w.ProgressBar(value).Width(10).Flex(0))
	case m.Pending:
		return w.Text("Pending").Width(10)
	case m.Divergent:
		break
	default:
		return w.Text("").Width(10)
	}

	buf := &strings.Builder{}
	for _, count := range entry.Counts {
		fmt.Fprintf(buf, "%c", countRune(count))
	}
	return w.Text(buf.String()).Width(10)
}

func countRune(count int) rune {
	if count == 0 {
		return '-'
	}
	if count > 9 {
		return '*'
	}
	return '0' + rune(count)
}

func (v *View) sortIndicator(column m.SortColumn) string {
	if column == v.SortColumn {
		if v.SortAscending {
			return " ▲"
		}
		return " ▼"
	}
	return ""
}

func (v *View) breadcrumbs() w.Widget {
	names := strings.Split(v.Path.String(), "/")
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

func (a *View) progressWidget() w.Widget {
	if a.Progress == nil {
		return w.NullWidget{}
	}
	return w.Styled(styleStatusLine, w.Row(w.Constraint{Size: w.Size{Width: 0, Height: 1}, Flex: w.Flex{X: 1, Y: 0}},
		w.Text(a.Progress.Tab), w.Text(" "),
		w.Text(fmt.Sprintf(" %6.2f%%", a.Progress.Value*100)),
		w.Text(fmt.Sprintf(" ETA %6s", a.Progress.TimeRemaining.Truncate(time.Second))), w.Text(" "),
		w.Styled(styleProgressBar, w.ProgressBar(a.Progress.Value)),
		w.Text(" "),
	))
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

func (a *View) styleFile(file Entry, selected bool) w.Style {
	bg, flags := byte(17), w.Flags(0)
	if file.Kind == Folder {
		bg = byte(18)
	}
	result := w.Style{FG: a.statusColor(file), BG: bg, Flags: flags}
	if selected {
		result.Flags |= w.Reverse
	}
	return result
}

var styleBreadcrumbs = w.Style{FG: 250, BG: 17, Flags: w.Bold + w.Italic}

func (a *View) statusColor(file Entry) (color byte) {
	log.Printf("statusColor: file: %q, color: %s", file.Id, file.State)
	switch file.State {
	case m.Resolved, m.Hashed:
		return 195
	case m.Scanned, m.Hashing:
		return 248
	case m.Pending, m.Copying:
		return 214
	case m.Divergent:
		return 196
	}
	return 231
}
