package file_fs

import (
	"arc/lifecycle"
	m "arc/model"
	"arc/stream"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/text/unicode/norm"
)

const hashFileName = ".meta.csv"

type scanner struct {
	root   string
	events *stream.Stream[any]
	lc     *lifecycle.Lifecycle
	metas  map[uint64]*m.Meta
	hashes map[uint64]string
	iNodes []uint64
}

func (s *scanner) scanArchive() {
	defer func() {
		s.events.Push(m.ArchiveHashed{
			Root: s.root,
		})
	}()

	fsys := os.DirFS(s.root.String())
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if s.lc.ShoudStop() || !d.Type().IsRegular() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		if err != nil {
			s.events.Push(m.Error{
				Id:    m.Id{Root: s.root, Path: m.Path(path).ParentName()},
				Error: err})
			return nil
		}

		meta, err := d.Info()
		if err != nil {
			s.events.Push(m.Error{
				Id:    m.Id{Root: s.root, Path: m.Path(path).ParentName()},
				Error: err})
			return nil
		}
		sys := meta.Sys().(*syscall.Stat_t)
		modTime := meta.ModTime()
		modTime = modTime.UTC().Round(time.Second)

		file := &m.Meta{
			Id:      m.Id{Root: s.root, Path: m.Path(path).ParentName()},
			ModTime: modTime,
			Size:    uint64(meta.Size()),
		}

		s.metas[sys.Ino] = file
		s.iNodes = append(s.iNodes, sys.Ino)

		s.events.Push(m.FileScanned{
			Meta: *file,
		})

		return nil
	})

	s.events.Push(m.ArchiveScanned{
		Root: s.root,
	})

	s.readMeta()
	defer func() {
		s.storeMeta()
	}()

	for ino, hash := range s.hashes {
		file := s.metas[ino]
		s.events.Push(m.FileHashed{Id: file.Id, Hash: hash})
	}

	for _, ino := range s.iNodes {
		if _, ok := s.hashes[ino]; ok {
			continue
		}
		file := s.metas[ino]

		hash := s.hashFile(file.Id)
		s.hashes[ino] = hash

		if s.lc.ShoudStop() {
			return
		}

		s.events.Push(m.FileHashed{Id: file.Id, Hash: hash})
	}
}

func (s *scanner) hashFile(id m.Id) m.Hash {
	hash := sha256.New()
	buf := make([]byte, 1024*1024)
	var hashed uint64

	fsys := os.DirFS(id.Root.String())
	file, err := fsys.Open(id.Path.String())
	if err != nil {
		s.events.Push(m.Error{Id: id, Error: err})
		return ""
	}
	defer file.Close()

	for {
		if s.lc.ShoudStop() {
			return ""
		}

		nr, er := file.Read(buf)
		if nr > 0 {
			nw, ew := hash.Write(buf[0:nr])
			if ew != nil {
				if err != nil {
					s.events.Push(m.Error{Id: id, Error: err})
					return ""
				}
			}
			if nr != nw {
				s.events.Push(m.Error{Id: id, Error: err})
				return ""
			}
		}

		if er == io.EOF {
			break
		}
		if er != nil {
			s.events.Push(m.Error{Id: id, Error: er})
			return ""
		}

		hashed += uint64(nr)
		s.events.Push(m.HashingProgress{
			Id:     id,
			Hashed: hashed,
		})
	}
	return m.Hash(base64.RawURLEncoding.EncodeToString(hash.Sum(nil)))
}

func (s *scanner) readMeta() {
	absHashFileName := filepath.Join(s.root.String(), hashFileName)
	hashInfoFile, err := os.Open(absHashFileName)
	if err != nil {
		return
	}
	defer hashInfoFile.Close()

	records, err := csv.NewReader(hashInfoFile).ReadAll()
	if err != nil || len(records) == 0 {
		return
	}

	for _, record := range records[1:] {
		if len(record) == 5 {
			iNode, er1 := strconv.ParseUint(record[0], 10, 64)
			size, er2 := strconv.ParseUint(record[2], 10, 64)
			modTime, er3 := time.Parse(time.RFC3339, record[3])
			modTime = modTime.UTC().Round(time.Second)
			hash := record[4]
			if hash == "" || er1 != nil || er2 != nil || er3 != nil {
				continue
			}

			info, ok := s.metas[iNode]
			if hash != "" && ok && info.ModTime == modTime && info.Size == size {
				s.hashes[iNode] = m.Hash(hash)
			}
		}
	}
}

func (s *scanner) storeMeta() error {
	result := make([][]string, 1, len(s.metas)+1)
	result[0] = []string{"INode", "Name", "Size", "ModTime", "Hash"}

	for iNode, hash := range s.hashes {
		file := s.metas[iNode]
		result = append(result, []string{
			fmt.Sprint(iNode),
			norm.NFC.String(file.Name.String()),
			fmt.Sprint(file.Size),
			file.ModTime.UTC().Format(time.RFC3339Nano),
			hash.String(),
		})
	}

	absHashFileName := filepath.Join(s.root.String(), hashFileName)
	hashInfoFile, err := os.Create(absHashFileName)

	if err != nil {
		return err
	}
	err = csv.NewWriter(hashInfoFile).WriteAll(result)
	hashInfoFile.Close()
	return err
}
