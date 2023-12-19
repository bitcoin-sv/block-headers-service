package config

import (
	"github.com/bitcoin-sv/pulse/config/p2pconfig"
	"github.com/bitcoin-sv/pulse/domains/logging"
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
	P2PConfig        *p2pconfig.Config `mapstructure:"p2p"`
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
