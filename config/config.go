package config

import (
	"net"
	"time"

	"github.com/bitcoin-sv/pulse/domains/logging"
	"github.com/bitcoin-sv/pulse/internal/chaincfg"
)

const (
	ApplicationName       = "pulse"
	APIVersion            = "v1"
	Version               = "v0.6.0"
	ConfigFilePathKey     = "config_file"
	DefaultConfigFilePath = "config.yaml"
	ConfigEnvPrefix       = "pulse_"
)

// DbType database type.
type DbType string

// AppConfig returns strongly typed config values.
type AppConfig struct {
	ConfigFile       string            `mapstructure:"configFile"`
	DbConfig         *DbConfig         `mapstructure:"db"`
	P2PConfig        *P2PConfig        `mapstructure:"p2p"`
	MerkleRootConfig *MerkleRootConfig `mapstructure:"merkleroot"`
	WebhookConfig    *WebhookConfig    `mapstructure:"webhook"`
	WebsocketConfig  *WebsocketConfig  `mapstructure:"websocket"`
	HTTPConfig       *HTTPConfig       `mapstructure:"http"`
	LoggerFactory    logging.LoggerFactory
}

// DbConfig represents a database connection.
type DbConfig struct {
	Type               DbType `mapstructure:"type"`
	SchemaPath         string `mapstructure:"schemaPath"`
	Dsn                string `mapstructure:"dsn"`
	FilePath           string `mapstructure:"dbFilePath"`
	PreparedDb         bool   `mapstructure:"preparedDb"`
	PreparedDbFilePath string `mapstructure:"preparedDbFilePath"`
}

// MerkleRootConfig represents merkleroots verification config.
type MerkleRootConfig struct {
	MaxBlockHeightExcess int `mapstructure:"maxBlockHeightExcess"`
}

// WebhookConfig represents a webhook config.
type WebhookConfig struct {
	MaxTries int `mapstructure:"maxTries"`
}

// WebsocketConfig represents a websocket config.
type WebsocketConfig struct {
	HistoryMax int `mapstructure:"historyMax"`
	HistoryTTL int `mapstructure:"historyTTL"`
}

// HTTPConfig represents a HTTPConfig config.
type HTTPConfig struct {
	ReadTimeout  int    `mapstructure:"readTimeout"`
	WriteTimeout int    `mapstructure:"writeTimeout"`
	Port         int    `mapstructure:"port"`
	UrlPrefix    string `mapstructure:"urlPrefix"`
	UseAuth      bool   `mapstructure:"useAuth"`
	AuthToken    string `mapstructure:"authToken"`
}

// P2PConfig represents a p2p config.
type P2PConfig struct {
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
	BlocksForForkConfirmation int           `mapstructure:"blocksforconfirmation" description:"Minimum number of blocks to consider a block confirmed"`
	lookup                    func(string) ([]net.IP, error)
	dial                      func(string, string, time.Duration) (net.Conn, error)
	Checkpoints               []chaincfg.Checkpoint
	TimeSource                MedianTimeSource
}
