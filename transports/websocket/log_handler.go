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
	var event *zerolog.Event
	switch entry.Level {
	case centrifuge.LogLevelTrace:
		event = l.logger.Trace()
	case centrifuge.LogLevelDebug:
		event = l.logger.Debug()
	case centrifuge.LogLevelInfo:
		event = l.logger.Info()
	case centrifuge.LogLevelWarn:
		event = l.logger.Warn()
	case centrifuge.LogLevelError:
		event = l.logger.Error()
	case centrifuge.LogLevelNone:
		event = nil
	}
	if event != nil {
		event.Msgf("%s Context: %v.", entry.Message, entry.Fields)
	}
}
