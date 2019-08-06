package app

import (
	"github.com/fsnotify/fsnotify"
	"runtime"
	"sync"
	"time"
)

type RunnerMode string

const (
	ModeRebuild RunnerMode = "REBUILD"
	ModeDebug   RunnerMode = "DEBUG"
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
	logger       Logger
	appLogger    *RunnerOutLog
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

func NewRunner(watcher *Watcher, builder *Builder, options RunnerOpts, logger Logger, appLogger *RunnerOutLog) *Runner {
	return &Runner{
		builder:      builder,
		mode:         ModeRebuild,
		watcher:      watcher,
		options:      options,
		eventsBuffer: make([]fsnotify.Event, 0),
		quit:         make(chan bool, 1),
		logger:       logger,
		appLogger:    appLogger,
	}
}

func (r *Runner) Start() error {
	if err := r.watcher.Start(); err != nil {
		return err
	}

	// don't stop on build error
	buildErr := false
	if err := r.Build(); err != nil {
		_, buildErr = err.(BuildErr)
		if !buildErr {
			return err
		}
	}

	// start worker only on successful initial build
	if !buildErr {
		r.worker = NewWorker(r.options.runCommand, r.logger, r.appLogger)
		if err := r.worker.Run(); err != nil {
			return err
		}
	}

	r.watcher.AddListener(func(event fsnotify.Event) {
		r.Lock()
		defer r.Unlock()

		if r.options.buildDelay == 0 {
			r.logger.Debug("Watched files changed, triggering event\n")
			r.events <- struct{}{}
			return
		}

		r.eventsBuffer = append(r.eventsBuffer, event)
		if len(r.eventsBuffer) == 1 {
			time.AfterFunc(r.options.buildDelay, func() {
				r.Lock()
				defer r.Unlock()
				r.logger.Debug("Watched files changed, triggering event after delay\n")

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

	r.logger.Infof("Switching mode to %s\n", mode)
	command := r.options.runCommand
	if mode == ModeDebug {
		command = r.options.debugCommand
		if r.options.buildBeforeDebug {
			err := r.Build()
			if err != nil {
				r.logger.Infof("Build error: %s\n", err)
				return
			}
		}
	}

	r.worker = NewWorker(command, r.logger, r.appLogger)
	if err := r.worker.Run(); err != nil {
		r.logger.Infof("Failed to run \"%s\", %s", command, err.Error())
	}
}

func (r *Runner) mainLoop() {
	for {
		r.loopIndex++

		r.logger.Infof("Waiting (loop %d)...\n", r.loopIndex)
		<-r.events

		r.logger.Debugf("Rebuild triggered! (%d Go routines)\n", runtime.NumGoroutine())

		if r.Mode() == ModeDebug {
			r.logger.Debug("ignoring code changes while debugging\n")
			continue
		}

		err := r.Build()
		if err == nil {
			r.SetMode(r.Mode())
		}

		select {
		case <-r.quit:
			return
		default:
			continue
		}
	}
}
