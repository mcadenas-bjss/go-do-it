package logging

import (
	"fmt"
	"time"
)

type Logger struct {
	Level int
}

type Log struct {
	Time string
	msg  string
	Args []interface{}
}

// func NewLogger() *slog.Logger {
// 	jsonHandler := slog.NewJSONHandler(os.Stderr, nil)
// 	return slog.New(jsonHandler)
// }

func NewLogger(level int) *Logger {
	return &Logger{
		Level: level,
	}
}

func newLog(msg string, args ...interface{}) Log {
	return Log{
		Time: time.Now().Format("2006-01-02 15:04:05"),
		msg:  msg,
		Args: args,
	}
}

func (l *Logger) SetLevel(level int) {
	l.Level = level
}

func (l *Logger) Info(msg string, args ...interface{}) {
	lg := newLog(msg, args...)
	if l.Level >= 0 {
		printLog(lg)
	}
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	lg := newLog(msg, args...)
	if l.Level >= 1 {
		printLog(lg)
	}
}

func (l *Logger) Error(msg string, args ...interface{}) {
	lg := newLog(msg, args...)
	if l.Level >= 2 {
		printLog(lg)
	}
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	lg := newLog(msg, args...)
	if l.Level >= 3 {
		printLog(lg)
	}
}

func printLog(lg Log) {
	if len(lg.Args) > 0 {
		fmt.Printf("%s %s : %#v\n", lg.Time, lg.msg, lg.Args)
	} else {
		fmt.Printf("%s %s\n", lg.Time, lg.msg)
	}
}
