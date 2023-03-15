package vconfig

var (
	EnvResetDbOnStartup = "db.resetState"
	EnvDbFilePath       = "db.dbFile.path"
)

const (
	EnvDb        = "db.type"
	EnvDbSchema  = "db.schema.path"
	EnvDbDsn     = "db.dsn"
	EnvDbMigrate = "db.migrate"
)

type DbType string

const (
	DBSqlite DbType = "sqlite"
)

// Config returns strongly typed config values.
type Config struct {
	Db *Db
}

type Db struct {
	Type       DbType
	SchemaPath string
	Dsn        string
	MigrateDb  bool
}
