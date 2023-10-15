package logger

import (
	"io"

	"github.com/libsv/bitcoin-hc/domains/logging"
	"github.com/rs/zerolog"
	"go.elastic.co/ecszerolog"
)

type factory struct {
	logger zerolog.Logger
}

// NewLoggerFactory creates a logger factory, that can be used to create loggers.
func NewLoggerFactory(systemTag string, level logging.Level, writer io.Writer) logging.LoggerFactory {
	logLevel := ecszerolog.Level(toZerologLevel(level))

	logger := ecszerolog.New(writer, logLevel).
		Level(zerolog.TraceLevel).
		With().
		Str("application", systemTag).
		Logger()

	return &factory{
		logger: logger,
	}
}

// Level current logging level of logger factory - with which level new logger will be created.
func (f *factory) Level() logging.Level {
	return toLoggingLevel(f.logger.GetLevel())
}

// SetLevel set current logging level of logger factory - with which level new logger will be created.
func (f *factory) SetLevel(level logging.Level) {
	zeroLogLevel := toZerologLevel(level)
	f.logger = f.logger.Level(zeroLogLevel).With().Logger()
}

// NewLogger create new logger.
func (f *factory) NewLogger(name string) logging.Logger {
	l := f.logger.With().Str("module", name).Logger()
	a := newZerologAdapter(l)
	return a
}

func toZerologLevel(level logging.Level) zerolog.Level {
	switch level {
	case logging.Trace:
		return zerolog.TraceLevel
	case logging.Debug:
		return zerolog.DebugLevel
	case logging.Info:
		return zerolog.InfoLevel
	case logging.Warn:
		return zerolog.WarnLevel
	case logging.Error:
		return zerolog.ErrorLevel
	case logging.Critical:
		return zerolog.FatalLevel
	case logging.Off:
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}

func toLoggingLevel(level zerolog.Level) logging.Level {
	switch level { //nolint:exhaustive
	case zerolog.TraceLevel:
		return logging.Trace
	case zerolog.DebugLevel:
		return logging.Debug
	case zerolog.InfoLevel :
		return logging.Info
	case zerolog.WarnLevel:
		return logging.Warn
	case zerolog.ErrorLevel:
		return logging.Error
	case zerolog.FatalLevel:
		return logging.Critical
	case zerolog.Disabled:
		return logging.Off
	default:
		return logging.Info
	}
}
