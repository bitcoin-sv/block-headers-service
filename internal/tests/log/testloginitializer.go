package testlog

import (
	"encoding/json"
	"io"
	"os"

	"github.com/bitcoin-sv/pulse/app/logger"
	"github.com/bitcoin-sv/pulse/domains/logging"
)

// NewTestLoggerFactory creates new logger factory for tests.
func NewTestLoggerFactory() logging.LoggerFactory {
	return logger.NewLoggerFactory("PULSE_TEST", logging.Debug, os.Stdout)
}

// NewTestLoggerFactoryWithRecorder creates new logger factory for tests and log recorder to check logged messages.
func NewTestLoggerFactoryWithRecorder() (logging.LoggerFactory, *LogRecorder) {
	recorder := newLogRecorder()
	return logger.NewLoggerFactory("PULSE_TEST", logging.Debug, recorder), recorder
}

// LogRecorder helper that is recording log messages send to logger created with NewTestLoggerFactoryWithRecorder.
type LogRecorder struct {
	Logs []LogMessage
	io.Writer
}

func newLogRecorder() *LogRecorder {
	return &LogRecorder{
		Logs: make([]LogMessage, 0),
	}
}

func (r *LogRecorder) record(p []byte) error {
	lm := LogMessage{}
	err := json.Unmarshal(p, &lm)
	if err != nil {
		return err
	}
	r.Logs = append(r.Logs, lm)
	return nil
}

// Clear clears recorded logs.
func (r *LogRecorder) Clear() {
	r.Logs = make([]LogMessage, 0)
}

// Last recorded raw log entry.
func (r *LogRecorder) Last() LogMessage {
	return r.Logs[len(r.Logs)-1]
}

// LastUnmarshaled last recorded log unmarshaled.
func (r *LogRecorder) LastUnmarshaled() LogMessage {
	return r.Last()
}

func (r *LogRecorder) Write(p []byte) (int, error) {
	err := r.record(p)
	if err != nil {
		return 0, err
	}
	return os.Stdout.Write(p)
}

// LogMessage represents log message.
type LogMessage struct {
	Level       string `json:"log.level"`
	Message     string `json:"message"`
	Application string `json:"application"`
	Module      string `json:"module"`
}
