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
	debuggerPort int
}

func NewRunner(executablePath string, debuggerPort int) *Runner {
	return &Runner{
		mode: RunnerModeLiveRebuild,
		executablePath: executablePath,
		debuggerPort: debuggerPort,
	}
}

func (r *Runner) GetMode() RunnerMode {
	return r.mode
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
		r.worker = NewWorker("dlv", "--headless", fmt.Sprintf("--listen=:%d", r.debuggerPort), "--api-version=2", "exec", r.executablePath)
		r.worker.Run()

		break
	}
}
