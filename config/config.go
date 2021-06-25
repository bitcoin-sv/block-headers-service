package config

import (
	"fmt"
	"time"
)

// Environment variable constants.
const (
	EnvServerPort  = "server.port"
	EnvServerHost  = "server.host"
	EnvEnvironment = "env.environment"
	EnvMainNet     = "env.mainnet"
	EnvRegion      = "env.region"
	EnvVersion     = "env.version"
	EnvCommit      = "env.commit"
	EnvBuildDate   = "env.builddate"
	EnvLogLevel    = "log.level"
	EnvDb          = "db.type"
	EnvDbSchema    = "db.schema.path"
	EnvDbDsn       = "db.dsn"
	EnvWocURL      = "woc.url"

	LogDebug = "debug"
	LogInfo  = "info"
	LogError = "error"
	LogWarn  = "warn"
)

// Config returns strongly typed config values.
type Config struct {
	Logging    *Logging
	Server     *Server
	Deployment *Deployment
	Db         *Db
	Woc        *WocConfig
}

// Deployment contains information relating to the current
// deployed instance.
type Deployment struct {
	Environment string
	AppName     string
	Region      string
	Version     string
	Commit      string
	BuildDate   time.Time
	MainNet     bool
}

// IsDev determines if this app is running on a dev environment.
func (d *Deployment) IsDev() bool {
	return d.Environment == "dev"
}

func (d *Deployment) String() string {
	return fmt.Sprintf("Environment: %s \n AppName: %s\n Region: %s\n Version: %s\n Commit:%s\n BuildDate: %s\n",
		d.Environment, d.AppName, d.Region, d.Version, d.Commit, d.BuildDate)
}

// Logging contains log configuration.
type Logging struct {
	Level string
}

// Server contains all settings required to run a web server.
type Server struct {
	Port     string
	Hostname string
}

// Db contains database information.
type Db struct {
	Type       string
	SchemaPath string
	Dsn        string
}

// WocConfig contains params for connecting to whatsOnChain.
type WocConfig struct {
	URL string
}
