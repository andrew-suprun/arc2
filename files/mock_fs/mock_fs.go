package mock_fs

import (
	m "arc/model"
	"arc/stream"
	"log"
	"math/rand"
	"time"
)

var Scan bool

type mockFs struct {
	events   *stream.Stream[any]
	commands *stream.Stream[m.FileCommand]
}

type scanner struct {
	root   string
	events *stream.Stream[any]
}

func NewFs(eventStream *stream.Stream[any]) m.FS {
	fs := &mockFs{
		events:   eventStream,
		commands: stream.NewStream[m.FileCommand]("mock-fs"),
	}
	go fs.handleCommands()
	return fs
}

func (fs *mockFs) Scan(root string) {
	s := &scanner{
		root:   root,
		events: fs.events,
	}
	go s.scanArchive()
}

func (s *mockFs) Send(cmd m.FileCommand) {
	s.commands.Push(cmd)
}

func (fs *mockFs) handleCommands() {
	for {
		cmds, _ := fs.commands.Pull()
		for _, cmd := range cmds {
			fs.handleCommand(cmd)
		}
	}
}

func (fs *mockFs) handleCommand(cmd m.FileCommand) {
	log.Printf("mock: cmd: %T: %v", cmd, cmd)
	switch cmd := cmd.(type) {
	case m.DeleteFile:
		fs.events.Push(m.FileDeleted(cmd))

	case m.RenameFile:
		fs.events.Push(m.FileRenamed(cmd))

	case m.CopyFile:
		size := 0
		for _, meta := range metas {
			if cmd.From
		}
		for _, meta := range metas[cmd.From.Root] {
			if meta.Id.Name == cmd.From.NaPath
				for copied := uint64(0); ; copied += 10000 {
					if copied > meta.size {
						copied = meta.size
					}
					fs.events.Push(m.CopyingProgress(copied))
					if copied == meta.size {
						break
					}
					time.Sleep(time.Millisecond)
				}
				break
			}
		}
		fs.events.Push(m.FileCopied(cmd))
	}
}

func (s *scanner) scanArchive() {
	archFiles := metas[s.root]
	totalSize := uint64(0)
	for _, file := range archFiles {
		totalSize += file.size
	}

	for _, meta := range archFiles {
		meta := &m.Meta{
			Id:      meta.Id,
			Size:    meta.size,
			ModTime: meta.modTime,
		}
		s.events.Push(m.FileScanned{
			Meta: *meta,
		})

	}

	s.events.Push(m.ArchiveScanned{
		Root: s.root,
	})

	scans := make([]bool, len(archFiles))

	for i := range archFiles {
		scans[i] = Scan
	}
	for i := range archFiles {
		if !scans[i] {
			meta := archFiles[i]
			s.events.Push(m.FileHashed{
				Id:   meta.Id,
				Hash: meta.Hash,
			})
		}
	}
	for i := range archFiles {
		if scans[i] {
			meta := archFiles[i]
			for hashed := uint64(0); ; hashed += 50000 {
				if hashed > meta.size {
					hashed = meta.size
				}
				s.events.Push(m.HashingProgress{Id: meta.Id, Hashed: hashed})
				if hashed == meta.size {
					break
				}
				time.Sleep(time.Millisecond)
			}
			s.events.Push(m.FileHashed{
				Id:   meta.Id,
				Hash: meta.Hash,
			})
		}
	}

	s.events.Push(m.ArchiveHashed{
		Root: s.root,
	})
}

var beginning = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
var end = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
var duration = end.Sub(beginning)

type fileMeta struct {
	root    string
	name    string
	hash    string
	size    uint64
	modTime time.Time
}

var sizes = map[string]uint64{}
var modTimes = map[string]time.Time{}
var inode = uint64(0)

func init() {
	for _, meta := range metas {
		size, ok := sizes[meta.hash]
		if !ok {
			size = uint64(rand.Intn(100000000))
			meta.size = size
			sizes[meta.hash] = size
		}
		modTime, ok := modTimes[meta.hash]
		if !ok {
			modTime = beginning.Add(time.Duration(rand.Int63n(int64(duration))))
			meta.modTime = modTime
			modTimes[meta.hash] = modTime
		}
	}
}

var metas = []*fileMeta{
	{root: "origin", name: "b/0000", hash: "0000"},
	{root: "copy 1", name: "1111", hash: "0000"},
	{root: "copy 1", name: "b/0000", hash: "0000"},
	{root: "copy 2", name: "2222", hash: "0000"},
	{root: "copy 2", name: "b/0000", hash: "0000"},
	// ----
	{root: "copy 2", name: "q/w/e/r/t/y.txt", hash: "12345"},
	// ----
	{root: "copy 1", name: "4444", hash: "4444"},
	{root: "copy 2", name: "5555", hash: "4444"},
	// ----
	{root: "origin", name: "6666", hash: "6666"},
	{root: "copy 1", name: "6666", hash: "6666"},
	{root: "copy 2", name: "7777", hash: "6666"},
	// ----
	{root: "origin", name: "7777", hash: "7777"},
	{root: "copy 1", name: "7777", hash: "7777"},
	{root: "copy 2", name: "6666", hash: "7777"},
	// ----
	{root: "copy 2", name: "8888", hash: "8888"},
	// ----
	{root: "copy 1", name: "8888", hash: "9999"},
	{root: "copy 1", name: "9999", hash: "9999"},
	{root: "copy 2", name: "9999", hash: "9999"},
	// ----
	{root: "origin", name: "a/b/c/d", hash: "abcd"},
	{root: "copy 1", name: "a/b", hash: "abcd"},
	// ----
	{root: "copy 2", name: "x", hash: "asdfg"},
	// ----
	{root: "copy 1", name: "b/bbb.txt", hash: "bbbb"},
	{root: "copy 2", name: "c/ccc.txt", hash: "bbbb"},
	// ----
	{root: "origin", name: "bla", hash: "bla"},
	{root: "copy 1", name: "bla", hash: "bla"},
	{root: "copy 2", name: "bla", hash: "bla"},
	// ----
	{root: "origin", name: "different", hash: "different"},
	// ----
	{root: "copy 1", name: "different", hash: "different-copy1"},
	// ----
	{root: "copy 2", name: "different", hash: "different-copy2"},
	// ----
	{root: "origin", name: "a/b/e/f.txt", hash: "gggg"},
	{root: "copy 1", name: "y.txt", hash: "gggg"},
	{root: "copy 2", name: "a/b/e/x.txt", hash: "gggg"},
	// ----
	{root: "origin", name: "a/b/e/h.txt", hash: "hhhh"},
	{root: "origin", name: "uuu.txt", hash: "hhhh"},
	{root: "origin", name: "qqq.txt", hash: "hhhh"},
	{root: "origin", name: "x/xxx.txt", hash: "hhhh"},
	{root: "copy 1", name: "a/b/e/f.txt", hash: "hhhh"},
	{root: "copy 1", name: "x/xxx.txt", hash: "hhhh"},
	{root: "copy 1", name: "zzz.txt", hash: "hhhh"},
	// ----
	{root: "copy 1", name: "a/b/c/d.txt", hash: "llll"},
	// ----
	{root: "copy 1", name: "qqq.txt", hash: "mmmm"},
	// ----
	{root: "origin", name: "q/w/e/r/t/y.txt", hash: "qwerty"},
	// ----
	{root: "origin", name: "same", hash: "same"},
	// ----
	{root: "copy 1", name: "same", hash: "same-copy"},
	{root: "copy 2", name: "same", hash: "same-copy"},
	// ----
	{root: "origin", name: "a/b/e/g.txt", hash: "tttt"},
	{root: "copy 1", name: "a/b/e/g.txt", hash: "tttt"},
	{root: "copy 2", name: "a/b/e/g.txt", hash: "tttt"},
	// ----
	{root: "origin", name: "xxx.txt", hash: "xxxx"},
	{root: "copy 1", name: "xxx.txt", hash: "xxxx"},
	{root: "copy 2", name: "xxx.txt", hash: "xxxx"},
	// ----
	{root: "origin", name: "xyz/bla", hash: "xyz/bla"},
	{root: "copy 1", name: "xyz/bla", hash: "xyz/bla"},
	{root: "copy 2", name: "xyz/bla", hash: "xyz/bla"},
	// ----
	{root: "origin", name: "yyy.txt", hash: "yyyy"},
	{root: "copy 1", name: "yyy.txt", hash: "yyyy"},
	// ----
	{root: "copy 1", name: "x/y/z.txt", hash: "zzzz"},
}
