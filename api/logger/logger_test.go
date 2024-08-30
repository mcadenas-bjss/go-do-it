package logger_test

import (
	"bytes"
	"testing"

	"github.com/mcadenas-bjss/go-do-it/logger"
)

const (
	info  = "INFO"
	warn  = "WARN"
	error = "ERROR"
	debug = "DEBUG"
)

var levelMap = map[int]string{
	logger.Info:  info,
	logger.Warn:  warn,
	logger.Error: error,
	logger.Debug: debug,
}

func TestStandardLoggers(t *testing.T) {
	buffer := bytes.Buffer{}
	log := logger.NewLogger(&buffer)

	tests := []struct {
		name    string
		level   int
		message string
	}{
		{"Test Info", logger.Info, "This is an info message"},
		{"Test Warn", logger.Warn, "This is a warning message"},
		{"Test Error", logger.Error, "This is an error message"},
		{"Test Debug", logger.Debug, "This is a debug message"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.SetLevel(tt.level)
			switch {
			case tt.level == logger.Info:
				log.Info(tt.message)
			case tt.level == logger.Warn:
				log.Warn(tt.message)
			case tt.level == logger.Error:
				log.Error(tt.message)
			case tt.level == logger.Debug:
				log.Debug(tt.message)
			default:
				log.Info(tt.message)
			}
			got := buffer.String()

			assertLogMessage(t, got, tt.level, tt.message)
		})
	}
}

func TestFormattedLoggers(t *testing.T) {
	buffer := bytes.Buffer{}
	log := logger.NewLogger(&buffer)

	tests := []struct {
		name    string
		level   int
		message string
		args    []interface{}
	}{
		{"Test Info", logger.Info, "This is an %s message", []interface{}{"info"}},
		{"Test Warn", logger.Warn, "This is a %s message", []interface{}{"warning"}},
		{"Test Error", logger.Error, "This is an %s message", []interface{}{"error"}},
		{"Test Debug", logger.Debug, "This is a %s message", []interface{}{"debug"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.SetLevel(tt.level)
			switch {
			case tt.level == logger.Info:
				log.Infof(tt.message, tt.args...)
			case tt.level == logger.Warn:
				log.Warnf(tt.message, tt.args...)
			case tt.level == logger.Error:
				log.Errorf(tt.message, tt.args...)
			case tt.level == logger.Debug:
				log.Debugf(tt.message, tt.args...)
			default:
				log.Info(tt.message)
			}
			got := buffer.String()

			assertLogMessage(t, got, tt.level, tt.message)
		})
	}
}

func assertLogMessage(t testing.TB, actual string, level int, message string) {
	t.Helper()
	if !bytes.Contains([]byte(actual), []byte(levelMap[level])) {
		t.Errorf("Expected log message to contain %q, but got %q", levelMap[level], actual)
	}

	if !bytes.Contains([]byte(actual), []byte(message)) {
		t.Errorf("Expected log message to contain %q, but got %q", message, actual)
	}
}
