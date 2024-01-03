package websocket

import (
	"github.com/bitcoin-sv/pulse/logging"
	"testing"

	"github.com/bitcoin-sv/pulse/internal/tests/assert"
	"github.com/centrifugal/centrifuge"
	"github.com/rs/zerolog"
)

func TestReturningExpectedLoggingLevel(t *testing.T) {
	//setup

	for name, params := range map[string]struct {
		appLoggingLevel      zerolog.Level
		expectedLoggingLevel centrifuge.LogLevel
	}{
		"off": {
			appLoggingLevel:      zerolog.NoLevel,
			expectedLoggingLevel: centrifuge.LogLevelNone,
		},
		"trace": {
			appLoggingLevel:      zerolog.TraceLevel,
			expectedLoggingLevel: centrifuge.LogLevelTrace,
		},
		"debug": {
			appLoggingLevel:      zerolog.DebugLevel,
			expectedLoggingLevel: centrifuge.LogLevelDebug,
		},
		"info": {
			appLoggingLevel:      zerolog.InfoLevel,
			expectedLoggingLevel: centrifuge.LogLevelInfo,
		},
		"warn": {
			appLoggingLevel:      zerolog.WarnLevel,
			expectedLoggingLevel: centrifuge.LogLevelWarn,
		},
		"error": {
			appLoggingLevel:      zerolog.ErrorLevel,
			expectedLoggingLevel: centrifuge.LogLevelError,
		},
		"fatal": {
			appLoggingLevel:      zerolog.FatalLevel,
			expectedLoggingLevel: centrifuge.LogLevelNone,
		},
	} {
		title := "handle " + name + " level"
		t.Run(title, func(t *testing.T) {
			//given
			testLogger, err := logging.CreateLogger("test-logger", "console", params.appLoggingLevel.String(), false)
			assert.NoError(t, err)
			handler := newLogHandler(testLogger)

			//when
			actual := handler.Level()

			//then
			assert.Equal(t, actual, params.expectedLoggingLevel)
		})
	}
}
