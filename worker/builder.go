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
	rootPath     string
	outputPath   string
	errorLogPath string
}

func NewBuilder(rootPath string, outputPath string, errorLogPath string) *Builder {
	return &Builder{
		rootPath:   rootPath,
		outputPath: outputPath,
		errorLogPath: errorLogPath,
	}
}

func (b *Builder) Build() error {
	b.removeBuildErrorsLog()

	log.Println("Building...")

	cmd := exec.Command("go", "build", "-gcflags", "all=-N -l", "-o", b.outputPath, b.rootPath)

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
		b.createBuildErrorsLog(string(errBuf))
		return errors.New("build failed")
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
	if _, err := os.Stat(b.errorLogPath); os.IsExist(err) {
		os.Remove(b.errorLogPath)
	}
}

