package logging

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"go.elastic.co/ecszerolog"
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
			Str("application", instanceName).
			Logger()
	} else {
		logger = ecszerolog.New(writer, logLevel).
			With().
			Str("application", instanceName).
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
		Caller().
		Str("application", "pulse-default").
		Logger()

	return &logger
}

// SampledLogger returns a logger that samples messages with the rate level.
func SampledLogger(baseLogger *zerolog.Logger, rate uint32) *zerolog.Logger {
	logger := baseLogger.Sample(&zerolog.BasicSampler{N: rate}).With().Uint32("sample.rate", rate).Logger()
	return &logger
}
