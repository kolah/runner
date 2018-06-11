package worker

import (
	"log"
	"os/exec"
	"io"
	"os"
	"io/ioutil"
	"errors"
)

type Builder struct {
	rootPath   string
	outputPath string
}

func NewBuilder(rootPath string, outputPath string) *Builder {
	return &Builder{
		rootPath:   rootPath,
		outputPath: outputPath,
	}
}

func (b *Builder) Build() error {
	log.Println("Building...")

	cmd := exec.Command("go", "build", "-gcflags=-N", "-gcflags=-l", "-o", b.outputPath, b.rootPath)

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

	io.Copy(os.Stdout, stdout)
	errBuf, _ := ioutil.ReadAll(stderr)

	err = cmd.Wait()
	if err != nil {
		return errors.New(string(errBuf))
	}
	log.Println("Build finished")

	return nil
}
