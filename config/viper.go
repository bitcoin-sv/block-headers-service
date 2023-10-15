package config

import (
	"fmt"
	"log"

	"os"
	"strings"

	"github.com/libsv/bitcoin-hc/config/p2pconfig"
	"github.com/libsv/bitcoin-hc/domains/logging"
	"github.com/libsv/bitcoin-hc/version"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Init creates and returns a new viper config.
func Init(lf logging.LoggerFactory) *Config {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	setDefaults()

	cfg := ParseConfig()
	err := cfg.P2P.Validate()
	if err != nil {
		log.Printf("p2p config is invalid: %v", err)
		os.Exit(1)
	}
	cfg.P2P.TimeSource = p2pconfig.NewMedianTime(lf)
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

func setDefaultAuthorization() {
	viper.SetDefault(EnvHttpServerUseAuth, true)
	viper.SetDefault(EnvHttpServerAuthToken, "mQZQ6WmxURxWz5ch")
}

func setHttpServerDefaults() {
	viper.SetDefault(EnvHttpServerReadTimeout, 10)
	viper.SetDefault(EnvHttpServerWriteTimeout, 10)
	viper.SetDefault(EnvHttpServerPort, 8080)
	viper.SetDefault(EnvHttpServerUrlPrefix, "/api/v1")
}

func setWebhookDefaults() {
	viper.SetDefault(EnvWebhookMaxTries, 10)
}

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
	viper.Set(EnvHttpServerUseAuth, false)
	c.HTTP.UseAuth = false
	return c
}

// ParseConfig init viper config based on flags, env variables and json config.
func ParseConfig() *Config {
	f := initFlags()
	parseFlags(f)

	configFile := viper.GetString(p2pConfigFilePath)
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Printf("config cannot be read from path[%s]: %v", configFile, err)
			os.Exit(1)
		}
	}

	c := new(Config)

	if err := viper.Unmarshal(&c); err != nil {
		log.Printf("config can't be unmarshaled %v", err)
		os.Exit(1)
	}

	return c
}

func initFlags() CLI {
	cli := CLI{}

	fs := PulseFlagSet{}
	fs.BoolVarP(&cli.ShowHelp, "help", "H", false, "show help")
	fs.BoolVarP(&cli.ShowVersion, "version", "V", false, "print the version")

	fs.pflagsMapping()
	fs.bindFlags()

	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Printf("Flags can't be parsed: %v", err)
		os.Exit(1)
	}

	return cli
}

func parseFlags(cli CLI) {
	if cli.ShowHelp {
		pflag.Usage()
		os.Exit(0)
	}

	if cli.ShowVersion {
		fmt.Println("pulse", "version", version.String())
		os.Exit(0)
	}
}
