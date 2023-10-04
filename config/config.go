package config

import "github.com/libsv/bitcoin-hc/config/p2pconfig"

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

// EnvWebhookMaxTries max tries for webhook.
const EnvWebhookMaxTries = "webhook.maxTries"

const (
	// EnvWebsocketHistoryMax max number of published events that should be hold
	// and send to client in case of restored lost connection.
	EnvWebsocketHistoryMax = "websocket.history.max"
	// EnvWebsocketHistoryTtl max minutes for which published events should be hold
	// and send to client in case of restored lost connection.
	EnvWebsocketHistoryTtl = "websocket.history.ttl"
)

const (
	// EnvP2PLogLevel p2p Logging level.
	EnvP2PLogLevel = "p2p_loglevel"
	// EnvP2PMaxPeers Max number of inbound and outbound peers.
	EnvP2PMaxPeers = "p2p_maxPeers"
	// EnvP2PMaxPeersPerIP Max number of inbound and outbound peers per IP.
	EnvP2PMaxPeersPerIP = "p2p_maxPeersPerIP"
	// EnvP2PMinSyncPeerNetworkSpeed Min Sync Speed.
	EnvP2PMinSyncPeerNetworkSpeed = "p2p_minSyncPeerNetworkSpeed"
	// EnvP2PBanDuration How long misbehaving peers should be banned for.
	EnvP2PBanDuration = "p2p_banduration"
	// EnvP2PLogDir Directory to log output.
	EnvP2PLogDir = "p2p_logdir"
	// EnvP2PExcessiveBlockSize The maximum size block (in bytes) this node will accept. Cannot be less than 32000000.
	EnvP2PExcessiveBlockSize = "p2p_excessiveBlockSize"
	// EnvP2PTrickleInterval Minimum time between attempts to send new inventory to a connected peer.
	EnvP2PTrickleInterval = "p2p_trickleInterval"
	// EnvP2PBlocksForForkConfirmation Minimum number of blocks to consider a block confirmed.
	EnvP2PBlocksForForkConfirmation = "p2p_blocksforconfirmation"
)

// DbType database type.
type DbType string

// DBSqlite creating config for sqlite db.
const DBSqlite DbType = "sqlite"

// Config returns strongly typed config values.
type Config struct {
	Db  *Db
	P2P *p2pconfig.Config
}

// Db represents a database connection.
type Db struct {
	Type       DbType
	SchemaPath string
	Dsn        string
	MigrateDb  bool
}
