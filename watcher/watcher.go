package watcher

import (
	"github.com/windmilleng/fsnotify"
	"log"
	"os"
	"strings"
	"path/filepath"
)

type Watcher struct {
	rootDir         string
	ignoredDirs     []string
	validExtensions []string
	tmpDir          string
	stopChannel     chan bool
	EventChannel    chan string
}

func NewWatcher(rootDir string, tmpDir string, ignoredDirectories []string, validExtensions []string) *Watcher {
	return &Watcher{
		rootDir:         rootDir,
		ignoredDirs:     ignoredDirectories,
		validExtensions: validExtensions,
		tmpDir:          tmpDir,
		EventChannel:    make(chan string, 1000),
		stopChannel:     make(chan bool),
	}
}

func (w *Watcher) watchDirectory(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				if w.isWatchedFile(ev.Name) {
					log.Printf("Watcher: sending event %s", ev)
					w.EventChannel <- ev.String()
				}
			case err := <-watcher.Errors:
				log.Printf("Watcher: error: %s", err)
			case <-w.stopChannel:
				watcher.Close()
				return
			}
		}
	}()

	log.Printf("Watcher: Watching %s", path)
	err = watcher.Add(path)

	if err != nil {
		log.Fatal(err)
	}
}

func (w *Watcher) Start() {
	filepath.Walk(w.rootDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !w.isTmpDir(path) {
			if len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}

			if w.isIgnoredFolder(path) {
				log.Printf("Watcher: Ignoring %s", path)
				return filepath.SkipDir
			}

			w.watchDirectory(path)
		}

		return err
	})
}

func (w *Watcher) Stop() {
	w.stopChannel <- true
}

func (w *Watcher) isTmpDir(path string) bool {
	absolutePath, _ := filepath.Abs(path)
	absoluteTmpPath, _ := filepath.Abs(w.tmpDir)

	return absolutePath == absoluteTmpPath
}

func (w *Watcher) isIgnoredFolder(path string) bool {
	paths := strings.Split(path, "/")
	if len(paths) <= 0 {
		return false
	}

	for _, e := range w.ignoredDirs {
		if strings.TrimSpace(e) == paths[0] {
			return true
		}
	}
	return false
}

func (w *Watcher) isWatchedFile(path string) bool {
	absolutePath, _ := filepath.Abs(path)
	absoluteTmpPath, _ := filepath.Abs(w.tmpDir)

	if strings.HasPrefix(absolutePath, absoluteTmpPath) {
		return false
	}

	ext := filepath.Ext(path)

	for _, e := range w.validExtensions {
		if strings.TrimSpace(e) == ext {
			return true
		}
	}

	return false
}
