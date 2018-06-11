package worker

import (
	"os/exec"
	"log"
	"io"
	"os"
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
		cmd.Wait()
		w.FinishedChannel <- true
	}()

	go func() {
		<-w.stopChannel

		alreadyExited := cmd.ProcessState.Exited()
		pid := cmd.Process.Pid

		if !alreadyExited {
			log.Println("Killing PID", pid)
			cmd.Process.Kill()
		} else {
			log.Println("Process already exited", pid)
		}
	}()
}

func (w *Worker) Stop() {
	w.stopChannel <- true
}
