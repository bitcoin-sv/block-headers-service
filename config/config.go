package config

import (
	"github.com/bitcoin-sv/pulse/config/p2pconfig"
	"github.com/bitcoin-sv/pulse/domains/logging"
)

// Define basic db config.
const (
	// EnvDb db type e.g. sqlite.
	EnvDb = "db.type"
	// EnvDbSchema path to db migration files.
	EnvDbSchema = "db.schemaPath"
	// EnvDbDsn data source name.
	EnvDbDsn = "db.dsn"
	// EnvResetDbOnStartup flag specifying wheather to clear db
	// and start synchronization from genesis header or start from last header in db.
	EnvResetDbOnStartup = "db.resetState"
	// EnvDbFilePath path to db file.
	EnvDbFilePath = "db.dbFilePath"
	// EnvPreparedDb flag specifying wheather to use prepared db.
	EnvPreparedDb = "db.preparedDb"
	// EnvPreparedDbFilePath path to prepared db file.
	EnvPreparedDbFilePath = "db.preparedDbFilePath"
)

// Define basic http server config.
const (
	// EnvHttpServerReadTimeout http server read timeout.
	EnvHttpServerReadTimeout = "http.readTimeout"
	// EnvHttpServerWriteTimeout http server write timeout.
	EnvHttpServerWriteTimeout = "http.writeTimeout"
	// EnvHttpServerPort http server port.
	EnvHttpServerPort = "http.port"
	// EnvHttpServerUrlPrefix http server url prefix.
	EnvHttpServerUrlPrefix = "http.urlPrefix"
	// EnvHttpServerUseAuth http server use auth.
	EnvHttpServerUseAuth = "http.useAuth"
	// EnvHttpServerAuthToken http server admin auth token.
	EnvHttpServerAuthToken = "http.authToken" // nolint:gosec
)

// EnvMerklerootMaxBlockHeightExcess is a maximum number of blocks over the current Pulse
// top height that allows the given merkleroot to be UNABLE_TO_VERIFY instead of INVALID.
//
// Merkleroots with block height above that value are considered INVALID.
const EnvMerklerootMaxBlockHeightExcess = "merkleroot.maxBlockHeightExcess"

// EnvWebhookMaxTries max tries for webhook.
const EnvWebhookMaxTries = "webhook.maxTries"

const (
	// EnvWebsocketHistoryMax max number of published events that should be hold
	// and send to client in case of restored lost connection.
	EnvWebsocketHistoryMax = "websocket.historyMax"
	// EnvWebsocketHistoryTtl max minutes for which published events should be hold
	// and send to client in case of restored lost connection.
	EnvWebsocketHistoryTtl = "websocket.historyTTL"
)

const (
	// EnvP2PLogLevel p2p Logging level.
	EnvP2PLogLevel = "p2p.loglevel"
	// EnvP2PMaxPeers Max number of inbound and outbound peers.
	EnvP2PMaxPeers = "p2p.maxPeers"
	// EnvP2PMaxPeersPerIP Max number of inbound and outbound peers per IP.
	EnvP2PMaxPeersPerIP = "p2p.maxPeersPerIP"
	// EnvP2PMinSyncPeerNetworkSpeed Min Sync Speed.
	EnvP2PMinSyncPeerNetworkSpeed = "p2p.minSyncPeerNetworkSpeed"
	// EnvP2PBanDuration How long misbehaving peers should be banned for.
	EnvP2PBanDuration = "p2p.banduration"
	// EnvP2PLogDir Directory to log output.
	EnvP2PLogDir = "p2p.logdir"
	// EnvP2PExcessiveBlockSize The maximum size block (in bytes) this node will accept. Cannot be less than 32000000.
	EnvP2PExcessiveBlockSize = "p2p.excessiveBlockSize"
	// EnvP2PTrickleInterval Minimum time between attempts to send new inventory to a connected peer.
	EnvP2PTrickleInterval = "p2p.trickleInterval"
	// EnvP2PBlocksForForkConfirmation Minimum number of blocks to consider a block confirmed.
	EnvP2PBlocksForForkConfirmation = "p2p.blocksforconfirmation"
)

const (
	ApplicationName       = "pulse"
	APIVersion            = "v1"
	Version               = "v0.6.0"
	ConfigFilePathKey     = "config_file"
	DefaultConfigFilePath = "config.yaml"
	ConfigEnvPrefix       = "pulse_"
)

// DbType database type.
type DbType string

// DBSqlite creating config for sqlite db.
const DBSqlite DbType = "sqlite"

// AppConfig returns strongly typed config values.
type AppConfig struct {
	ConfigFile    string            `mapstructure:"configFile"`
	Db            *Db               `mapstructure:"db"`
	P2P           *p2pconfig.Config `mapstructure:"p2p"`
	Merkleroot    *Merkleroot       `mapstructure:"merkleroot"`
	Webhook       *Webhook          `mapstructure:"webhook"`
	Websocket     *Websocket        `mapstructure:"websocket"`
	HTTP          *HTTP             `mapstructure:"http"`
	LoggerFactory logging.LoggerFactory
}

// Db represents a database connection.
type Db struct {
	Type               DbType `mapstructure:"type"`
	SchemaPath         string `mapstructure:"schemaPath"`
	Dsn                string `mapstructure:"dsn"`
	ResetState         bool   `mapstructure:"resetState"`
	FilePath           string `mapstructure:"dbFilePath"`
	PreparedDb         bool   `mapstructure:"preparedDb"`
	PreparedDbFilePath string `mapstructure:"preparedDbFilePath"`
}

// Merkleroot represents merkleroots verification config.
type Merkleroot struct {
	MaxBlockHeightExcess int `mapstructure:"maxBlockHeightExcess"`
}

// Webhook represents a webhook config.
type Webhook struct {
	MaxTries int `mapstructure:"maxTries"`
}

// Websocket represents a websocket config.
type Websocket struct {
	HistoryMax int `mapstructure:"historyMax"`
	HistoryTTL int `mapstructure:"historyTTL"`
}

// HTTP represents a HTTP config.
type HTTP struct {
	ReadTimeout  int    `mapstructure:"readTimeout"`
	WriteTimeout int    `mapstructure:"writeTimeout"`
	Port         int    `mapstructure:"port"`
	UrlRefix     string `mapstructure:"urlPrefix"`
	UseAuth      bool   `mapstructure:"useAuth"`
	AuthToken    string `mapstructure:"authToken"`
}
