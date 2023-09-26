package vconfig

import (
	"fmt"
	"log"
	"strings"

	"github.com/libsv/bitcoin-hc/vconfig/p2pconfig"
	"github.com/spf13/viper"
)

// NewViperConfig creates and returns new viper config.
func NewViperConfig(appname string, cliCfg *p2pconfig.Config) *Config {
	// Use env vars
	viper.AutomaticEnv()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setP2PDefaults()
	p2pCfg := ParseP2PConfig(cliCfg)
	p2pCfg.Override(cliCfg)
	err := p2pCfg.Validate()
	if err != nil {
		log.Fatal(err)
	}

	setHttpServerDefaults()
	setWebhookDefaults()
	setWebsocketDefaults()
	return &Config{
		P2PConfig: p2pCfg,
	}
}

// WithDb edits and returns database-based viper configuration.
func (c *Config) WithDb() *Config {
	viper.SetDefault(EnvDb, "sqlite")
	viper.SetDefault(EnvResetDbOnStartup, false)
	viper.SetDefault(EnvDbFilePath, "./data/blockheaders.db")
	viper.SetDefault(EnvDbDsn, "file:./data/blockheaders.db?_foreign_keys=true&pooling=true")
	viper.SetDefault(EnvDbSchema, "./data/sql/migrations")
	viper.SetDefault(EnvDbMigrate, true)
	viper.SetDefault(EnvPreparedDb, false)
	viper.SetDefault(EnvPreparedDbFilePath, "./data/blockheaders.xz")
	c.Db = &Db{
		Type:       DbType(viper.GetString(EnvDb)),
		Dsn:        viper.GetString(EnvDbDsn),
		SchemaPath: viper.GetString(EnvDbSchema),
		MigrateDb:  viper.GetBool(EnvDbMigrate),
	}
	return c
}

// WithAuthorization edits and returns authorization-based viper configuration.
func (c *Config) WithAuthorization() *Config {
	viper.SetDefault(EnvHttpServerUseAuth, true)
	viper.SetDefault(EnvHttpServerAuthToken, "mQZQ6WmxURxWz5ch")
	return c
}

// WithoutAuthorization edits and returns viper configuration with disabled authorization.
func (c *Config) WithoutAuthorization() *Config {
	viper.SetDefault(EnvHttpServerUseAuth, false)
	return c
}

// setHttpServerDefaults sets default values for http server.
func setHttpServerDefaults() {
	viper.SetDefault(EnvHttpServerReadTimeout, 10)
	viper.SetDefault(EnvHttpServerWriteTimeout, 10)
	viper.SetDefault(EnvHttpServerPort, 8080)
	viper.SetDefault(EnvHttpServerUrlPrefix, "/api/v1")
}

// setWebhookDefaults sets default values for webhook.
func setWebhookDefaults() {
	viper.SetDefault(EnvWebhookMaxTries, 10)
}

// setWebhookDefaults sets default values for websocket.
func setWebsocketDefaults() {
	viper.SetDefault(EnvWebsocketHistoryMax, 300)
	viper.SetDefault(EnvWebsocketHistoryTtl, 10)
}
func setP2PDefaults() {
	viper.SetDefault(EnvP2PLogLevel, p2pconfig.DefaultLogLevel)
	viper.SetDefault(EnvP2PMaxPeers, p2pconfig.DefaultMaxPeers)
	viper.SetDefault(EnvP2PMaxPeersPerIP, p2pconfig.DefaultMaxPeersPerIP)
	viper.SetDefault(EnvP2PMinSyncPeerNetworkSpeed, p2pconfig.DefaultMinSyncPeerNetworkSpeed)
	viper.SetDefault(EnvP2PBanDuration, p2pconfig.DefaultBanDuration)
	viper.SetDefault(EnvP2PLogDir, p2pconfig.DefaultLogDir)
	viper.SetDefault(EnvP2PExcessiveBlockSize, p2pconfig.DefaultExcessiveBlockSize)
	viper.SetDefault(EnvP2PTrickleInterval, p2pconfig.DefaultTrickleInterval)
	viper.SetDefault(EnvP2PBlocksForForkConfirmation, p2pconfig.DefaultBlocksToConfirmFork)
}

// ParseP2PConfig init p2p viper config based from specific config file.
func ParseP2PConfig(cliCfg *p2pconfig.Config) *p2pconfig.Config {
	viper.SetConfigFile(cliCfg.ConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		err = fmt.Errorf("config can't read %s", err.Error())
		log.Fatal(err)
	}

	p2pCfg := new(p2pconfig.Config)

	if err := viper.Unmarshal(&p2pCfg); err != nil {
		err = fmt.Errorf("config can't be unmarshaled %s", err.Error())
		log.Fatal(err)
	}
	return p2pCfg
}
