package app

import (
	"github.com/kballard/go-shellquote"
	"log"
	"os/exec"
	"io"
	"os"
	"fmt"
	"syscall"
)

type Worker struct {
	command         string
	arguments       []string
	stopChannel     chan bool
	FinishedChannel chan bool
}

func NewWorker(command string) *Worker {
	return &Worker{
		command:         command,
		stopChannel:     make(chan bool),
		FinishedChannel: make(chan bool, 1),
	}
}

func (w *Worker) Run() {
	log.Println("Running...")

	parts, err := shellquote.Split(w.command)

	if err != nil {
		fmt.Println("Error", err)
		return
	}

	head := parts[0]
	parts = parts[1:]

	cmd := exec.Command(head, parts...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

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
	//noinspection ALL
	go io.Copy(os.Stderr, stderr)
	//noinspection ALL
	go io.Copy(os.Stdout, stdout)

	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Println(err)
		}
		w.FinishedChannel <- true
	}()

	go func() {
		fmt.Println("Waiting for stop signal")
		<-w.stopChannel

		pid := cmd.Process.Pid
		fmt.Println("Killing PID", pid)

		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err == nil {
			_ = syscall.Kill(-pgid, 15) // note the minus sign
		}

		if err := cmd.Process.Kill(); err != nil {
			fmt.Println("Error killing process", pid, err)
		}
		w.FinishedChannel <- true
	}()
}

func (w *Worker) Stop() {
	w.stopChannel <- true
	<-w.FinishedChannel
}
