package config

import (
	"net"
	"time"

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
	DbConfig         *DbConfig         `mapstructure:"db"`
	P2PConfig        *P2PConfig        `mapstructure:"p2p"`
	MerkleRootConfig *MerkleRootConfig `mapstructure:"merkleroot"`
	WebhookConfig    *WebhookConfig    `mapstructure:"webhook"`
	WebsocketConfig  *WebsocketConfig  `mapstructure:"websocket"`
	HTTPConfig       *HTTPConfig       `mapstructure:"http"`
}

// DbConfig represents a database connection.
type DbConfig struct {
	Type               DbType `mapstructure:"type"`
	SchemaPath         string `mapstructure:"schema_path"`
	Dsn                string `mapstructure:"dsn"`
	FilePath           string `mapstructure:"db_file_path"`
	PreparedDb         bool   `mapstructure:"prepared_db"`
	PreparedDbFilePath string `mapstructure:"prepared_db_file_path"`
}

// MerkleRootConfig represents merkleroots verification config.
type MerkleRootConfig struct {
	MaxBlockHeightExcess int `mapstructure:"max_block_height_excess"`
}

// WebhookConfig represents a webhook config.
type WebhookConfig struct {
	MaxTries int `mapstructure:"max_tries"`
}

// WebsocketConfig represents a websocket config.
type WebsocketConfig struct {
	HistoryMax int `mapstructure:"history_max"`
	HistoryTTL int `mapstructure:"history_ttl"`
}

// HTTPConfig represents a HTTPConfig config.
type HTTPConfig struct {
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	Port         int    `mapstructure:"port"`
	UrlPrefix    string `mapstructure:"url_prefix"`
	UseAuth      bool   `mapstructure:"use_auth"`
	AuthToken    string `mapstructure:"auth_token"`
}

// P2PConfig represents a p2p config.
type P2PConfig struct {
	MaxPeers                  int           `mapstructure:"max_peers" description:"Max number of inbound and outbound peers"`
	MaxPeersPerIP             int           `mapstructure:"max_peers_per_ip" description:"Max number of inbound and outbound peers per IP"`
	BanDuration               time.Duration `mapstructure:"ban_duration" description:"How long to ban misbehaving peers.  Valid time units are {s, m, h}.  Minimum 1 second"`
	MinSyncPeerNetworkSpeed   uint64        `mapstructure:"min_sync_peer_network_speed" description:"Disconnect sync peers slower than this threshold in bytes/sec"`
	DisableCheckpoints        bool          `mapstructure:"disable_checkpoints" description:"Disable built-in checkpoints.  Don't do this unless you know what you're doing."`
	LogLevel                  string        `mapstructure:"log_level" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical} -- You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set the log level for individual subsystems -- Use show to list available subsystems"`
	ExcessiveBlockSize        uint32        `mapstructure:"excessive_block_size" description:"The maximum size block (in bytes) this node will accept. Cannot be less than 32000000."`
	TrickleInterval           time.Duration `mapstructure:"trickle_interval" description:"Minimum time between attempts to send new inventory to a connected peer"`
	BlocksForForkConfirmation int           `mapstructure:"blocks_for_confirmation" description:"Minimum number of blocks to consider a block confirmed"`
	DefaultConnectTimeout     time.Duration `mapstructure:"default_connect_timeout" description:"The default connection timeout"`
	Lookup                    func(string) ([]net.IP, error)
	Dial                      func(string, string, time.Duration) (net.Conn, error)
	Checkpoints               []chaincfg.Checkpoint
	TimeSource                MedianTimeSource
}

func (c *AppConfig) WithoutAuthorization() *AppConfig {
	c.HTTPConfig.UseAuth = false
	return c
}
