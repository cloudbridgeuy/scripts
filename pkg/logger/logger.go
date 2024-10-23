package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

var Logger *log.Logger

func init() {
	Logger = log.New(os.Stderr)
	Logger.SetLevel(log.WarnLevel)
}

// Verbose enables verbose mode for the logger
func Verbose() {
	Logger.SetLevel(log.DebugLevel)
}

// Debug logs a debug message.
func Debug(msg interface{}, keyvals ...interface{}) {
	Logger.Debug(msg, keyvals)
}

// Debugf logs a formatted info message.
func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

// Info logs an info message.
func Info(msg interface{}, keyvals ...interface{}) {
	Logger.Info(msg, keyvals)
}

// Infof logs a formatted info message.
func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

// Error logs an error message.
func Error(msg interface{}, keyvals ...interface{}) {
	Logger.Error(msg, keyvals)
}

// Errorf logs an error message.
func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}
