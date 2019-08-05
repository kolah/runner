package app

import (
	"errors"
	"github.com/kballard/go-shellquote"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

type Builder struct {
	buildCommand string
	errorLogPath string
}

func NewBuilder(buildCommand, errorLogPath string) *Builder {
	return &Builder{
		buildCommand: buildCommand,
		errorLogPath: errorLogPath,
	}
}

func (b *Builder) Build() error {
	b.removeBuildErrorsLog()

	log.Println("Building...")

	parts, err := shellquote.Split(b.buildCommand)

	if err != nil {
		return err
	}

	head := parts[0]
	parts = parts[1:]

	cmd := exec.Command(head, parts...)

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
	//noinspection ALL
	io.Copy(os.Stdout, stdout)
	errBuf, _ := ioutil.ReadAll(stderr)

	err = cmd.Wait()
	if err != nil {
		errorMessage := string(errBuf)
		b.createBuildErrorsLog(errorMessage)

		return errors.New(errorMessage)
	}
	log.Println("Build finished")

	return nil
}

func (b *Builder) createBuildErrorsLog(message string) {
	file, err := os.Create(b.errorLogPath)
	if err != nil {
		return
	}

	_, err = file.WriteString(message)
	if err != nil {
		return
	}

	return
}

func (b *Builder) removeBuildErrorsLog() {
	if _, err := os.Stat(b.errorLogPath); !os.IsNotExist(err) {
		_ = os.Remove(b.errorLogPath)
	}
}

