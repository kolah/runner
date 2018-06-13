package worker

import (
	"sync"
	"fmt"
)

type RunnerMode string

const (
	RunnerModeLiveRebuild RunnerMode = "LIVEREBUILD"
	RunnerModeDebug       RunnerMode = "DEBUG"
)

type Runner struct {
	sync.Mutex
	worker *Worker
	mode RunnerMode
	executablePath string
}

func NewRunner(executablePath string) *Runner {
	return &Runner{
		mode: RunnerModeLiveRebuild,
		executablePath: executablePath,
	}
}

func (r *Runner) SetMode(mode RunnerMode) {
	r.Lock()
	defer r.Unlock()

	if r.worker != nil {
		r.worker.Stop()
	}

	fmt.Println("Switching mode to", mode)

	switch mode {
	case RunnerModeLiveRebuild:
		r.worker = NewWorker(r.executablePath)
		r.worker.Run()
		break
	case RunnerModeDebug:
		r.worker = NewWorker("dlv", "--headless", "--listen=:2345", "--api-version=2", "exec", r.executablePath)
		r.worker.Run()

		go func() {
			<-r.worker.FinishedChannel
			// return to live rebuild
			if r.mode != RunnerModeLiveRebuild {
				r.SetMode(RunnerModeLiveRebuild)
			}
		}()

		break
	}
}
