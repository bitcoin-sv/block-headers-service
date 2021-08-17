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
	viper.SetDefault(EnvServerPort, ":8443")
	viper.SetDefault(EnvServerHost, "localhost:8443")
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
	viper.SetDefault(EnvDbDsn, "file:data/blockheaders.db?_foreign_keys=true&pooling=true;")
	viper.SetDefault(EnvDbSchema, "data/sql/migrations")
	viper.SetDefault(EnvDbMigrate, true)
	c.Db = &Db{
		Type:       DbType(viper.GetString(EnvDb)),
		Dsn:        viper.GetString(EnvDbDsn),
		SchemaPath: viper.GetString(EnvDbSchema),
		MigrateDb:  viper.GetBool(EnvDbMigrate),
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

// WithBitcoinNode sets up and returns bitcoin node configuration.
func (c *Config) WithBitcoinNode() *Config {
	c.Node = &BitcoinNode{
		Host:     viper.GetString(EnvNodeHost),
		Port:     viper.GetInt(EnvNodePort),
		Username: viper.GetString(EnvNodeUser),
		Password: viper.GetString(EnvNodePassword),
		UseSSL:   viper.GetBool(EnvNodeSSL),
	}
	return c
}

// WithHeaderClient sets up the header client with the type of
// syncing we wish to do.
func (c *Config) WithHeaderClient() *Config {
	c.Client = &HeaderClient{
		SyncType: viper.GetString(EnvHeaderType),
	}
	return c
}
