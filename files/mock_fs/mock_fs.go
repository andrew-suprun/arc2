package mock_fs

import (
	m "arc/model"
	"arc/stream"
	"log"
	"math/rand"
	"path/filepath"
	"sort"
	"time"
)

var Scan bool

type mockFs struct {
	eventStream *stream.Stream[m.Event]
	commands    *stream.Stream[m.FileCommand]
}

type scanner struct {
	root        m.Root
	eventStream *stream.Stream[m.Event]
}

func NewFs(eventStream *stream.Stream[m.Event]) m.FS {
	fs := &mockFs{
		eventStream: eventStream,
		commands:    stream.NewStream[m.FileCommand]("mock-fs"),
	}
	go fs.handleCommands()
	return fs
}

func (fs *mockFs) Scan(root m.Root) {
	s := &scanner{
		root:        root,
		eventStream: fs.eventStream,
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
		fs.eventStream.Push(m.FileDeleted(cmd))

	case m.RenameFile:
		fs.eventStream.Push(m.FileRenamed(cmd))

	case m.CopyFile:
		for _, meta := range metas[cmd.From.Root] {
			if meta.Id.Name == cmd.From.Name {
				for copied := uint64(0); ; copied += 10000 {
					if copied > meta.Size {
						copied = meta.Size
					}
					fs.eventStream.Push(m.CopyingProgress(copied))
					if copied == meta.Size {
						break
					}
					time.Sleep(time.Millisecond)
				}
				break
			}
		}
		fs.eventStream.Push(m.FileCopied(cmd))
	}
}

func (s *scanner) scanArchive() {
	archFiles := metas[s.root]
	totalSize := uint64(0)
	for _, file := range archFiles {
		totalSize += file.Size
	}

	for _, meta := range archFiles {
		meta := &m.Meta{
			Id:      meta.Id,
			Size:    meta.Size,
			ModTime: meta.ModTime,
		}
		s.eventStream.Push(m.FileScanned{
			Meta: meta,
		})

	}

	s.eventStream.Push(m.ArchiveScanned{
		Root: s.root,
	})

	scans := make([]bool, len(archFiles))

	for i := range archFiles {
		scans[i] = Scan
	}
	for i := range archFiles {
		if !scans[i] {
			meta := archFiles[i]
			s.eventStream.Push(m.FileHashed{
				Id:   meta.Id,
				Hash: meta.Hash,
			})
		}
	}
	for i := range archFiles {
		if scans[i] {
			meta := archFiles[i]
			for hashed := uint64(0); ; hashed += 50000 {
				if hashed > meta.Size {
					hashed = meta.Size
				}
				s.eventStream.Push(m.HashingProgress{Id: meta.Id, Hashed: hashed})
				if hashed == meta.Size {
					break
				}
				time.Sleep(time.Millisecond)
			}
			s.eventStream.Push(m.FileHashed{
				Id:   meta.Id,
				Hash: meta.Hash,
			})
		}
	}

	s.eventStream.Push(m.ArchiveHashed{
		Root: s.root,
	})
}

var beginning = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
var end = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
var duration = end.Sub(beginning)

type fileMeta struct {
	INode uint64
	m.Id
	m.Hash
	Size    uint64
	ModTime time.Time
}

var sizeByName = map[string]uint64{}
var sizeByHash = map[m.Hash]uint64{}
var modTimes = map[m.Hash]time.Time{}
var inode = uint64(0)

func init() {
	sizeByHash["yyyy"] = 50000000
	sizeByHash["hhhh"] = 50000000
	for root, metaStrings := range metaMap {
		for name, hash := range metaStrings {
			size, ok := sizeByName[name]
			if !ok {
				size, ok = sizeByHash[hash]
				if !ok {
					size = uint64(rand.Intn(100000000))
					sizeByHash[hash] = size
				}
			}
			modTime, ok := modTimes[hash]
			if !ok {
				modTime = beginning.Add(time.Duration(rand.Int63n(int64(duration))))
				modTimes[hash] = modTime
			}
			inode++
			file := &fileMeta{
				INode:   inode,
				Id:      m.Id{Root: root, Name: m.Name{Path: dir(name), Base: base(name)}},
				Hash:    hash,
				Size:    size,
				ModTime: modTime,
			}
			metas[root] = append(metas[root], file)
		}
		slice := metas[root]
		sort.Slice(slice, func(i, j int) bool {
			return slice[i].Id.String() < slice[j].Id.String()
		})
		metas[root] = slice
	}
}

var metas = map[m.Root][]*fileMeta{}
var metaMap = map[m.Root]map[string]m.Hash{
	"origin": {
		"a/b/c/d": "abcd",
		// "0000":            "0000",
		// "6666":            "6666",
		// "7777":            "7777",
		// "a/b/e/f.txt":     "gggg",
		// "a/b/e/g.txt":     "tttt",
		// "a/b/e/h.txt":     "hhhh",
		// "x/xxx.txt":       "hhhh",
		// "q/w/e/r/t/y.txt": "qwerty",
		// "qqq.txt":         "hhhh",
		// "uuu.txt":         "hhhh",
		// "xxx.txt":         "xxxx",
		// "yyy.txt":         "yyyy",
		// "same":            "same",
		// "different":       "different",
		// "bla":             "bla",
		// "xyz/bla":         "xyz/bla",
	},
	"copy 1": {
		"a/b": "abcd",
		// "xxx.txt":     "xxxx",
		// "a/b/c/d.txt": "llll",
		// "a/b/e/f.txt": "hhhh",
		// "a/b/e/g.txt": "tttt",
		// "qqq.txt":     "mmmm",
		// "y.txt":       "gggg",
		// "x/xxx.txt":   "hhhh",
		// "zzz.txt":     "hhhh",
		// "x/y/z.txt":   "zzzz",
		// "yyy.txt":     "yyyy",
		// "1111":        "0000",
		// "9999":        "9999",
		// "4444":        "4444",
		// "8888":        "9999",
		// "b/bbb.txt":   "bbbb",
		// "6666":        "6666",
		// "7777":        "7777",
		// "same":        "same-copy",
		// "different":   "different-copy1",
		// "bla":         "bla",
		// "xyz/bla":     "xyz/bla",
	},
	"copy 2": {
		// "xxx.txt":         "xxxx",
		// "a/b/e/x.txt":     "gggg",
		// "a/b/e/g.txt":     "tttt",
		// "x":               "asdfg",
		// "q/w/e/r/t/y.txt": "12345",
		// "2222":            "0000",
		// "9999":            "9999",
		// "5555":            "4444",
		// "6666":            "7777",
		// "7777":            "6666",
		// "8888":            "8888",
		// "c/ccc.txt":       "bbbb",
		// "same":            "same-copy",
		// "different":       "different-copy2",
		// "bla":             "bla",
		// "xyz/bla":         "xyz/bla",
	},
}

func dir(path string) m.Path {
	path = filepath.Dir(path)
	if path == "." {
		return ""
	}
	return m.Path(path)
}

func base(path string) m.Base {
	return m.Base(filepath.Base(path))
}
