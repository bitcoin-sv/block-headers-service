package config

import (
	"log"

	"os"
	"strings"

	cliFlags "github.com/bitcoin-sv/pulse/cli/flags"
	"github.com/bitcoin-sv/pulse/config/p2pconfig"
	"github.com/bitcoin-sv/pulse/domains/logging"
	"github.com/spf13/viper"
)

// Init creates and returns a new viper config.
func Init(lf logging.LoggerFactory) (*Config, *cliFlags.CliFlags) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	setDefaults()

	fs := &PulseFlagSet{}
	cli := &cliFlags.CliFlags{}

	initFlags(fs, cli)

	cfg := ParseConfig()
	err := cfg.P2P.Validate()
	if err != nil {
		log.Printf("p2p config is invalid: %v", err)
		os.Exit(1)
	}

	cfg.P2P.TimeSource = p2pconfig.NewMedianTime(lf)

	return cfg, cli
}

func setDefaults() {
	setHttpServerDefaults()
	setMerkleRootDefaults()
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
	viper.SetDefault(EnvDbFilePath, "./data/blockheaders.db")
	viper.SetDefault(EnvDbDsn, "file:./data/blockheaders.db?_foreign_keys=true&pooling=true")
	viper.SetDefault(EnvDbSchema, "./database/migrations")
	viper.SetDefault(EnvPreparedDb, false)
	viper.SetDefault(EnvPreparedDbFilePath, "./data/blockheaders.csv.gz")
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

func setMerkleRootDefaults() {
	viper.SetDefault(EnvMerklerootMaxBlockHeightExcess, 6)
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

func initFlags(fs *PulseFlagSet, cli *cliFlags.CliFlags) {
	fs.pflagsMapping()
	fs.bindFlags()
	initCliFlags(fs, cli)

	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Printf("Flags can't be parsed: %v", err)
		os.Exit(1)
	}
}

func initCliFlags(fs *PulseFlagSet, cli *cliFlags.CliFlags) {
	fs.BoolVarP(&cli.ShowHelp, "help", "H", false, "show help")
	fs.BoolVarP(&cli.ShowVersion, "version", "V", false, "print the version")
	fs.BoolVar(&cli.ExportHeaders, "exportHeaders", false, "export headers from database to CSV file")
}
