package vconfig

import (
	"strings"

	"github.com/spf13/viper"
)

// NewViperConfig creates and returns new viper config.
func NewViperConfig(appname string) *Config {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
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
	c.Db = &Db{
		Type:       DbType(viper.GetString(EnvDb)),
		Dsn:        viper.GetString(EnvDbDsn),
		SchemaPath: viper.GetString(EnvDbSchema),
		MigrateDb:  viper.GetBool(EnvDbMigrate),
	}
	return c
}
