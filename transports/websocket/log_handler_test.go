package websocket

import (
	"testing"

	"github.com/centrifugal/centrifuge"
	"github.com/libsv/bitcoin-hc/domains/logging"
	"github.com/libsv/bitcoin-hc/internal/tests/assert"
	testlog "github.com/libsv/bitcoin-hc/internal/tests/log"
	"github.com/rs/zerolog"
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
	lf.SetLevel(logging.Trace)
	for name, params := range map[string]struct {
		level    centrifuge.LogLevel
		expected testlog.LogMessage
	}{
		"trace": {
			level:    centrifuge.LogLevelTrace,
			expected: testlog.LogMessage{
				Level: zerolog.LevelTraceValue,
				Message: "Example message Context: map[client:123456].",
				Application: "PULSE_TEST",
				Module: "centrifuge",
			},
		},
		"debug": {
			level:    centrifuge.LogLevelDebug,
			expected: testlog.LogMessage{
				Level: zerolog.LevelDebugValue,
				Message: "Example message Context: map[client:123456].",
				Application: "PULSE_TEST",
				Module: "centrifuge",
			},
		},
		"info": {
			level:    centrifuge.LogLevelInfo,
			expected: testlog.LogMessage{
				Level: zerolog.LevelInfoValue,
				Message: "Example message Context: map[client:123456].",
				Application: "PULSE_TEST",
				Module: "centrifuge",
			},
		},
		"warn": {
			level:    centrifuge.LogLevelWarn,
			expected: testlog.LogMessage{
				Level: zerolog.LevelWarnValue,
				Message: "Example message Context: map[client:123456].",
				Application: "PULSE_TEST",
				Module: "centrifuge",
			},
		},
		"error": {
			level:    centrifuge.LogLevelError,
			expected: testlog.LogMessage{
				Level: zerolog.LevelErrorValue,
				Message: "Example message Context: map[client:123456].",
				Application: "PULSE_TEST",
				Module: "centrifuge",
			},
		},
	} {
		title := "handle log of " + name + " level"
		t.Run(title, func(t *testing.T) {
			//given
			
			handler := newLogHandler(lf)

			//when
			handler.Log(newLogEntry(params.level))
			//and
			actual := recorder.LastUnmarshaled()

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
