package app

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"
)

type RunnerMode string

const (
	RunnerModeLiveRebuild RunnerMode = "LIVEREBUILD"
	RunnerModeDebug       RunnerMode = "DEBUG"
)

type Runner struct {
	sync.Mutex
	worker       *Worker
	builder      *Builder
	watcher      *Watcher
	mode         RunnerMode
	options      RunnerOpts
	loopIndex    int
	events       chan interface{}
	eventsBuffer []fsnotify.Event
	quit         chan bool
}

type RunnerOpts struct {
	buildDelay       time.Duration
	runCommand       string
	debugCommand     string
	buildBeforeDebug bool
}

func NewRunnerOptions(buildDelay time.Duration, runCommand string, debugCommand string, buildBeforeDebug bool) RunnerOpts {
	return RunnerOpts{buildDelay: buildDelay, runCommand: runCommand, debugCommand: debugCommand, buildBeforeDebug: buildBeforeDebug}
}

func NewRunner(watcher *Watcher, builder *Builder, options RunnerOpts) *Runner {
	return &Runner{
		builder:      builder,
		mode:         RunnerModeLiveRebuild,
		watcher:      watcher,
		options:      options,
		eventsBuffer: make([]fsnotify.Event, 0),
		quit:         make(chan bool, 1),
	}
}

func (r *Runner) Start() error {
	if err := r.watcher.Start(); err != nil {
		return err
	}

	if err := r.Build(); err != nil {
		return err
	}

	r.worker = NewWorker(r.options.runCommand)
	r.worker.Run()

	r.watcher.AddListener(func(event fsnotify.Event) {
		r.Lock()
		defer r.Unlock()

		if r.options.buildDelay == 0 {
			r.events <- struct{}{}
			return
		}

		r.eventsBuffer = append(r.eventsBuffer, event)
		if len(r.eventsBuffer) == 1 {
			time.AfterFunc(r.options.buildDelay, func() {
				r.Lock()
				defer r.Unlock()
				fmt.Println("Len: ", len(r.eventsBuffer))
				r.events <- struct{}{}

				// reset events buffer
				r.eventsBuffer = make([]fsnotify.Event, 0)
			})
		}
	})

	r.events = make(chan interface{})

	go r.mainLoop()

	return nil
}

func (r *Runner) Stop() error {
	r.quit <- true
	if r.worker != nil {
		r.worker.Stop()
	}

	return r.watcher.Stop()
}

func (r *Runner) Build() error {
	return r.builder.Build()
}

func (r *Runner) Mode() RunnerMode {
	return r.mode
}

func (r *Runner) SetMode(mode RunnerMode) {
	r.Lock()
	defer r.Unlock()

	if r.worker != nil {
		r.worker.Stop()
	}

	fmt.Println("Switching mode to", mode)
	command := r.options.runCommand
	if mode == RunnerModeDebug {
		command = r.options.debugCommand
		if r.options.buildBeforeDebug {
			err := r.Build()
			if err != nil {
				log.Print(err)
			}
		}
	}

	r.worker = NewWorker(command)
	r.worker.Run()
}

func (r *Runner) mainLoop() {
	for {
		r.loopIndex++

		fmt.Printf("Waiting (loop %d)...\n", r.loopIndex)
		<-r.events

		fmt.Printf("Started! (%d Goroutines)\n", runtime.NumGoroutine())

		if r.Mode() == RunnerModeDebug {
			fmt.Println("ignoring code changes while debugging")
			continue
		}

		err := r.Build()
		if err == nil {
			r.SetMode(r.Mode())
		}

		fmt.Printf(strings.Repeat("-", 20) + "\n")
		select {
		case <-r.quit:
			return
		default:
			continue
		}
	}
}

func (r *Runner) flushFSEvents() {
	for {
		select {
		case eventName := <-r.events:
			fmt.Printf("flushing event %s\n", eventName)
		default:
			return
		}
	}
}
