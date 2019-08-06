package app

import (
	"github.com/kballard/go-shellquote"
	"github.com/shirou/gopsutil/process"
	"io"
	"os/exec"
)

type Worker struct {
	command   string
	arguments []string
	quit      chan bool
	finished  chan bool
	logger    Logger
	appLogger *RunnerOutLog
}

func NewWorker(command string, logger Logger, appLogger *RunnerOutLog) *Worker {
	return &Worker{
		command:   command,
		quit:      make(chan bool),
		finished:  make(chan bool, 1),
		logger:    logger,
		appLogger: appLogger,
	}
}

func (w *Worker) Run() error {
	w.logger.Infof("Running %s...\n", w.command)

	parts, err := shellquote.Split(w.command)

	if err != nil {
		w.logger.Infof("Error parsing command \"%s\": %s\n", w.command, err.Error())
		return nil
	}

	head := parts[0]
	parts = parts[1:]

	cmd := exec.Command(head, parts...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		w.logger.Infof("Cannot execute command \"%s\": %s", w.command, err.Error())
		return nil
	}
	//noinspection ALL
	go io.Copy(w.appLogger.errWriter, stderr)
	//noinspection ALL
	go io.Copy(w.appLogger.outWriter, stdout)

	go func() {
		if err := cmd.Wait(); err != nil {
			w.logger.Debugf("Error while waiting for process to finish: %s", err.Error())
		}
		w.finished <- true
	}()

	go func() {
		<-w.quit

		pid := cmd.Process.Pid
		if err := w.killChildProcesses(int32(pid)); err != nil {
			w.logger.Debugf("Error killing children processes %d: %s\n", pid, err.Error())
		}

		if err := cmd.Process.Kill(); err != nil {
			w.logger.Debugf("Error killing process %d: %s", pid, err.Error())
		}
		w.finished <- true
	}()

	return nil
}

func (w *Worker) Stop() {
	w.quit <- true
	<-w.finished
}

func (w *Worker) killChildProcesses(pid int32) error {
	proc, err := process.NewProcess(pid)

	if err != nil {
		return err
	}

	children, err := proc.Children()
	if err != nil {
		return err
	}

	if len(children) == 0 {
		return nil
	}

	for _, child := range children {
		_ = child.Kill()
	}

	return nil
}
