package logger

import (
	"github.com/libsv/bitcoin-hc/domains/logging"
	"github.com/libsv/bitcoin-hc/transports/p2p/p2plog"
)

type factory struct {
	system  string
	level   logging.Level
	backend *p2plog.Backend
}

// NewLoggerFactory creates a logger factory, that can be used to create loggers.
func NewLoggerFactory(systemTag string, level logging.Level, writer logWriterFunc) logging.LoggerFactory {
	return &factory{
		system:  systemTag,
		level:   level,
		backend: p2plog.NewBackend(writer),
	}
}

// Level current logging level of logger factory - with which level new logger will be created.
func (f *factory) Level() logging.Level {
	return f.level
}

// SetLevel set current logging level of logger factory - with which level new logger will be created.
func (f *factory) SetLevel(level logging.Level) {
	f.level = level
}

// NewLogger create new logger.
func (f *factory) NewLogger(name string) logging.Logger {
	tag := f.system
	if name != "" {
		tag = tag + "::" + name
	}
	l := f.backend.Logger(tag)
	a := newP2pLogAdapter(l)
	a.SetLevel(f.level)
	return a
}

type logWriterFunc func(p []byte) (n int, err error)

func (f logWriterFunc) Write(p []byte) (n int, err error) {
	return f(p)
}
