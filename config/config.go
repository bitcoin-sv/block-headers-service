package config

import (
	"errors"
	"fmt"
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

// DbType database type.
type DbType string

const (
	DBSqlite     DbType = "sqlite"
	DBPostgreSql DbType = "postgres"
)

// DbConfig represents a database connection.
type DbConfig struct {
	// Type is the type of database [sqlite|postgres].
	Type DbType `mapstructure:"type"`
	// SchemaPath is the path to the database schema.
	SchemaPath string `mapstructure:"schema_path"`
	// PreparedDb is a flag for enabling prepared database.
	PreparedDb bool `mapstructure:"prepared_db"`
	// PreparedDbFilePath is the path to the prepared database file.
	PreparedDbFilePath string `mapstructure:"prepared_db_file_path"`

	Postgres *PostgreSqlConfig `mapstructure:"postgres"`
	Sqlite   SqliteConfig      `mapstructure:"sqlite"`
}

type SqliteConfig struct {
	// FilePath is the path to the database file.
	FilePath string `mapstructure:"file_path"`
}

type PostgreSqlConfig struct {
	Host     string
	Port     uint16
	User     string
	Password string
	DbName   string
	Sslmode  string
}

// MerkleRootConfig represents merkleroots verification config.
type MerkleRootConfig struct {
	// MaxBlockHeightExcess is the maximum number of blocks that can be ahead of the current tip.
	MaxBlockHeightExcess int `mapstructure:"max_block_height_excess"`
}

// WebhookConfig represents a webhook config.
type WebhookConfig struct {
	// MaxTries is the maximum number of tries to send a webhook.
	MaxTries int `mapstructure:"max_tries"`
}

// WebsocketConfig represents a websocket config.
type WebsocketConfig struct {
	// HistoryMax is the maximum number of history items to keep in memory.
	HistoryMax int `mapstructure:"history_max"`
	// HistoryTTL is the maximum duration for keeping history in memory.
	HistoryTTL int `mapstructure:"history_ttl"`
}

// HTTPConfig represents a HTTPConfig config.
type HTTPConfig struct {
	// ReadTimeout is the maximum duration for reading the request.
	ReadTimeout int `mapstructure:"read_timeout"`
	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout int `mapstructure:"write_timeout"`
	// Port is the port to listen on for connections.
	Port int `mapstructure:"port"`
	// UseAuth is a flag for enabling authorization.
	UseAuth bool `mapstructure:"use_auth"`
	// AuthToken is a token for authorization.
	AuthToken string `mapstructure:"auth_token"`
	// ProfilingEndpointsEnabled is a flag for enabling additional endpoits for profiling with use of pprof.
	ProfilingEndpointsEnabled bool `mapstructure:"debug_profiling"`
}

// P2PConfig represents a p2p config.
type P2PConfig struct {
	// BanDuration is the duration to ban misbehaving peers.
	BanDuration time.Duration `mapstructure:"ban_duration" description:"How long to ban misbehaving peers.  Valid time units are {s, m, h}.  Minimum 1 second"`
	// DisableCheckpoints is a flag for disabling built-in checkpoints.
	DisableCheckpoints bool `mapstructure:"disable_checkpoints" description:"Disable built-in checkpoints.  Don't do this unless you know what you're doing."`
	// BlocksForForkConfirmation is the minimum number of blocks to consider a block confirmed.
	BlocksForForkConfirmation int `mapstructure:"blocks_for_confirmation" description:"Minimum number of blocks to consider a block confirmed"`
	// DefaultConnectTimeout is the default connection timeout.
	DefaultConnectTimeout time.Duration `mapstructure:"default_connect_timeout" description:"The default connection timeout"`
	Lookup                func(string) ([]net.IP, error)
	Dial                  func(string, string, time.Duration) (net.Conn, error)
	// Checkpoints is a list of checkpoints.
	Checkpoints []chaincfg.Checkpoint
	// TimeSource is the time source.
	TimeSource MedianTimeSource
}

// LoggingConfig represents a logging config.
type LoggingConfig struct {
	// Level is the log level.
	Level string `mapstructure:"level"`
	// Format is the log format.
	Format string `mapstructure:"format"`
	// InstanceName is the name of the instance.
	InstanceName string `mapstructure:"instance_name"`
	// LogOrigin is a flag for enabling log origin.
	LogOrigin bool `mapstructure:"origin"`
}

func (c *AppConfig) WithoutAuthorization() *AppConfig {
	c.HTTP.UseAuth = false
	return c
}

func (c *AppConfig) Validate() error {
	if err := c.Db.Validate(); err != nil {
		return err
	}

	return nil
}

func (c *DbConfig) Validate() error {
	if c == nil {
		return errors.New("db: configuration cannot be empty")
	}

	switch c.Type {
	case DBSqlite:
		if len(c.Sqlite.FilePath) == 0 {
			return fmt.Errorf("db: sqlite configuration cannot be empty wher db type is set to %s", DBSqlite)
		}

	case DBPostgreSql:
		if c.Postgres == nil {
			return fmt.Errorf("db: postgres configuration cannot be empty wher db type is set to %s", DBPostgreSql)
		}

	default:
		return errors.New("db: unsupported type")
	}

	return nil
}
