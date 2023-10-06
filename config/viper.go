package config

import (
	"fmt"
	"log"

	"os"
	"strings"

	"github.com/libsv/bitcoin-hc/config/p2pconfig"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// NewViperConfig creates and returns new viper config.
func Init() *Config {
	// Use env vars
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	setDefaults()

	cfg := ParseConfig()
	err := cfg.P2P.Validate()
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

func setDefaults() {
	setHttpServerDefaults()
	setWebhookDefaults()
	setWebsocketDefaults()
	setP2PDefaults()
	setDefaultDb()
	setDefaultAuthorization()
}

// setDefaultDb edits and returns database-based viper config.
func setDefaultDb() {
	viper.SetDefault(EnvDb, "sqlite")
	viper.SetDefault(EnvResetDbOnStartup, false)
	viper.SetDefault(EnvDbFilePath, "../data/blockheaders.db")
	viper.SetDefault(EnvDbDsn, "file:../data/blockheaders.db?_foreign_keys=true&pooling=true")
	viper.SetDefault(EnvDbSchema, "../data/sql/migrations")
	viper.SetDefault(EnvDbMigrate, true)
	viper.SetDefault(EnvPreparedDb, false)
	viper.SetDefault(EnvPreparedDbFilePath, "../data/blockheaders.xz")
}

// WithAuthorization edits and returns authorization-based viper config.
func setDefaultAuthorization() {
	viper.SetDefault(EnvHttpServerUseAuth, true)
	viper.SetDefault(EnvHttpServerAuthToken, "mQZQ6WmxURxWz5ch")
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

// WithoutAuthorization edits and returns viper configuration with disabled authorization.
func (c *Config) WithoutAuthorization() *Config {
	viper.SetDefault(EnvHttpServerUseAuth, false)
	c.HTTP.UseAuth = false
	return c
}

// ParseP2PConfig init p2p viper config based on flags, env variables and json config.
func ParseConfig() *Config {
	f := initFlags()

	if f.ShowHelp {
		pflag.Usage()
		os.Exit(0)
	}

	if f.IgnoreFileConfig {
		fmt.Println("file config is ignored")
		viper.SetConfigFile(f.ConfigFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal(err)
		}
	}

	c := new(Config)

	if err := viper.Unmarshal(&c); err != nil {
		err = fmt.Errorf("config can't be unmarshaled %s", err.Error())
		log.Fatal(err)
	}

	c.P2P.Logger = p2pconfig.UseDefaultP2PLogger()

	return c
}

func initFlags() CLI {
	cli := CLI{}

	pflag.BoolVarP(&cli.ShowHelp, "help", "H", false, "show help")
	pflag.BoolVarP(&cli.ShowVersion, "version", "V", false, "print the version")
	pflag.BoolVar(&cli.IgnoreFileConfig, "ignoreconfig", false, "ignore file config")
	pflag.StringVarP(&cli.ConfigFile, "config", "C", p2pconfig.DefaultConfigDir, "path to configuration file")
	
	pflagsMapping()
	bindFlags()

	pflag.Parse()

	return cli
}
