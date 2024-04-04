package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
)

const (
	ApplicationName       = "block-headers-service"
	ConfigFilePathKey     = "config_file"
	DefaultConfigFilePath = "config.yaml"
	ConfigEnvPrefix       = "bhs"
)

var version = "should-be-overridden-by-setDefaults"

var Lookup func(string) ([]net.IP, error)
var Dial func(string, string, time.Duration) (net.Conn, error)
var Checkpoints []chaincfg.Checkpoint
var TimeSource MedianTimeSource

// DbEngine database engine.
type DbEngine string

const (
	DBSqlite     DbEngine = "sqlite"
	DBPostgreSql DbEngine = "postgres"
)

func Version() string {
	return version
}

// AppConfig returns strongly typed config values.
type AppConfig struct {
	Db         *DbConfig         `mapstructure:"db"`
	P2P        *P2PConfig        `mapstructure:"p2p"`
	MerkleRoot *MerkleRootConfig `mapstructure:"merkleroot"`
	Webhook    *WebhookConfig    `mapstructure:"webhook"`
	Websocket  *WebsocketConfig  `mapstructure:"websocket"`
	HTTP       *HTTPConfig       `mapstructure:"http"`
	Logging    *LoggingConfig    `mapstructure:"logging"`
	Metrics    *MetricsConfig    `mapstructure:"metrics"`
}

// DbConfig represents a database connection.
type DbConfig struct {
	// Engine is the engine of database [sqlite|postgres].
	Engine DbEngine `mapstructure:"engine"`
	// SchemaPath is the path to the database schema.
	SchemaPath string `mapstructure:"schema_path"`
	// PreparedDb is a flag for enabling prepared database.
	PreparedDb bool `mapstructure:"prepared_db"`
	// PreparedDbFilePath is the path to the prepared database file.
	PreparedDbFilePath string `mapstructure:"prepared_db_file_path"`

	Postgres PostgreSqlConfig `mapstructure:"postgres"`
	Sqlite   SqliteConfig     `mapstructure:"sqlite"`
}

type SqliteConfig struct {
	// FilePath is the path to the database file.
	FilePath string `mapstructure:"file_path"`
}

type PostgreSqlConfig struct {
	Host     string `mapstructure:"host"`
	Port     uint16 `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DbName   string `mapstructure:"db_name"`
	Sslmode  string `mapstructure:"ssl_mode"`
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
	UserAgentName         string        `mapstructure:"user_agent_name" description:"The name that should be used during announcement of the client on the p2p network"`
	UserAgentVersion      string        `mapstructure:"user_agent_version" description:"By default will be equal to application version, but can be overridden for development purposes"`
	Experimental          bool          `mapstructure:"experimental" description:"Turns on a new (highly experimental) way of getting headers with the usage of /internal/transports/p2p instead of /transports/p2p"`
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

// MetricsConfig represents a metrics config.
type MetricsConfig struct {
	// Enabled is a flag for enabling metrics.
	Enabled bool `mapstructure:"enabled"`
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

	if c.PreparedDb {
		if c.PreparedDbFilePath == "" {
			return errors.New("headers import: prepared database file path cannot be empty when prepared database is enabled")
		}
		if !fileExists(c.PreparedDbFilePath) {
			return fmt.Errorf("headers import: prepared database file does not exist at path %s", c.PreparedDbFilePath)
		}
	}

	switch c.Engine {
	case DBSqlite:
		if len(c.Sqlite.FilePath) == 0 {
			return fmt.Errorf("db: sqlite configuration cannot be empty where db type is set to %s", DBSqlite)
		}

	case DBPostgreSql:
		if c.Postgres.Host == "" || c.Postgres.Port == 0 || c.Postgres.User == "" || c.Postgres.DbName == "" {
			return fmt.Errorf("db: postgres configuration should be filled properly to use postgres engine %s", DBPostgreSql)
		}

	default:
		return errors.New("db: unsupported type")
	}

	return nil
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
