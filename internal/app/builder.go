package app

import (
	"github.com/kballard/go-shellquote"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

type BuildErr struct {
	output string
}

func newBuildErr(o string) error {
	return BuildErr{output: o}
}

func (e BuildErr) Error() string {
	return e.output
}

type Builder struct {
	buildCommand string
	errorLogPath string
	logger       Logger
}

func NewBuilder(buildCommand, errorLogPath string, logger Logger) *Builder {
	return &Builder{
		buildCommand: buildCommand,
		errorLogPath: errorLogPath,
		logger:       logger,
	}
}

func (b *Builder) Build() error {
	b.removeBuildErrorsLog()

	b.logger.Info("Building...\n")

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
		return err
	}

	err = cmd.Start()
	if err != nil {
		b.logger.Infof("Unable to execute build %s\n", err.Error())

		return err
	}
	//noinspection ALL
	io.Copy(os.Stdout, stdout)
	errBuf, _ := ioutil.ReadAll(stderr)

	err = cmd.Wait()
	if err != nil {
		errorMessage := string(errBuf)
		b.createBuildErrorsLog(errorMessage)
		b.logger.Infof("Build failed %s\n", err.Error())

		return newBuildErr(errorMessage)
	}

	b.logger.Info("Build finished\n")

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
