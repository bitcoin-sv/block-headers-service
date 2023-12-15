package config

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"os"

	"github.com/bitcoin-sv/pulse/domains/logging"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Added a mutex lock for a race-condition
var viperLock sync.Mutex

// Load creates and returns a new viper config.
func Load(lf logging.LoggerFactory) (*AppConfig, error) {
	logger := lf.NewLogger("config")

	viperLock.Lock()
	defer viperLock.Unlock()

	if err := setDefaults(); err != nil {
		return nil, err
	}

	envConfig()

	if err := loadFlags(DefaultAppConfig); err != nil {
		return nil, err
	}

	if err := loadFromFile(logger); err != nil {
		return nil, err
	}

	appConfig := new(AppConfig)
	if err := unmarshallToAppConfig(appConfig); err != nil {
		return nil, err
	}

	// TODO: check if this is needed
	// err := cfg.P2P.Validate()

	// cfg.P2P.TimeSource = p2pconfig.NewMedianTime(lf)

	return appConfig, nil
}

func setDefaults() error {
	viper.SetDefault(ConfigFilePathKey, DefaultConfigFilePath)

	defaultsMap := make(map[string]interface{})
	if err := mapstructure.Decode(DefaultAppConfig, &defaultsMap); err != nil {
		return err
	}

	for key, value := range defaultsMap {
		viper.SetDefault(key, value)
	}

	return nil

	// setHttpServerDefaults()
	// setMerkleRootDefaults()
	// setWebhookDefaults()
	// setWebsocketDefaults()
	// setP2PDefaults()
	// setDefaultDb()
	// setDefaultAuthorization()
}

func loadFromFile(logger logging.Logger) error {
	configFilePath := viper.GetString(ConfigFilePathKey)

	if configFilePath == DefaultConfigFilePath {
		_, err := os.Stat(DefaultConfigFilePath)
		if os.IsNotExist(err) {
			logger.Debug("Config file not specified. Using defaults")
			return nil
		}
		configFilePath = DefaultConfigFilePath
	}

	viper.SetConfigFile(configFilePath)
	if err := viper.ReadInConfig(); err != nil {
		err = fmt.Errorf("config cannot be read from path[%s]: %v", configFilePath, err.Error())
		logger.Error(err.Error())
		return err
	}

	return nil
}

func unmarshallToAppConfig(appConfig *AppConfig) error {
	if err := viper.Unmarshal(appConfig); err != nil {
		err = fmt.Errorf("config can't be unmarshaled %v", err.Error())
		return err
	}
	return nil
}

func envConfig() {
	viper.SetEnvPrefix("pulse")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

// // setDefaultDb edits and returns database-based viper config.
// func setDefaultDb() {
// 	viper.SetDefault(EnvDb, "sqlite")
// 	viper.SetDefault(EnvResetDbOnStartup, false)
// 	viper.SetDefault(EnvDbFilePath, "./data/blockheaders.db")
// 	viper.SetDefault(EnvDbDsn, "file:./data/blockheaders.db?_foreign_keys=true&pooling=true")
// 	viper.SetDefault(EnvDbSchema, "./database/migrations")
// 	viper.SetDefault(EnvPreparedDb, false)
// 	viper.SetDefault(EnvPreparedDbFilePath, "./data/blockheaders.csv.gz")
// }

// func setDefaultAuthorization() {
// 	viper.SetDefault(EnvHttpServerUseAuth, true)
// 	viper.SetDefault(EnvHttpServerAuthToken, "mQZQ6WmxURxWz5ch")
// }

// func setHttpServerDefaults() {
// 	viper.SetDefault(EnvHttpServerReadTimeout, 10)
// 	viper.SetDefault(EnvHttpServerWriteTimeout, 10)
// 	viper.SetDefault(EnvHttpServerPort, 8080)
// 	viper.SetDefault(EnvHttpServerUrlPrefix, "/api/v1")
// }

// func setMerkleRootDefaults() {
// 	viper.SetDefault(EnvMerklerootMaxBlockHeightExcess, 6)
// }

// func setWebhookDefaults() {
// 	viper.SetDefault(EnvWebhookMaxTries, 10)
// }

// func setWebsocketDefaults() {
// 	viper.SetDefault(EnvWebsocketHistoryMax, 300)
// 	viper.SetDefault(EnvWebsocketHistoryTtl, 10)
// }

// func setP2PDefaults() {
// 	viper.SetDefault(EnvP2PLogLevel, p2pconfig.DefaultLogLevel)
// 	viper.SetDefault(EnvP2PMaxPeers, p2pconfig.DefaultMaxPeers)
// 	viper.SetDefault(EnvP2PMaxPeersPerIP, p2pconfig.DefaultMaxPeersPerIP)
// 	viper.SetDefault(EnvP2PMinSyncPeerNetworkSpeed, p2pconfig.DefaultMinSyncPeerNetworkSpeed)
// 	viper.SetDefault(EnvP2PBanDuration, p2pconfig.DefaultBanDuration)
// 	viper.SetDefault(EnvP2PLogDir, p2pconfig.DefaultLogDir)
// 	viper.SetDefault(EnvP2PExcessiveBlockSize, p2pconfig.DefaultExcessiveBlockSize)
// 	viper.SetDefault(EnvP2PTrickleInterval, p2pconfig.DefaultTrickleInterval)
// 	viper.SetDefault(EnvP2PBlocksForForkConfirmation, p2pconfig.DefaultBlocksToConfirmFork)
// }

// WithoutAuthorization edits and returns viper configuration with disabled authorization.
func (c *AppConfig) WithoutAuthorization() *AppConfig {
	viper.Set(EnvHttpServerUseAuth, false)
	c.HTTP.UseAuth = false
	return c
}

// ParseConfig init viper config based on flags, env variables and json config.
func ParseConfig() *AppConfig {
	configFile := viper.GetString(p2pConfigFilePath)
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Printf("config cannot be read from path[%s]: %v", configFile, err)
			os.Exit(1)
		}
	}

	c := new(AppConfig)

	if err := viper.Unmarshal(&c); err != nil {
		log.Printf("config can't be unmarshaled %v", err)
		os.Exit(1)
	}

	return c
}
