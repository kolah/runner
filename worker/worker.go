package worker

import (
	"os/exec"
	"io"
	"os"
	"fmt"
)

type Worker struct {
	command         string
	arguments       []string
	stopChannel     chan bool
	FinishedChannel chan bool
}

func NewWorker(command string, arguments ...string) *Worker {
	return &Worker{
		command:         command,
		arguments:       arguments,
		stopChannel:     make(chan bool, 1),
		FinishedChannel: make(chan bool, 1),
	}
}

func (w *Worker) Run() {
	cmd := exec.Command(w.command, w.arguments...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}

	err = cmd.Start()
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}

	go io.Copy(os.Stderr, stderr)
	go io.Copy(os.Stdout, stdout)

	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Println(err)
		}
		w.FinishedChannel <- true
	}()

	go func() {
		<-w.stopChannel

		pid := cmd.Process.Pid
		fmt.Println("Killing PID", pid)
		if err := cmd.Process.Kill(); err != nil {
			fmt.Println("Error killing process", pid, err)
		}
		w.FinishedChannel <- true
	}()
}

func (w *Worker) Stop() {
	w.stopChannel <- true
	<-w.FinishedChannel
	w.FinishedChannel <- true
}
