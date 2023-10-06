package p2pconfig

import (
	"path/filepath"
	"time"

	"github.com/libsv/bitcoin-hc/transports/p2p/p2putil"
)

// Default config for p2p app.
const (
	DefaultLogLevel                = "info"
	DefaultLogDirname              = "logs"
	DefaultLogFilename             = "p2p.log"
	DefaultMaxPeers                = 125
	DefaultMaxPeersPerIP           = 5
	DefaultBanDuration             = time.Hour * 24
	DefaultConnectTimeout          = time.Second * 30
	DefaultTrickleInterval         = 50 * time.Millisecond
	DefaultExcessiveBlockSize      = 128000000
	DefaultMinSyncPeerNetworkSpeed = 51200
	DefaultTargetOutboundPeers     = uint32(8)
	DefaultBlocksToConfirmFork     = 10
)

var (
	// DefaultHomeDir default app data dir for p2p.
	DefaultHomeDir = p2putil.AppDataDir("p2p", false)
	// Defaultp2pConfigPath default config path.
	Defaultp2pConfigPath = "config/config.json"
	// DefaultLogDir default directory for logs.
	DefaultLogDir = filepath.Join(DefaultHomeDir, DefaultLogDirname)
	// DefaultLogDir default directory for logs.
	DefaultConfigDir = filepath.Join(getWorkingDirectory(), Defaultp2pConfigPath)
)
