package config

import (
	"fmt"
	"regexp"
	"time"

	validator "github.com/theflyingcodr/govalidator"
)

// Environment variable constants.
const (
	EnvServerPort   = "server.port"
	EnvServerHost   = "server.host"
	EnvEnvironment  = "env.environment"
	EnvMainNet      = "env.mainnet"
	EnvRegion       = "env.region"
	EnvVersion      = "env.version"
	EnvCommit       = "env.commit"
	EnvBuildDate    = "env.builddate"
	EnvLogLevel     = "log.level"
	EnvDb           = "db.type"
	EnvDbSchema     = "db.schema.path"
	EnvDbDsn        = "db.dsn"
	EnvDbMigrate    = "db.migrate"
	EnvWocURL       = "woc.url"
	EnvNodeHost     = "node.host"
	EnvNodePort     = "node.port"
	EnvNodeUser     = "node.username"
	EnvNodePassword = "node.password"
	EnvNodeSSL      = "node.usessl"
	EnvHeaderType   = "header.sync.type"

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
	Node       *BitcoinNode
	Client     *HeaderClient
}

// Validate will check config values are valid and return a list of failures
// if any have been found.
func (c *Config) Validate() error {
	vl := validator.New()
	if c.Db != nil {
		c.Db.Validate(vl)
	}
	c.Client.Validate(vl)
	return vl.Err()
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

var reDbType = regexp.MustCompile(`^(sqlite|mysql|postgres)$`)

// DbType is used to restrict the dbs we can support.
type DbType string

// Supported database types.
const (
	DBSqlite   DbType = "sqlite"
	DBMySql    DbType = "mysql"
	DBPostgres DbType = "postgres"
)

// Db contains database information.
type Db struct {
	Type       DbType
	SchemaPath string
	Dsn        string
	MigrateDb  bool
}

// Validate will ensure the HeaderClient config is valid.
func (d *Db) Validate(v validator.ErrValidation) {
	v = v.Validate("db.type", validator.MatchString(string(d.Type), reDbType))
}

// WocConfig contains params for connecting to whatsOnChain.
type WocConfig struct {
	URL string
}

// BitcoinNode config params for connecting to a bitcoin node.
type BitcoinNode struct {
	Host     string
	Port     int
	Username string
	Password string
	UseSSL   bool
}

const (
	SyncWoc  = "woc"
	SyncNode = "node"
)

var reSyncType = regexp.MustCompile(`^(woc|node)$`)

// HeaderClient contains params for setting up the header client server.
type HeaderClient struct {
	SyncType string
}

// Validate will ensure the HeaderClient config is valid.
func (h *HeaderClient) Validate(v validator.ErrValidation) {
	v = v.Validate("syncType", validator.MatchString(h.SyncType, reSyncType))
}
