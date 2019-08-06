package app

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"github.com/kolah/runner/internal/pkg/set"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ListenerFunc func(event fsnotify.Event)

type Watcher struct {
	sync.Mutex
	watchDirs     set.Set
	ignoredDirs   set.Set
	watchPatterns set.Set
	file          []string
	watcher       *fsnotify.Watcher
	quit          chan bool
	verbose       bool
	listeners     []ListenerFunc
	logger        Logger
}

func NewWatcher(watchDirs []string, ignoredDirs []string, watchPatterns []string, logger Logger) *Watcher {
	return &Watcher{
		watchDirs:     set.NewSet(watchDirs),
		ignoredDirs:   set.NewSet(ignoredDirs),
		watchPatterns: set.NewSet(watchPatterns),
		verbose:       false,
		listeners:     make([]ListenerFunc, 0),
		logger:        logger,
	}
}

func (w *Watcher) Start() (err error) {
	w.logger.Debug("Watcher: starting\n")
	if w.watcher != nil {
		return errors.New("watcher already started")
	}

	w.watcher, err = fsnotify.NewWatcher()

	if err != nil {
		return err
	}

	w.quit = make(chan bool)

	for _, dir := range w.watchDirs.Values() {
		if err := w.AddRecursive(dir); err != nil {
			_ = w.Stop()
			return err
		}
	}

	w.watchLoop()

	return nil
}

// AddListener adds a listener function to run on event,
// the listener function will receive the event object as argument.
func (w *Watcher) AddListener(l ListenerFunc) {
	w.Lock()
	defer w.Unlock()

	w.listeners = append(w.listeners, l)
}

func (w *Watcher) Stop() error {
	w.logger.Debug("Watcher: stopping\n")
	if w.quit != nil {
		w.quit <- true
		close(w.quit)
	}

	if w.watcher != nil {
		return w.watcher.Close()
	}

	return nil
}

func (w *Watcher) AddRecursive(dir string) error {
	w.Lock()
	defer w.Unlock()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}

			if w.isIgnoredDir(path) {
				w.logger.Debugf("Watcher: ignoring \"%s\"\n", path)

				return filepath.SkipDir
			}

			w.logger.Debugf("Watcher: watching \"%s\"\n", path)
			if err := w.watcher.Add(path); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// send events to observers
func (w *Watcher) notify(event fsnotify.Event) {
	for _, listener := range w.listeners {
		go listener(event)
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	w.Lock()
	defer w.Unlock()

	w.logger.Debugf("Watcher: handling event for %s\n", event)
	// when new directory is created, add to watch
	if event.Op&fsnotify.Create == fsnotify.Create {
		info, err := os.Stat(event.Name)
		if err == nil && info.IsDir() {
			// noinspection ALL
			go w.AddRecursive(event.Name)
		}
	}

	if w.fileMatches(&event.Name) {
		w.logger.Debugf("Watcher: file matching pattern \"%s\"\n", event.Name)
		w.notify(event)
	}
}

func (w *Watcher) watchLoop() {
	go func() {
		for {
			select {
			case event := <-w.watcher.Events:
				w.handleEvent(event)
			case <-w.quit:
				return
			}
		}
	}()
}

func (w *Watcher) isIgnoredDir(path string) bool {
	paths := strings.Split(path, "/")
	if len(paths) <= 0 {
		return false
	}

	for _, e := range w.ignoredDirs.Values() {
		if strings.TrimSpace(e) == path {
			return true
		}
	}
	return false
}

func (w *Watcher) fileMatches(f *string) bool {
	if f == nil {
		return true
	}
	// check exact match
	match := w.watchPatterns.Has(*f)
	if match {
		return true
	}

	for _, p := range w.watchPatterns.Values() {
		match, _ = filepath.Match(p, filepath.Base(*f))
		if match {
			return true
		}
	}

	return false
}
