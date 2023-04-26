package vconfig

import (
	"strings"

	"github.com/spf13/viper"
)

// NewViperConfig creates and returns new viper config.
func NewViperConfig(appname string) *Config {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	setHttpServerDefaults()
	return &Config{}
}

// WithDb edits and returns database-based viper configuration.
func (c *Config) WithDb() *Config {
	viper.SetDefault(EnvDb, "sqlite")
	viper.SetDefault(EnvResetDbOnStartup, false)
	viper.SetDefault(EnvDbFilePath, "./data/blockheaders.db")
	viper.SetDefault(EnvDbDsn, "file:./data/blockheaders.db?_foreign_keys=true&pooling=true")
	viper.SetDefault(EnvDbSchema, "./data/sql/migrations")
	viper.SetDefault(EnvDbMigrate, true)
	viper.SetDefault(EnvPreparedDb, true)
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

// setHttpServerDefaults sets default values for http server.
func setHttpServerDefaults() {
	viper.SetDefault(EnvHttpServerReadTimeout, 10)
	viper.SetDefault(EnvHttpServerWriteTimeout, 10)
	viper.SetDefault(EnvHttpServerPort, 8080)
}
