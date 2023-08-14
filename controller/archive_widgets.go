package controller

// import (
// 	m "arc/model"
// 	w "arc/widgets"
// 	"fmt"
// 	"path/filepath"
// 	"strings"
// 	"time"
// )

// func (a *archive) rootWidget() w.Widget {
// 	return w.Styled(styleDefault,
// 		w.Column(colConstraint,
// 			a.title(),
// 			a.folderWidget(),
// 			a.progressWidget(),
// 			a.fileStats(),
// 		),
// 	)
// }

// func (a *archive) title() w.Widget {
// 	return w.Row(rowConstraint,
// 		w.Styled(styleAppTitle, w.Text(" Archive")), w.Text(" "),
// 		w.Styled(styleArchive, w.Text(a.root.String()).Flex(1)),
// 	)
// }

// func (a *archive) folderWidget() w.Widget {
// 	return w.Column(colConstraint,
// 		a.breadcrumbs(),
// 		w.Styled(styleArchiveHeader,
// 			w.Row(rowConstraint,
// 				w.Text(" Status").Width(13),
// 				w.MouseTarget(m.SortByName, w.Text(" Document"+a.sortIndicator(m.SortByName)).Width(20).Flex(1)),
// 				w.MouseTarget(m.SortByTime, w.Text("  Date Modified"+a.sortIndicator(m.SortByTime)).Width(19)),
// 				w.MouseTarget(m.SortBySize, w.Text(fmt.Sprintf("%22s", "Size"+a.sortIndicator(m.SortBySize)+" "))),
// 			),
// 		),
// 		w.Scroll(m.Scroll{}, w.Constraint{Size: w.Size{Width: 0, Height: 0}, Flex: w.Flex{X: 1, Y: 1}},
// 			func(size w.Size) w.Widget {
// 				folder := a.currentFolder()
// 				a.fileTreeLines = size.Height
// 				if folder.offsetIdx > len(folder.entries)+1-size.Height {
// 					folder.offsetIdx = len(folder.entries) + 1 - size.Height
// 				}
// 				if folder.offsetIdx < 0 {
// 					folder.offsetIdx = 0
// 				}
// 				rows := []w.Widget{}
// 				selected := folder.selected()
// 				for i, file := range folder.entries[folder.offsetIdx:] {
// 					if i >= size.Height {
// 						break
// 					}
// 					rows = append(rows, w.Styled(a.styleFile(file, selected == file),
// 						w.MouseTarget(m.SelectFile(file.Meta().Id), w.Row(rowConstraint,
// 							a.fileRow(file)...,
// 						)),
// 					))
// 				}
// 				rows = append(rows, w.Spacer{})
// 				return w.Column(colConstraint, rows...)
// 			},
// 		),
// 	)
// }

// func (a *archive) fileRow(file m.Entry) []w.Widget {
// 	result := []w.Widget{w.Text(" "), state(file)}

// 	switch file.(type) {
// 	case *m.File:
// 		result = append(result, w.Text("   "))
// 	case *m.Folder:
// 		result = append(result, w.Text(" ▶ "))
// 	}

// 	result = append(result, w.Text(file.Meta().Id.Base.String()).Width(20).Flex(1))
// 	result = append(result, w.Text("  "))
// 	result = append(result, w.Text(file.Meta().ModTime.Format(time.DateTime)))
// 	result = append(result, w.Text("  "))
// 	result = append(result, w.Text(formatSize(file.Meta().Size)).Width(18))
// 	return result
// }

// func state(entry m.Entry) w.Widget {
// 	hashed := entry.Meta().Size
// 	if folder, ok := entry.(*m.Folder); ok {
// 		hashed += folder.TotalHashed
// 	}
// 	if hashed > 0 && hashed < entry.Meta().Size {
// 		value := float64(hashed) / float64(entry.Meta().Size)
// 		return w.Styled(styleProgressBar, w.ProgressBar(value).Width(10).Flex(0))
// 	}
// 	switch entry.State() {
// 	case m.Pending:
// 		return w.Text("Pending").Width(10)
// 	case m.Divergent:
// 		break
// 	default:
// 		return w.Text("").Width(10)
// 	}

// 	buf := &strings.Builder{}
// 	if file, ok := entry.(*m.File); ok {
// 		for _, count := range file.Counts {
// 			fmt.Fprintf(buf, "%c", countRune(count))
// 		}
// 	}
// 	return w.Text(buf.String()).Width(10)
// }

// func countRune(count int) rune {
// 	if count == 0 {
// 		return '-'
// 	}
// 	if count > 9 {
// 		return '*'
// 	}
// 	return '0' + rune(count)
// }

// func (a *archive) sortIndicator(column m.SortColumn) string {
// 	folder := a.currentFolder()
// 	if column == folder.sortColumn {
// 		if folder.sortAscending[column] {
// 			return " ▲"
// 		}
// 		return " ▼"
// 	}
// 	return ""
// }

// func (a *archive) breadcrumbs() w.Widget {
// 	names := strings.Split(a.currentPath.String(), "/")
// 	widgets := make([]w.Widget, 0, len(names)*2+2)
// 	widgets = append(widgets, w.MouseTarget(m.SelectFolder(""),
// 		w.Styled(styleBreadcrumbs, w.Text(" Root")),
// 	))
// 	for i := range names {
// 		widgets = append(widgets, w.Text(" / "))
// 		widgets = append(widgets,
// 			w.MouseTarget(m.SelectFolder(m.Path(filepath.Join(names[:i+1]...))),
// 				w.Styled(styleBreadcrumbs, w.Text(names[i])),
// 			),
// 		)
// 	}
// 	widgets = append(widgets, w.Spacer{})
// 	return w.Row(rowConstraint, widgets...)
// }

// func (a *archive) progressWidget() w.Widget {
// 	if a.progressInfo == nil {
// 		return w.NullWidget{}
// 	}
// 	return w.Styled(styleStatusLine, w.Row(w.Constraint{Size: w.Size{Width: 0, Height: 1}, Flex: w.Flex{X: 1, Y: 0}},
// 		w.Text(a.progressInfo.tab), w.Text(" "),
// 		w.Text(fmt.Sprintf(" %6.2f%%", a.progressInfo.value*100)),
// 		w.Text(fmt.Sprintf(" %5.1f Mb/S", a.progressInfo.speed)),
// 		w.Text(fmt.Sprintf(" ETA %6s", a.progressInfo.timeRemaining.Truncate(time.Second))), w.Text(" "),
// 		w.Styled(styleProgressBar, w.ProgressBar(a.progressInfo.value)),
// 		w.Text(" "),
// 	))
// }

// func (a *archive) fileStats() w.Widget {
// 	stats := []w.Widget{}
// 	if a.duplicateFiles == 0 && a.absentFiles == 0 && a.pendingFiles == 0 {
// 		stats = append(stats, w.Text(" All Clear").Flex(1))
// 	} else {
// 		stats = append(stats, w.Text(" Stats:"))
// 		if a.duplicateFiles > 0 {
// 			stats = append(stats, w.Text(fmt.Sprintf(" Duplicates: %d", a.duplicateFiles)))
// 		}
// 		if a.absentFiles > 0 {
// 			stats = append(stats, w.Text(fmt.Sprintf(" Absent: %d", a.absentFiles)))
// 		}
// 		if a.pendingFiles > 0 {
// 			stats = append(stats, w.Text(fmt.Sprintf(" Pending: %d", a.pendingFiles)))
// 		}
// 		stats = append(stats, w.Text("").Flex(1))
// 	}
// 	stats = append(stats, w.Text(fmt.Sprintf(" FPS: %d ", a.fps)))
// 	return w.Styled(
// 		styleAppTitle,
// 		w.Row(w.Constraint{Size: w.Size{Width: 0, Height: 1}, Flex: w.Flex{X: 1, Y: 0}}, stats...),
// 	)

// }

// func formatSize(size uint64) string {
// 	str := fmt.Sprintf("%13d ", size)
// 	slice := []string{str[:1], str[1:4], str[4:7], str[7:10]}
// 	b := strings.Builder{}
// 	for _, s := range slice {
// 		b.WriteString(s)
// 		if s == " " || s == "   " {
// 			b.WriteString(" ")
// 		} else {
// 			b.WriteString(",")
// 		}
// 	}
// 	b.WriteString(str[10:])
// 	return b.String()
// }

// func (a *archive) styleFile(file m.Entry, selected bool) w.Style {
// 	bg, flags := byte(17), w.Flags(0)
// 	if _, ok := file.(*m.Folder); ok {
// 		bg = byte(18)
// 	}
// 	result := w.Style{FG: a.statusColor(file), BG: bg, Flags: flags}
// 	if selected {
// 		result.Flags |= w.Reverse
// 	}
// 	return result
// }

// var styleBreadcrumbs = w.Style{FG: 250, BG: 17, Flags: w.Bold + w.Italic}

// func (a *archive) statusColor(file m.Entry) byte {
// 	switch file.State() {
// 	case m.Scanned:
// 		return 248
// 	case m.Hashing:
// 		return 248
// 	case m.Resolved:
// 		return 195
// 	case m.Pending:
// 		return 214
// 	case m.Divergent:
// 		return 196
// 	}
// 	return 231
// }
