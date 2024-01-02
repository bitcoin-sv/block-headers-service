package logging

import (
	"github.com/rs/zerolog"
	"go.elastic.co/ecszerolog"
	"io"
	"os"
	"time"
)

const (
	consoleLogFormat = "console"
	jsonLogFormat    = "json"
)

// CreateLogger create and configure zerolog logger based on app config.
func CreateLogger(instanceName, format, level string, logOrigin bool) (*zerolog.Logger, error) {
	var writer io.Writer
	if format == consoleLogFormat {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05.000",
		}
	} else {
		writer = os.Stdout
	}

	parsedLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	logLevel := ecszerolog.Level(parsedLevel)
	origin := ecszerolog.Origin()
	var logger zerolog.Logger

	if logOrigin {
		logger = ecszerolog.New(writer, logLevel, origin).
			With().
			Timestamp().
			Str("application", instanceName).
			Str("service", "bux-server").
			Logger()
	} else {
		logger = ecszerolog.New(writer, logLevel).
			With().
			Timestamp().
			Str("application", instanceName).
			Str("service", "bux-server").
			Logger()
	}

	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(time.Local) //nolint:gosmopolitan // We want local time inside logger.
	}

	return &logger, nil
}

// GetDefaultLogger create and configure default zerolog logger. It should be used before config is loaded.
func GetDefaultLogger() *zerolog.Logger {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05.000",
	}
	logger := ecszerolog.New(writer, ecszerolog.Level(zerolog.DebugLevel)).
		With().
		Timestamp().
		Caller().
		Str("application", "pulse-default").
		Logger()

	return &logger
}
