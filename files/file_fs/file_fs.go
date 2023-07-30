package file_fs

import (
	"arc/lifecycle"
	m "arc/model"
	"arc/stream"
	"os"
	"path/filepath"

	"golang.org/x/text/unicode/norm"
)

type fileFs struct {
	events   *stream.Stream[m.Event]
	lc       *lifecycle.Lifecycle
	commands *stream.Stream[m.FileCommand]
}

func NewFs(events *stream.Stream[m.Event], lc *lifecycle.Lifecycle) m.FS {
	fs := &fileFs{
		events:   events,
		lc:       lc,
		commands: stream.NewStream[m.FileCommand]("file-fs"),
	}

	go fs.handleEvents()

	return fs
}

func (fs *fileFs) Scan(root m.Root) {
	s := &scanner{
		root:    root,
		events:  fs.events,
		lc:      fs.lc,
		byInode: map[uint64]*m.Meta{},
		stored:  map[uint64]*m.Meta{},
		sent:    map[m.Id]struct{}{},
	}
	go s.scanArchive()
}

func (fs *fileFs) Send(cmd m.FileCommand) {
	fs.commands.Push(cmd)
}

func (fs *fileFs) handleEvents() {
	for {
		for _, cmd := range fs.commands.Pull() {
			fs.handleCommand(cmd)
		}
	}
}

func (fs *fileFs) handleCommand(cmd m.FileCommand) {
	fs.lc.Started()
	defer fs.lc.Done()

	switch cmd := cmd.(type) {
	case m.DeleteFile:
		fs.deleteFile(cmd)

	case m.RenameFile:
		fs.renameFile(cmd)

	case m.CopyFile:
		fs.copyFile(cmd)
	}
}

func AbsPath(path string) (string, error) {
	var err error
	path, err = filepath.Abs(path)
	path = norm.NFC.String(path)
	if err != nil {
		return "", err
	}

	_, err = os.Stat(path)
	if err != nil {
		return "", err
	}
	return path, nil
}
