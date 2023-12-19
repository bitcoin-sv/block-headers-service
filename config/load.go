package config

import (
	"fmt"
	"strings"
	"sync"

	"os"

	"github.com/bitcoin-sv/pulse/config/p2pconfig"
	"github.com/bitcoin-sv/pulse/domains/logging"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Added a mutex lock for a race-condition
var viperLock sync.Mutex

// Load creates and returns a new viper config.
func Load(lf logging.LoggerFactory, cfg *AppConfig) (*AppConfig, error) {
	logger := lf.NewLogger("config")

	viperLock.Lock()
	defer viperLock.Unlock()

	if err := setDefaults(); err != nil {
		return nil, err
	}

	envConfig()

	if err := loadFromFile(logger); err != nil {
		return nil, err
	}

	if err := unmarshallToAppConfig(cfg); err != nil {
		return nil, err
	}

	err := cfg.P2PConfig.Validate()
	if err != nil {
		return nil, err
	}

	cfg.P2PConfig.TimeSource = p2pconfig.NewMedianTime(lf)

	return cfg, nil
}

func setDefaults() error {
	viper.SetDefault(ConfigFilePathKey, DefaultConfigFilePath)

	defaultsMap := make(map[string]interface{})
	if err := mapstructure.Decode(GetDefaultAppConfig(), &defaultsMap); err != nil {
		return err
	}

	for key, value := range defaultsMap {
		viper.SetDefault(key, value)
	}

	return nil
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
