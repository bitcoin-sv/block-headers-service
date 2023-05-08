package websocket

import (
	"testing"

	"github.com/centrifugal/centrifuge"
	"github.com/libsv/bitcoin-hc/domains/logging"
	"github.com/libsv/bitcoin-hc/internal/tests/assert"
	testlog "github.com/libsv/bitcoin-hc/internal/tests/log"
)

func TestReturningExpectedLoggingLevel(t *testing.T) {
	//setup
	lf := testlog.NewTestLoggerFactory()

	for name, params := range map[string]struct {
		appLoggingLevel      logging.Level
		expectedLoggingLevel centrifuge.LogLevel
	}{
		"off": {
			appLoggingLevel:      logging.Off,
			expectedLoggingLevel: centrifuge.LogLevelNone,
		},
		"trace": {
			appLoggingLevel:      logging.Trace,
			expectedLoggingLevel: centrifuge.LogLevelTrace,
		},
		"debug": {
			appLoggingLevel:      logging.Debug,
			expectedLoggingLevel: centrifuge.LogLevelDebug,
		},
		"info": {
			appLoggingLevel:      logging.Info,
			expectedLoggingLevel: centrifuge.LogLevelInfo,
		},
		"warn": {
			appLoggingLevel:      logging.Warn,
			expectedLoggingLevel: centrifuge.LogLevelWarn,
		},
		"error": {
			appLoggingLevel:      logging.Error,
			expectedLoggingLevel: centrifuge.LogLevelError,
		},
		"critical": {
			appLoggingLevel:      logging.Critical,
			expectedLoggingLevel: centrifuge.LogLevelNone,
		},
	} {
		title := "handle " + name + " level"
		t.Run(title, func(t *testing.T) {
			//given
			lf.SetLevel(params.appLoggingLevel)
			handler := newLogHandler(lf)

			//when
			actual := handler.Level()

			//then
			assert.Equal(t, actual, params.expectedLoggingLevel)
		})
	}
}

func TestLoggerHandler(t *testing.T) {
	//setup
	lf, recorder := testlog.NewTestLoggerFactoryWithRecorder()

	for name, params := range map[string]struct {
		level    centrifuge.LogLevel
		expected string
	}{
		"trace": {
			level:    centrifuge.LogLevelTrace,
			expected: "[TRC] PULSE_TEST::centrifuge: Example message Context: map[client:123456].",
		},
		"debug": {
			level:    centrifuge.LogLevelDebug,
			expected: "[DBG] PULSE_TEST::centrifuge: Example message Context: map[client:123456].",
		},
		"info": {
			level:    centrifuge.LogLevelInfo,
			expected: "[INF] PULSE_TEST::centrifuge: Example message Context: map[client:123456].",
		},
		"warn": {
			level:    centrifuge.LogLevelWarn,
			expected: "[WRN] PULSE_TEST::centrifuge: Example message Context: map[client:123456].",
		},
		"error": {
			level:    centrifuge.LogLevelError,
			expected: "[ERR] PULSE_TEST::centrifuge: Example message Context: map[client:123456].",
		},
	} {
		title := "handle log of " + name + " level"
		t.Run(title, func(t *testing.T) {
			//given
			lf.SetLevel(logging.Trace)
			handler := newLogHandler(lf)

			//when
			handler.Log(newLogEntry(params.level))
			//and
			actual := recorder.LastNormalized()

			//then
			assert.Equal(t, actual, params.expected)
		})
	}
}

func newLogEntry(level centrifuge.LogLevel) centrifuge.LogEntry {
	return centrifuge.LogEntry{
		Level:   level,
		Message: "Example message",
		Fields:  map[string]any{"client": "123456"},
	}
}
