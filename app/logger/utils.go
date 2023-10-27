package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/libsv/bitcoin-hc/domains/logging"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
	"github.com/libsv/bitcoin-hc/internal/wire"
)

const (
	// maxRejectReasonLen is the maximum length of a sanitized reject reason
	// that will be logged.
	maxRejectReasonLen = 250
)

// DefaultLoggerFactory creates default factory with default system tag, level and writing to std out.
func DefaultLoggerFactory() logging.LoggerFactory {
	return NewLoggerFactory("HEADERS", logging.Info, os.Stdout)
}

// SetLevelFromString sets logger level based on string.
// Defaults to Info if string doesn't match expected level string.
func SetLevelFromString(target interface{}, level string) {
	l, _ := ParseLevel(level)
	SetLevel(target, l)
}

// SetLevel tries to set a logging level.
// If target is logging.CurrentLevelSetter then it is setting a logging level and returning true,
// otherwise returning false.
func SetLevel(target interface{}, l logging.Level) (ok bool) {
	t, ok := target.(logging.CurrentLevelSetter)
	if ok {
		t.SetLevel(l)
	}
	return
}

// ParseLevel returns a level based on the input string s.  If the input
// can't be interpreted as a valid log level, the info level and false is
// returned.
func ParseLevel(s string) (l logging.Level, ok bool) {
	switch strings.ToLower(s) {
	case "trace", "trc":
		return logging.Trace, true
	case "debug", "dbg":
		return logging.Debug, true
	case "info", "inf":
		return logging.Info, true
	case "warn", "wrn":
		return logging.Warn, true
	case "error", "err":
		return logging.Error, true
	case "critical", "crt":
		return logging.Critical, true
	case "off":
		return logging.Off, true
	default:
		return logging.Info, false
	}
}

// DirectionString returns string direction.
func DirectionString(inbound bool) string {
	if inbound {
		return "inbound"
	}
	return "outbound"
}

// LogClosure is a closure that can be printed with %v to be used to
// generate expensive-to-create data for a detailed log level and avoid doing
// the work if the data isn't printed.
type logClosure func() string

// String String() realization.
func (c logClosure) String() string {
	return c()
}

// NewLogClosure returns logClosure.
func NewLogClosure(c func() string) logClosure {
	return logClosure(c)
}

// MessageSummary returns a human-readable string which summarizes a message.
// Not all messages have or need a summary.  This is used for debug logging.
func MessageSummary(msg wire.Message) string {
	switch msg := msg.(type) {
	case *wire.MsgVersion:
		return fmt.Sprintf("agent %s, pver %d, block %d",
			msg.UserAgent, msg.ProtocolVersion, msg.LastBlock)

	case *wire.MsgVerAck:
		// No summary.

	case *wire.MsgGetAddr:
		// No summary.

	case *wire.MsgAddr:
		return fmt.Sprintf("%d addr", len(msg.AddrList))

	case *wire.MsgPing:
		// No summary - perhaps add nonce.

	case *wire.MsgPong:
		// No summary - perhaps add nonce.

	case *wire.MsgMemPool:
		// No summary.

	case *wire.MsgBlock:
		header := &msg.Header
		return fmt.Sprintf("hash %s, ver %d, %d tx, %s", msg.BlockHash(),
			header.Version, len(msg.Transactions), header.Timestamp)

	case *wire.MsgInv:
		return invSummary(msg.InvList)

	case *wire.MsgNotFound:
		return invSummary(msg.InvList)

	case *wire.MsgGetData:
		return invSummary(msg.InvList)

	case *wire.MsgGetBlocks:
		return locatorSummary(msg.BlockLocatorHashes, &msg.HashStop)

	case *wire.MsgGetHeaders:
		return locatorSummary(msg.BlockLocatorHashes, &msg.HashStop)

	case *wire.MsgHeaders:
		return fmt.Sprintf("num %d", len(msg.Headers))

	case *wire.MsgGetCFHeaders:
		return fmt.Sprintf("start_height=%d, stop_hash=%v",
			msg.StartHeight, msg.StopHash)

	case *wire.MsgCFHeaders:
		return fmt.Sprintf("stop_hash=%v, num_filter_hashes=%d",
			msg.StopHash, len(msg.FilterHashes))

	case *wire.MsgReject:
		// Ensure the variable length strings don't contain any
		// characters which are even remotely dangerous such as HTML
		// control characters, etc.  Also limit them to sane length for
		// logging.
		rejCommand := sanitizeString(msg.Cmd, wire.CommandSize)
		rejReason := sanitizeString(msg.Reason, maxRejectReasonLen)
		summary := fmt.Sprintf("cmd %v, code %v, reason %v", rejCommand,
			msg.Code, rejReason)
		if rejCommand == wire.CmdBlock || rejCommand == wire.CmdTx {
			summary += fmt.Sprintf(", hash %v", msg.Hash)
		}
		return summary
	}

	// No summary for other messages.
	return ""
}

// invSummary returns an inventory message as a human-readable string.
func invSummary(invList []*wire.InvVect) string {
	// No inventory.
	invLen := len(invList)
	if invLen == 0 {
		return "empty"
	}

	// One inventory item.
	if invLen == 1 {
		iv := invList[0]
		switch iv.Type {
		case wire.InvTypeError:
			return fmt.Sprintf("error %s", iv.Hash)
		case wire.InvTypeBlock:
			return fmt.Sprintf("block %s", iv.Hash)
		case wire.InvTypeTx:
			return fmt.Sprintf("tx %s", iv.Hash)
		case wire.InvTypeFilteredBlock:
			return fmt.Sprintf("filteredBlocks %s", iv.Hash)
		default:
			return fmt.Sprintf("unknown (%d) %s", uint32(iv.Type), iv.Hash)
		}
	}

	// More than one inv item.
	return fmt.Sprintf("size %d", invLen)
}

// locatorSummary returns a block locator as a human-readable string.
func locatorSummary(locator []*chainhash.Hash, stopHash *chainhash.Hash) string {
	if len(locator) > 0 {
		return fmt.Sprintf("locator %s, stop %s", locator[0], stopHash)
	}

	return fmt.Sprintf("no locator, stop %s", stopHash)

}

// sanitizeString strips any characters which are even remotely dangerous, such
// as html control characters, from the passed string.  It also limits it to
// the passed maximum size, which can be 0 for unlimited.  When the string is
// limited, it will also add "..." to the string to indicate it was truncated.
func sanitizeString(str string, maxLength uint) string {
	const safeChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY" +
		"Z01234567890 .,;_/:?@"

	// Strip any characters not in the safeChars string removed.
	str = strings.Map(func(r rune) rune {
		if strings.ContainsRune(safeChars, r) {
			return r
		}
		return -1
	}, str)

	// Limit the string to the max allowed length.
	if maxLength > 0 && uint(len(str)) > maxLength {
		str = str[:maxLength]
		str = str + "..."
	}
	return str
}
