package logger

import (
	"github.com/libsv/bitcoin-hc/domains/logging"
	"github.com/libsv/bitcoin-hc/transports/p2p/p2plog"
)

type adapter struct {
	logging.Logger
}

func newP2pLogAdapter(logger p2plog.Logger) *adapter {
	return &adapter{Logger: logger}
}

func (a *adapter) SetLevel(level logging.Level) {
	l := a.Logger.(p2plog.Logger)
	l.SetLevel(p2plog.Level(level))
}

func (a *adapter) Level() logging.Level {
	l := a.Logger.(p2plog.Logger)
	level := l.Level()
	return ToLevel(level)
}
