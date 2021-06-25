package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// NewViperConfig will setup and return a new viper based configuration handler.
func NewViperConfig(appname string) *Config {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return &Config{}
}

// WithServer will setup the web server configuration if required.
func (c *Config) WithServer() *Config {
	viper.SetDefault(EnvServerPort, ":8442")
	viper.SetDefault(EnvServerHost, "localhost:8442")
	c.Server = &Server{
		Port:     viper.GetString(EnvServerPort),
		Hostname: viper.GetString(EnvServerHost),
	}
	return c
}

// WithDeployment sets up the deployment configuration if required.
func (c *Config) WithDeployment(appName string) *Config {
	viper.SetDefault(EnvEnvironment, "dev")
	viper.SetDefault(EnvRegion, "test")
	viper.SetDefault(EnvCommit, "test")
	viper.SetDefault(EnvVersion, "test")
	viper.SetDefault(EnvBuildDate, time.Now().UTC())
	viper.SetDefault(EnvMainNet, false)

	c.Deployment = &Deployment{
		Environment: viper.GetString(EnvEnvironment),
		Region:      viper.GetString(EnvRegion),
		Version:     viper.GetString(EnvVersion),
		Commit:      viper.GetString(EnvCommit),
		BuildDate:   viper.GetTime(EnvBuildDate),
		AppName:     appName,
		MainNet:     viper.GetBool(EnvMainNet),
	}
	return c
}

// WithLog sets up and returns log config.
func (c *Config) WithLog() *Config {
	viper.SetDefault(EnvLogLevel, "info")
	c.Logging = &Logging{Level: viper.GetString(EnvLogLevel)}
	return c
}

// WithDb sets up and returns database configuration.
func (c *Config) WithDb() *Config {
	viper.SetDefault(EnvDb, "sqlite")
	viper.SetDefault(EnvDbDsn, "file:data/blockheaders.db?cache=shared&_foreign_keys=true;")
	viper.SetDefault(EnvDbSchema,"data/sqlite/migrations")
	c.Db = &Db{
		Type: viper.GetString(EnvDb),
		Dsn:  viper.GetString(EnvDbDsn),
		SchemaPath: viper.GetString(EnvDbSchema),
	}
	return c
}

// WithWoc sets up and returns whatsonchain configuration.
func (c *Config) WithWoc() *Config {
	viper.SetDefault(EnvWocURL, "wss://socket.whatsonchain.com/blockheaders/history?from=")
	c.Woc = &WocConfig{
		URL: viper.GetString(EnvWocURL),
	}
	return c
}
