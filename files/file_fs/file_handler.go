package file_fs

import (
	m "arc/model"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (f *fileFs) deleteFile(delete m.DeleteFile) {
	log.Printf("### delete %q", delete.Id)
	defer func() {
		f.events.Push(m.FileDeleted(delete))
	}()
	err := os.Remove(delete.Id.String())
	if err != nil {
		f.events.Push(m.Error{Id: delete.Id, Error: err})
	}
	path := filepath.Join(delete.Id.Root.String(), delete.Id.Path.String())
	fsys := os.DirFS(path)

	entries, _ := fs.ReadDir(fsys, ".")
	hasFiles := false
	for _, entry := range entries {
		if entry.Name() != ".DS_Store" && !strings.HasPrefix(entry.Name(), "._") {
			hasFiles = true
			break
		}
	}
	if !hasFiles {
		os.RemoveAll(path)
	}
}

func (f *fileFs) renameFile(rename m.RenameFile) {
	log.Printf("### rename %q to %q", rename.From, rename.To)
	defer func() {
		f.events.Push(m.FileRenamed(rename))
	}()
	path := filepath.Join(rename.From.Root.String(), rename.To.Path.String())
	err := os.MkdirAll(path, 0755)
	if err != nil {
		f.events.Push(m.Error{Id: rename.From, Error: err})
	}
	err = os.Rename(rename.From.String(), rename.To.String())
	if err != nil {
		f.events.Push(m.Error{Id: rename.From, Error: err})
	}
}

func (f *fileFs) copyFile(copy m.CopyFile) {
	log.Printf("### copy from %q", copy.From)
	for _, to := range copy.To {
		log.Printf("### copy   to %q", to)

	}
	defer func() {
		f.events.Push(m.FileCopied(copy))
	}()

	events := make([]chan event, len(copy.To))
	copied := make([]uint64, len(copy.To))
	reported := uint64(0)

	for i := range copy.To {
		events[i] = make(chan event, 1)
	}

	go f.reader(copy.From, copy.To, events)

	for {
		hasValue := false
		minCopied := uint64(0)
		for i := range events {
			if event, ok := <-events[i]; ok {
				hasValue = true
				switch event := event.(type) {
				case copyProgress:
					copied[i] = uint64(event)
					minCopied = copied[i]

				case copyError:
					f.events.Push(m.Error{Id: event.Id, Error: event.Error})
				}
			}
		}
		for _, fileCopied := range copied {
			if minCopied > fileCopied {
				minCopied = fileCopied
			}
		}
		if reported < minCopied {
			reported = minCopied
			f.events.Push(m.CopyingProgress(reported))
		}
		if !hasValue {
			break
		}
	}
}

type event interface {
	event()
}

type copyProgress uint64

func (copyProgress) event() {}

type copyError m.Error

func (copyError) event() {}

func (f *fileFs) reader(source m.Id, targets []m.Id, eventChans []chan event) {
	commands := make([]chan []byte, len(targets))
	defer func() {
		for _, cmdChan := range commands {
			close(cmdChan)
		}
	}()

	info, err := os.Stat(source.String())
	if err != nil {
		f.events.Push(m.Error{Id: source, Error: err})
		return
	}

	for i := range targets {
		commands[i] = make(chan []byte)
		go f.writer(m.Id{Root: targets[i].Root, Path: source.Path}, info.ModTime(), commands[i], eventChans[i])
	}

	sourceFile, err := os.Open(source.String())
	if err != nil {
		f.events.Push(m.Error{Id: source, Error: err})
		return
	}

	var n int
	for err != io.EOF && !f.lc.ShoudStop() {
		buf := make([]byte, 1024*1024)
		n, err = sourceFile.Read(buf)
		if err != nil && err != io.EOF {
			f.events.Push(m.Error{Id: source, Error: err})
			return
		}
		for _, cmd := range commands {
			cmd <- buf[:n]
		}
	}
}

func (f *fileFs) writer(id m.Id, modTime time.Time, cmdChan chan []byte, eventChan chan event) {
	var copied copyProgress

	filePath := filepath.Join(id.Root.String(), id.Path.String())
	os.MkdirAll(filePath, 0755)
	file, err := os.Create(id.String())
	if err != nil {
		f.events.Push(m.Error{Id: id, Error: err})
		return
	}

	defer func() {
		if file != nil {
			file.Close()
			if f.lc.ShoudStop() {
				os.Remove(filePath)
			}
			os.Chtimes(id.String(), time.Now(), modTime)
		}
		close(eventChan)
	}()

	for cmd := range cmdChan {
		if f.lc.ShoudStop() {
			return
		}

		n, err := file.Write([]byte(cmd))
		copied += copyProgress(n)
		if err != nil {
			f.events.Push(m.Error{Id: id, Error: err})
			return
		}
		eventChan <- copied
	}
}
