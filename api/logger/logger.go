package logger

import (
	"io"
	"log"
	"os"
)

const (
	Info = 1 << iota
	Warn
	Error
	Debug
)

type Logger interface {
	Info(v ...interface{})
	Debug(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})

	SetLevel(level int)
}

type ConsoleLogger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
	level int
}

func NewLogger(writer io.Writer) *ConsoleLogger {
	var stdout io.Writer
	var stderr io.Writer
	if writer == nil {
		stdout = os.Stdout
		stderr = os.Stderr
	} else {
		stdout = writer
		stderr = writer
	}
	return &ConsoleLogger{
		debug: log.New(stdout, "DEBUG: ", log.LstdFlags),
		info:  log.New(stdout, "INFO: ", log.LstdFlags),
		warn:  log.New(stdout, "WARN: ", log.LstdFlags),
		error: log.New(stderr, "ERROR: ", log.LstdFlags),
		level: Info,
	}
}

func (l *ConsoleLogger) SetLevel(level int) {
	l.level = level
}

func (l *ConsoleLogger) Info(v ...interface{}) {
	if l.level >= Info {
		l.info.Println(v...)
	}
}
func (l *ConsoleLogger) Infof(format string, v ...interface{}) {
	if l.level >= Info {
		l.info.Printf(format, v...)
	}
}

func (l *ConsoleLogger) Warn(v ...interface{}) {
	if l.level >= Warn {
		l.warn.Println(v...)
	}
}
func (l *ConsoleLogger) Warnf(format string, v ...interface{}) {
	if l.level >= Warn {
		l.warn.Printf(format, v...)
	}
}

func (l *ConsoleLogger) Error(v ...interface{}) {
	if l.level >= Error {
		l.error.Println(v...)
	}
}
func (l *ConsoleLogger) Errorf(format string, v ...interface{}) {
	if l.level >= Error {
		l.error.Printf(format, v...)
	}
}

func (l *ConsoleLogger) Debug(v ...interface{}) {
	if l.level >= Debug {
		l.debug.Println(v...)
	}
}
func (l *ConsoleLogger) Debugf(format string, v ...interface{}) {
	if l.level >= Debug {
		l.debug.Printf(format, v...)
	}
}

func (l *ConsoleLogger) Fatal(v ...any) {
	l.error.Println(v...)
	os.Exit(1)
}
