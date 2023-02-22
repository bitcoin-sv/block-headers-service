package bitcoin

import (
	"fmt"
	"log"
)

type Logger interface {
	// Debugf logs a message at debug level.
	Debugf(format string, args ...interface{})
	// Infof logs a message at info level.
	Infof(format string, args ...interface{})
	// Warnf logs a message at warn level.
	Warnf(format string, args ...interface{})
	// Errorf logs a message at error level.
	Errorf(format string, args ...interface{})
	// Fatalf logs a message at fatal level.
	Fatalf(format string, args ...interface{})
}

type DefaultLogger struct{}

func (l *DefaultLogger) Debugf(format string, args ...interface{}) {
	f := fmt.Sprintf("DEBUG: %s", format)
	log.Printf(f, args...)
}

func (l *DefaultLogger) Infof(format string, args ...interface{}) {
	f := fmt.Sprintf("INFO: %s", format)
	log.Printf(f, args...)
}

func (l *DefaultLogger) Warnf(format string, args ...interface{}) {
	f := fmt.Sprintf("wARN: %s", format)
	log.Printf(f, args...)
}

func (l *DefaultLogger) Errorf(format string, args ...interface{}) {
	f := fmt.Sprintf("ERROR: %s", format)
	log.Printf(f, args...)
}

func (l *DefaultLogger) Fatalf(format string, args ...interface{}) {
	f := fmt.Sprintf("FATAL: %s", format)
	log.Printf(f, args...)
}
