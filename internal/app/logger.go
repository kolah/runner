package app

import (
	"fmt"
	"github.com/gookit/color"
	"strings"
)

type LogLevel int

const (
	InfoLevel LogLevel = iota
	DebugLevel
)

// ParseLevel takes a string level and returns the log level constant.
func ParseLevel(lvl string) (LogLevel, error) {
	switch strings.ToLower(lvl) {
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	}

	var l LogLevel
	return l, fmt.Errorf("not a valid log level: %q", lvl)
}

type Logger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
}

type processorFunc func(message string) string

type StdoutLog struct {
	level     LogLevel
}

type RunnerOutLog struct {
	errWriter *WorkerLogWriter
	outWriter *WorkerLogWriter
}

func NewAppLog(logger Logger) *RunnerOutLog {
	return &RunnerOutLog{
		errWriter: NewWorkerLogWriter(logger, func(m string) string {
			return color.ResetSet + color.Red.Sprint(m)
		}),
		outWriter: NewWorkerLogWriter(logger, func(m string) string {
			return color.ResetSet + color.Green.Sprint(m)
		}),
	}
}

func NewStdoutLog(level LogLevel) *StdoutLog {
	return &StdoutLog{level: level}
}

func (l *StdoutLog) Info(args ...interface{}) {
	if l.level < InfoLevel {
		return
	}

	fmt.Print(args...)
}

func (l *StdoutLog) Infof(format string, args ...interface{}) {
	if l.level < InfoLevel {
		return
	}

	fmt.Printf(format, args...)
}

func (l *StdoutLog) Debug(args ...interface{}) {
	if l.level < DebugLevel {
		return
	}

	fmt.Print(args...)
}

func (l *StdoutLog) Debugf(format string, args ...interface{}) {
	if l.level < DebugLevel {
		return
	}

	fmt.Printf(format, args...)
}

type WorkerLogWriter struct {
	logger    Logger
	processor processorFunc
}

func NewWorkerLogWriter(logger Logger, processor processorFunc) *WorkerLogWriter {
	return &WorkerLogWriter{logger: logger, processor: processor}
}

func (l *WorkerLogWriter) Write(p []byte) (n int, err error) {
	l.logger.Infof(l.processor(string(p)))

	return len(p), nil
}
