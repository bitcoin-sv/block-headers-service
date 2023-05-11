package vconfig

// Define basic db config.
const (
	// EnvDb db type e.g. sqlite.
	EnvDb = "db.type"
	// EnvDbSchema path to db migration files.
	EnvDbSchema = "db.schema.path"
	// EnvDbDsn data source name.
	EnvDbDsn = "db.dsn"
	// EnvDbMigrate flag specifying wheather to run migrations.
	EnvDbMigrate = "db.migrate"
	// EnvResetDbOnStartup flag specifying wheather to clear db
	// and start synchronization from genesis header or start from last header in db.
	EnvResetDbOnStartup = "db.resetState"
	// EnvDbFilePath path to db file.
	EnvDbFilePath = "db.dbFile.path"
	// EnvPreparedDb flag specifying wheather to use prepared db.
	EnvPreparedDb = "db.preparedDb"
	// EnvPreparedDbFilePath path to prepared db file.
	EnvPreparedDbFilePath = "db.preparedDbFile.path"
)

// Define basic http server config.
const (
	// EnvHttpServerReadTimeout http server read timeout.
	EnvHttpServerReadTimeout = "http.server.readTimeout"
	// EnvHttpServerWriteTimeout http server write timeout.
	EnvHttpServerWriteTimeout = "http.server.writeTimeout"
	// EnvHttpServerPort http server port.
	EnvHttpServerPort = "http.server.port"
	// EnvHttpServerUrlPrefix http server url prefix.
	EnvHttpServerUrlPrefix = "http.server.urlPrefix"
	// EnvHttpServerUseAuth http server use auth.
	EnvHttpServerUseAuth = "http.server.useAuth"
	// EnvHttpServerAuthToken http server admin auth token.
	EnvHttpServerAuthToken = "http.server.authToken" // nolint:gosec
)

// DbType database type.
type DbType string

// DBSqlite creating config for sqlite db.
const DBSqlite DbType = "sqlite"

// Config returns strongly typed config values.
type Config struct {
	Db *Db
}

// Db represents a database connection.
type Db struct {
	Type       DbType
	SchemaPath string
	Dsn        string
	MigrateDb  bool
}
