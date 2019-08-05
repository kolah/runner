package app

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/kolah/runner/internal/pkg/set"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Listener func(event fsnotify.Event)

type Watcher struct {
	sync.Mutex
	watchDirs     set.Set
	ignoredDirs   set.Set
	watchPatterns set.Set
	file          []string
	watcher       *fsnotify.Watcher
	quit          chan bool
	verbose       bool
	listeners     []Listener
}

func NewWatcher(watchDirs []string, ignoredDirs []string, watchPatterns []string) *Watcher {
	return &Watcher{
		watchDirs:     set.NewSet(watchDirs),
		ignoredDirs:   set.NewSet(ignoredDirs),
		watchPatterns: set.NewSet(watchPatterns),
		verbose:       false,
		listeners:     make([]Listener, 0),
	}
}

func (w *Watcher) Verbose() *Watcher {
	w.verbose = true

	return w
}

func (w *Watcher) Silent() *Watcher {
	w.verbose = false

	return w
}

func (w *Watcher) Start() (err error) {
	w.logVerbose("Watcher: starting")
	if w.watcher != nil {
		return fmt.Errorf("watcher already started")
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
func (w *Watcher) AddListener(l Listener) {
	w.Lock()
	defer w.Unlock()

	w.listeners = append(w.listeners, l)
}

func (w *Watcher) Stop() error {
	w.logVerbose("Watcher: stopping")
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
				if w.verbose {
					w.logVerbose(fmt.Sprintf("Watcher: \"%s\" ignored", path))
				}
				return filepath.SkipDir
			}

			w.logVerbose(fmt.Sprintf("Watcher: \"%s\" watching", path))
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

	w.logVerbose(fmt.Sprintf("Watcher: handling event for %s", event))
	// when new directory is created, add to watch
	if event.Op&fsnotify.Create == fsnotify.Create {
		info, err := os.Stat(event.Name)
		if err == nil && info.IsDir() {
			go w.AddRecursive(event.Name)
		}
	}

	if w.fileMatches(&event.Name) {
		w.logVerbose(fmt.Sprintf("Watcher: file matching pattern \"%s\"", event.Name))
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

func (w *Watcher) logVerbose(msg string) {
	if w.verbose {
		log.Println(msg)
	}
}
