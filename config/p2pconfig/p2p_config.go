package p2pconfig

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/bitcoin-sv/pulse/internal/chaincfg"
)

// Config struct for p2pconfig.
type Config struct {
	LogDir                    string        `mapstructure:"logdir" description:"Directory to log output."`
	MaxPeers                  int           `mapstructure:"maxpeers" description:"Max number of inbound and outbound peers"`
	MaxPeersPerIP             int           `mapstructure:"maxpeersperip" description:"Max number of inbound and outbound peers per IP"`
	BanDuration               time.Duration `mapstructure:"banduration" description:"How long to ban misbehaving peers.  Valid time units are {s, m, h}.  Minimum 1 second"`
	MinSyncPeerNetworkSpeed   uint64        `mapstructure:"minsyncpeernetworkspeed" description:"Disconnect sync peers slower than this threshold in bytes/sec"`
	AddCheckpoints            []string      `mapstructure:"addcheckpoint" description:"Add a custom checkpoint.  Format: '<height>:<hash>'"`
	DisableCheckpoints        bool          `mapstructure:"nocheckpoints" description:"Disable built-in checkpoints.  Don't do this unless you know what you're doing."`
	LogLevel                  string        `mapstructure:"loglevel" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical} -- You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set the log level for individual subsystems -- Use show to list available subsystems"`
	ExcessiveBlockSize        uint32        `mapstructure:"excessiveblocksize" description:"The maximum size block (in bytes) this node will accept. Cannot be less than 32000000."`
	TrickleInterval           time.Duration `mapstructure:"trickleinterval" description:"Minimum time between attempts to send new inventory to a connected peer"`
	TargetOutboundPeers       uint32        `mapstructure:"targetoutboundpeers" description:"number of outbound connections to maintain"`
	BlocksForForkConfirmation int           `mapstructure:"blocksforconfirmation" description:"Minimum number of blocks to consider a block confirmed"`
	lookup                    func(string) ([]net.IP, error)
	dial                      func(string, string, time.Duration) (net.Conn, error)
	AddCheckpointsSlice       []chaincfg.Checkpoint
	Checkpoints               []chaincfg.Checkpoint
	TimeSource                MedianTimeSource
}

// Validate validates config and sets some parameters based on the config provided.
func (c *Config) Validate() (err error) {
	funcName := "validateConfig"

	// Append the network type to the log directory so it is "namespaced"
	// per network in the same fashion as the data directory.
	c.LogDir = cleanAndExpandPath(c.LogDir)
	c.LogDir = filepath.Join(c.LogDir, netName(ActiveNetParams))

	// Don't allow ban durations that are too short.
	if c.BanDuration < time.Second {
		str := "%s: The banduration option may not be less than 1s -- parsed [%v]"
		err := fmt.Errorf(str, funcName, c.BanDuration)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	// Excessive blocksize cannot be set less than the default but it can be higher.
	c.ExcessiveBlockSize = maxUint32(c.ExcessiveBlockSize, DefaultExcessiveBlockSize)

	c.dial = net.DialTimeout
	c.lookup = net.LookupIP

	c.AddCheckpointsSlice, err = parseCheckpoints(c.AddCheckpoints)
	if err != nil {
		str := "%s: Error parsing checkpoints: %v"
		err := fmt.Errorf(str, funcName, err)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	if !c.DisableCheckpoints {
		c.Checkpoints = mergeCheckpoints(ActiveNetParams.Checkpoints, c.AddCheckpointsSlice)
	}

	return nil
}
