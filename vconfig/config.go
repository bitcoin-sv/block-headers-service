package vconfig

// Define basic db config.
const (
	// EnvDb db type e.g. sqlite.
	EnvDb               = "db.type"
	// EnvDbSchema path to db migration files.
	EnvDbSchema         = "db.schema.path"
	// EnvDbDsn data source name.
	EnvDbDsn            = "db.dsn"
	// EnvDbMigrate flag specifying wheather to run migrations.
	EnvDbMigrate        = "db.migrate"
	// EnvResetDbOnStartup flag specifying wheather to clear db
	// and start synchronisation from genesis header or start from last header in db.
	EnvResetDbOnStartup = "db.resetState"
	// EnvDbFilePath path to db file.
	EnvDbFilePath       = "db.dbFile.path"
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
