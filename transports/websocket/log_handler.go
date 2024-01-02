package websocket

import (
	"github.com/centrifugal/centrifuge"
	"github.com/rs/zerolog"
)

type logHandler struct {
	logger *zerolog.Logger
	level  zerolog.Level
}

func newLogHandler(log *zerolog.Logger) *logHandler {
	websocketLogger := log.With().Str("subservice", "centrifuge").Logger()
	return &logHandler{
		logger: &websocketLogger,
		level:  websocketLogger.GetLevel(),
	}
}

func (l *logHandler) Level() (level centrifuge.LogLevel) {
	switch l.level {
	case zerolog.TraceLevel:
		level = centrifuge.LogLevelTrace
	case zerolog.DebugLevel:
		level = centrifuge.LogLevelDebug
	case zerolog.InfoLevel:
		level = centrifuge.LogLevelInfo
	case zerolog.WarnLevel:
		level = centrifuge.LogLevelWarn
	case zerolog.ErrorLevel:
		level = centrifuge.LogLevelError
	case zerolog.FatalLevel, zerolog.NoLevel, zerolog.PanicLevel, zerolog.Disabled:
		level = centrifuge.LogLevelNone
	}
	return level
}

func (l *logHandler) Log(entry centrifuge.LogEntry) {
	var log func(format string, params ...interface{})
	switch entry.Level {
	case centrifuge.LogLevelTrace:
		log = l.logger.Trace().Msgf
	case centrifuge.LogLevelDebug:
		log = l.logger.Debug().Msgf
	case centrifuge.LogLevelInfo:
		log = l.logger.Info().Msgf
	case centrifuge.LogLevelWarn:
		log = l.logger.Warn().Msgf
	case centrifuge.LogLevelError:
		log = l.logger.Error().Msgf
	case centrifuge.LogLevelNone:
		log = func(format string, params ...interface{}) {}
	}
	log("%s Context: %v.", entry.Message, entry.Fields)
}
