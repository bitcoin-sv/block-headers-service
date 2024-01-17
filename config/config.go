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
	Db         *DbConfig         `mapstructure:"db"`
	P2P        *P2PConfig        `mapstructure:"p2p"`
	MerkleRoot *MerkleRootConfig `mapstructure:"merkleroot"`
	Webhook    *WebhookConfig    `mapstructure:"webhook"`
	Websocket  *WebsocketConfig  `mapstructure:"websocket"`
	HTTP       *HTTPConfig       `mapstructure:"http"`
	Logging    *LoggingConfig    `mapstructure:"logging"`
}

// DbConfig represents a database connection.
type DbConfig struct {
	Type               DbType `mapstructure:"type"`
	SchemaPath         string `mapstructure:"schema_path"`
	Dsn                string `mapstructure:"dsn"`
	FilePath           string `mapstructure:"file_path"`
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
	ReadTimeout               int    `mapstructure:"read_timeout"`
	WriteTimeout              int    `mapstructure:"write_timeout"`
	Port                      int    `mapstructure:"port"`
	UseAuth                   bool   `mapstructure:"use_auth"`
	AuthToken                 string `mapstructure:"auth_token"`
	ProfilingEndpointsEnabled bool   `mapstructure:"debug_profiling"`
}

// P2PConfig represents a p2p config.
type P2PConfig struct {
	BanDuration               time.Duration `mapstructure:"ban_duration" description:"How long to ban misbehaving peers.  Valid time units are {s, m, h}.  Minimum 1 second"`
	DisableCheckpoints        bool          `mapstructure:"disable_checkpoints" description:"Disable built-in checkpoints.  Don't do this unless you know what you're doing."`
	BlocksForForkConfirmation int           `mapstructure:"blocks_for_confirmation" description:"Minimum number of blocks to consider a block confirmed"`
	DefaultConnectTimeout     time.Duration `mapstructure:"default_connect_timeout" description:"The default connection timeout"`
	Lookup                    func(string) ([]net.IP, error)
	Dial                      func(string, string, time.Duration) (net.Conn, error)
	Checkpoints               []chaincfg.Checkpoint
	TimeSource                MedianTimeSource
}

// LoggingConfig represents a logging config.
type LoggingConfig struct {
	Level        string `mapstructure:"level"`
	Format       string `mapstructure:"format"`
	InstanceName string `mapstructure:"instance_name"`
	LogOrigin    bool   `mapstructure:"origin"`
}

func (c *AppConfig) WithoutAuthorization() *AppConfig {
	c.HTTP.UseAuth = false
	return c
}
