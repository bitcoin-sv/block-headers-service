package testlog

import (
	"os"

	"github.com/libsv/bitcoin-hc/app/logger"
	"github.com/libsv/bitcoin-hc/domains/logging"
	"github.com/libsv/bitcoin-hc/transports/p2p/p2plog"
)

// InitializeMockLogger initialize logger for tests.
func InitializeMockLogger() p2plog.Logger {
	l := NewTestLoggerFactory().NewLogger("")
	log, ok := logger.UnwrapP2plog(l)
	if !ok {
		panic("expect to unwrap P2plog from logger")
	}
	return log
}

// NewTestLoggerFactory creates new logger factory for tests.
func NewTestLoggerFactory() logging.LoggerFactory {
	return logger.NewLoggerFactory("PULSE_TEST", logging.Debug, os.Stdout.Write)
}

// NewTestLoggerFactoryWithRecorder creates new logger factory for tests and log recorder to check logged messages.
func NewTestLoggerFactoryWithRecorder() (logging.LoggerFactory, *LogRecorder) {
	recorder := newLogRecorder()
	return logger.NewLoggerFactory("PULSE_TEST", logging.Debug, func(p []byte) (n int, err error) {
		recorder.record(p)
		return os.Stdout.Write(p)
	}), recorder
}

// LogRecorder helper that is recording log messages send to logger created with NewTestLoggerFactoryWithRecorder.
type LogRecorder struct {
	Logs []string
}

func newLogRecorder() *LogRecorder {
	return &LogRecorder{
		Logs: make([]string, 0),
	}
}

func (r *LogRecorder) record(p []byte) {
	r.Logs = append(r.Logs, string(p))
}

// Clear clears recorded logs.
func (r *LogRecorder) Clear() {
	r.Logs = make([]string, 0)
}

// Last recorded raw log entry.
func (r *LogRecorder) Last() string {
	return r.Logs[len(r.Logs)-1]
}

// LastNormalized last recorded log normalized: with stripped date time at the start and new line character at the end.
func (r *LogRecorder) LastNormalized() string {
	return normalize(r.Last())
}

func normalize(msg string) string {
	return msg[24 : len(msg)-1]
}
