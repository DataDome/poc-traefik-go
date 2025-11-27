package modulego

import (
	"fmt"
	"log"
	"os"
)

// Logger is an interface that defines the methods for logging at various levels of severity.
// Implementations of Logger can be used to handle logging output for the package.
//
// Methods:
//   - Debug: Logs debug-level messages, used for detailed troubleshooting information.
//   - Info: Logs informational messages.
//   - Warn: Logs warning messages about non-critical issues.
//   - Error: Logs error messages for issues that might affect the protection.
//
// If none provider are defined, a default logger is provided by using the standard log package.
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

// defaultLogger implements the Logger interface by using the standard log package.
type defaultLogger struct {
	logger *log.Logger
}

// NewDefaultLogger returns a new default [Logger] instance.
// This logger relies on the standard log package.
func NewDefaultLogger() Logger {
	return &defaultLogger{
		logger: log.New(os.Stdout, "[DataDome] ", log.LstdFlags),
	}
}

// Debug method for the default logger
func (l *defaultLogger) Debug(args ...interface{}) {
	l.logger.Println("DEBUG:", fmt.Sprint(args...))
}

// Info method for the default logger
func (l *defaultLogger) Info(args ...interface{}) {
	l.logger.Println("INFO:", fmt.Sprint(args...))
}

// Warn method for the default logger
func (l *defaultLogger) Warn(args ...interface{}) {
	l.logger.Println("WARN:", fmt.Sprint(args...))
}

// Error method for the default logger
func (l *defaultLogger) Error(args ...interface{}) {
	l.logger.Println("ERROR:", fmt.Sprint(args...))
}
