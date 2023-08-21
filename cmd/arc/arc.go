package main

import (
	"log"
	"os"

	"arc/controller"
	"arc/files/file_fs"
	"arc/files/mock_fs"
	"arc/lifecycle"
	m "arc/model"
	"arc/renderer/tcell"
	"arc/stream"
)

func main() {
	var err any
	var stack []byte

	log.SetFlags(0)
	logFile, err := os.Create("log.log")
	if err == nil {
		log.SetOutput(logFile)
		defer func() {
			if err != nil {
				log.Printf("ERROR: %v", err)
				log.Println(string(stack))
				logFile.Close()
				log.SetOutput(os.Stderr)
				log.Printf("ERROR: %v", err)
			} else {
				logFile.Close()
			}
		}()
	}

	var paths []string
	if len(os.Args) >= 1 && (os.Args[1] == "-sim" || os.Args[1] == "-sim2") {
		paths = []string{"origin", "copy 1", "copy 2"}
	} else {
		paths = make([]string, len(os.Args)-1)
		for i, path := range os.Args[1:] {
			path, err := file_fs.AbsPath(path)
			paths[i] = path
			if err != nil {
				log.Panicf("Failed to scan archives: %#v", err)
			}
		}
	}

	lc := lifecycle.New()

	events := stream.NewStream[m.Event]("contr")
	renderer, err := tcell.NewRenderer(lc, events)
	if err != nil {
		log.Printf("Failed to open terminal: %#v", err)
		return
	}

	var fs m.FS

	if len(os.Args) >= 1 && os.Args[1] == "-sim" {
		fs = mock_fs.NewFs(events)
		mock_fs.Scan = true
	} else if len(os.Args) >= 1 && os.Args[1] == "-sim2" {
		fs = mock_fs.NewFs(events)
	} else {
		fs = file_fs.NewFs(events, lc)
	}

	err, stack = controller.Run(fs, renderer, events, paths)

	renderer.Quit()
	lc.Stop()
}
