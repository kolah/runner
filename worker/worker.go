package worker

import (
	"os/exec"
	"log"
	"io"
	"os"
)

type Worker struct {
	command     string
	arguments   []string
	stopChannel chan bool
}

func NewWorker(command string, arguments ...string) *Worker {
	return &Worker{
		command: command,
		arguments: arguments,
		stopChannel: make(chan bool, 1),
	}
}

func (w *Worker) Run() {
	cmd := exec.Command(w.command, w.arguments...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	go io.Copy(os.Stderr, stderr)
	go io.Copy(os.Stdout, stdout)

	go func() {
		<-w.stopChannel
		pid := cmd.Process.Pid
		log.Println("Killing PID", pid)
		cmd.Process.Kill()
	}()
}

func (w *Worker) Stop() {
	w.stopChannel <- true
}
