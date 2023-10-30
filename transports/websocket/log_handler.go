package websocket

import (
	"github.com/bitcoin-sv/pulse/domains/logging"
	"github.com/centrifugal/centrifuge"
)

type logHandler struct {
	logger logging.Logger
	level  logging.Level
}

func newLogHandler(lf logging.LoggerFactory) *logHandler {
	return &logHandler{
		logger: lf.NewLogger("centrifuge"),
		level:  lf.Level(),
	}
}

func (l *logHandler) Level() (level centrifuge.LogLevel) {
	switch l.level {
	case logging.Trace:
		level = centrifuge.LogLevelTrace
	case logging.Debug:
		level = centrifuge.LogLevelDebug
	case logging.Info:
		level = centrifuge.LogLevelInfo
	case logging.Warn:
		level = centrifuge.LogLevelWarn
	case logging.Error:
		level = centrifuge.LogLevelError
	case logging.Critical, logging.Off:
		level = centrifuge.LogLevelNone
	}
	return level
}

func (l *logHandler) Log(entry centrifuge.LogEntry) {
	var log func(format string, params ...interface{})
	switch entry.Level {
	case centrifuge.LogLevelTrace:
		log = l.logger.Tracef
	case centrifuge.LogLevelDebug:
		log = l.logger.Debugf
	case centrifuge.LogLevelInfo:
		log = l.logger.Infof
	case centrifuge.LogLevelWarn:
		log = l.logger.Warnf
	case centrifuge.LogLevelError:
		log = l.logger.Errorf
	case centrifuge.LogLevelNone:
		log = func(format string, params ...interface{}) {}
	}
	log("%s Context: %v.", entry.Message, entry.Fields)
}
