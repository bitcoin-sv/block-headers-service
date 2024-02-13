package config

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/bitcoin-sv/block-headers-service/logging"
	"github.com/rs/zerolog"

	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Added a mutex lock for a race-condition.
var viperLock sync.Mutex

// Load creates and returns a new viper config.
func Load(cfg *AppConfig) (*AppConfig, *zerolog.Logger, error) {
	viperLock.Lock()
	defer viperLock.Unlock()

	if err := loadFromFile(); err != nil {
		return nil, nil, err
	}

	if err := unmarshallToAppConfig(cfg); err != nil {
		return nil, nil, err
	}

	logger, err := logging.CreateLogger(cfg.Logging.InstanceName, cfg.Logging.Format, cfg.Logging.Level, cfg.Logging.LogOrigin)
	if err != nil {
		return nil, nil, err
	}

	setP2PLogger(logger)
	return cfg, logger, nil
}

func SetDefaults(log *zerolog.Logger) error {
	viper.SetDefault(ConfigFilePathKey, DefaultConfigFilePath)

	defaultsMap := make(map[string]interface{})
	if err := mapstructure.Decode(GetDefaultAppConfig(), &defaultsMap); err != nil {
		return err
	}

	for key, value := range defaultsMap {
		viper.SetDefault(key, value)
	}

	setP2PDefaults(log)
	envConfig()

	return nil
}

func setP2PDefaults(defaultLog *zerolog.Logger) {
	setP2PLogger(defaultLog)
	Lookup = net.LookupIP
	Dial = net.DialTimeout
	Checkpoints = ActiveNetParams.Checkpoints
}

func loadFromFile() error {
	defaultLog := logging.GetDefaultLogger()
	configFilePath := viper.GetString(ConfigFilePathKey)

	if configFilePath == DefaultConfigFilePath {
		_, err := os.Stat(DefaultConfigFilePath)
		if os.IsNotExist(err) {
			defaultLog.Debug().Msg("Config file not specified. Using defaults")
			return nil
		}
		configFilePath = DefaultConfigFilePath
	}

	viper.SetConfigFile(configFilePath)
	if err := viper.ReadInConfig(); err != nil {
		err = fmt.Errorf("config cannot be read from path[%s]: %v", configFilePath, err.Error())
		defaultLog.Error().Msg(err.Error())
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
	viper.SetEnvPrefix(ConfigEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func setP2PLogger(log *zerolog.Logger) {
	TimeSource = NewMedianTime(log)
}
