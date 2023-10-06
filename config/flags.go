package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func pflagsMapping() {
	p2pFlagsMapping() 
	dbFlagsMapping()
	httpFlagsMapping()
	websocketFlagsMapping()
	webhookFlagsMapping()
}

func bindFlags() {
	bindP2PFlags()
	bindDBFlags()
	bindHTTPFlags()
	bindWebsocketFlags()
	bindWebhookFlags()
}

func p2pFlagsMapping() {
	pflag.String("logdir", "", "directory to log output")
	pflag.StringP("addpeer", "a", "", "add a peer to connect with at startup")
	pflag.String("connect", "", "connect only to the specified peers at startup")
	pflag.Bool("nolisten", false, "disable listening for incoming connections -- NOTE: Listening is automatically disabled if the --connect or --proxy options are used without also specifying listen interfaces via --listen")
	pflag.String("listen", "", "add an interface/port to listen for connections (default all interfaces port: 8333, testnet: 18333)")
	pflag.Int("maxpeers", 0, "max number of inbound and outbound peers")
	pflag.Int("maxpeersperip", 0, "max number of inbound and outbound peers per IP")
	pflag.Duration("banduration", 0, "how long to ban misbehaving peers. Valid time units are {s, m, h}. Minimum 1 second")
	pflag.Uint64("minsyncpeernetworkspeed", 0, "disconnect sync peers slower than this threshold in bytes/sec")
	pflag.Bool("nodnsseed", false, "disable DNS seeding for peers")
	pflag.String("externalip", "", "add an ip to the list of local addresses we claim to listen on to peers")

	pflag.String("proxy", "", "connect via SOCKS5 proxy (eg. 127.0.0.1:9050)")
	pflag.String("proxyuser", "", "username for proxy server")
	pflag.String("proxypass", "", "password for proxy server")

	pflag.String("onion", "", "connect to tor hidden services via SOCKS5 proxy (eg. 127.0.0.1:9050)")
	pflag.String("onionuser", "", "username for onion proxy server")
	pflag.String("onionpass", "", "password for onion proxy server")
	pflag.Bool("noonion", false, "disable connecting to tor hidden services")

	pflag.Bool("torisolation", false, "enable Tor stream isolation by randomizing user credentials for each connection.")

	pflag.Bool("testnet", false, "use the test network")
	pflag.Bool("regtest", false, "use the regression test network")
	pflag.Bool("simnet", false, "use the simulation test network")

	pflag.String("addcheckpoint", "", "add a custom checkpoint. Format: '<height>:<hash>'")
	pflag.Bool("nocheckpoints", false, "disable built-in checkpoints. Don't do this unless you know what you're doing.")
	pflag.StringP("loglevel", "l", "info", "logging level for all subsystems {trace, debug, info, warn, error, critical}")
	pflag.Bool("upnp", false, "use UPnP to map our listening port outside of NAT")
	pflag.Uint32("excessiveblocksize", 0, "the maximum size block (in bytes) this node will accept. Cannot be less than 32000000")
	pflag.Duration("trickleinterval", 0, "minimum time between attempts to send new inventory to a connected peer")
	pflag.String("uacomment", "", "comment to add to the user agent -- See BIP 14 for more information.")
	pflag.Bool("nocfilters", false, "disable committed filtering (CF) suppor")
	pflag.Uint32("targetoutboundpeers", 0, "number of outbound connections to maintain")
	pflag.Int("blocksforconfirmation", 0, "minimum number of blocks to consider a block confirmed")
}

func dbFlagsMapping() {
	pflag.String("schemaPath", "", "path to db migration files")
	pflag.String("dsn", "", "data source name")
	pflag.Bool("migrate", false, "flag specifying wheather to run migrations")
	pflag.Bool("resetState", false, "flag specifying wheather to clear db and start synchronization from genesis header or start from last header in db")
	pflag.String("dbFilePath", "", "path to db file")
	pflag.Bool("preparedDb", false, "flag specifying wheather to use prepared db")
	pflag.String("preparedDbFilePath", "", "path to prepared db file.")
}

func websocketFlagsMapping() {
	pflag.Int("historyMax", 0, "max number of published events that should be hold and send to client in case of restored lost connection")
	pflag.Int("historyTTL", 0, "max minutes for which published events should be hold and send to client in case of restored lost connection")
}

func webhookFlagsMapping() {
	pflag.Int("maxTries", 0, "max tries for webhook")
}

func httpFlagsMapping() {
	pflag.Int("readTimeout", 0, "http server read timeout")
	pflag.Int("writeTimeout", 0, "http server write timeout")
	pflag.Int("port", 0, "http server port")
	pflag.String("urlPrefix", "", "http server url prefix")
	pflag.Bool("useAuth", false, "http server use auth")
	pflag.String("authToken", "", "http server admin auth token")
}

//nolint:errcheck
func bindP2PFlags() {
	viper.BindPFlag("p2p.logdir", pflag.Lookup("logdir"))
	viper.BindPFlag("p2p.addpeer", pflag.Lookup("addpeer"))
	viper.BindPFlag("p2p.connect", pflag.Lookup("connect"))
	viper.BindPFlag("p2p.nolisten", pflag.Lookup("nolisten"))
	viper.BindPFlag("p2p.listen", pflag.Lookup("listen"))
	viper.BindPFlag("p2p.maxpeers", pflag.Lookup("maxpeers"))
	viper.BindPFlag("p2p.maxpeersperip", pflag.Lookup("maxpeersperip"))
	viper.BindPFlag("p2p.banduration", pflag.Lookup("banduration"))
	viper.BindPFlag("p2p.minsyncpeernetworkspeed", pflag.Lookup("minsyncpeernetworkspeed"))
	viper.BindPFlag("p2p.nodnsseed", pflag.Lookup("nodnsseed"))
	viper.BindPFlag("p2p.externalip", pflag.Lookup("externalip"))

	viper.BindPFlag("p2p.proxy", pflag.Lookup("proxy"))
	viper.BindPFlag("p2p.proxyuser", pflag.Lookup("proxyuser"))
	viper.BindPFlag("p2p.proxypass", pflag.Lookup("proxypass"))

	viper.BindPFlag("p2p.onion", pflag.Lookup("onion"))
	viper.BindPFlag("p2p.onionuser", pflag.Lookup("onionuser"))
	viper.BindPFlag("p2p.onionpass", pflag.Lookup("onionpass"))
	viper.BindPFlag("p2p.noonion", pflag.Lookup("noonion"))

	viper.BindPFlag("p2p.torisolation", pflag.Lookup("torisolation"))
	
	viper.BindPFlag("p2p.testnet", pflag.Lookup("testnet"))
	viper.BindPFlag("p2p.regtest", pflag.Lookup("regtest"))
	viper.BindPFlag("p2p.simnet", pflag.Lookup("simnet"))

	viper.BindPFlag("p2p.addcheckpoint", pflag.Lookup("addcheckpoint"))
	viper.BindPFlag("p2p.nocheckpoints", pflag.Lookup("nocheckpoints"))
	viper.BindPFlag("p2p.loglevel", pflag.Lookup("loglevel"))
	viper.BindPFlag("p2p.upnp", pflag.Lookup("upnp"))
	viper.BindPFlag("p2p.excessiveblocksize", pflag.Lookup("excessiveblocksize"))
	viper.BindPFlag("p2p.trickleinterval", pflag.Lookup("trickleinterval"))
	viper.BindPFlag("p2p.uacomment", pflag.Lookup("uacomment"))
	viper.BindPFlag("p2p.nocfilters", pflag.Lookup("nocfilters"))
	viper.BindPFlag("p2p.targetoutboundpeers", pflag.Lookup("targetoutboundpeers"))
	viper.BindPFlag("p2p.blocksforconfirmation", pflag.Lookup("blocksforconfirmation"))
}

//nolint:errcheck
func bindDBFlags() {
	viper.BindPFlag(EnvDbSchema, pflag.Lookup("schemaPath"))
	viper.BindPFlag(EnvDbDsn, pflag.Lookup("dsn"))
	viper.BindPFlag(EnvDbMigrate, pflag.Lookup("migrate"))
	viper.BindPFlag(EnvResetDbOnStartup, pflag.Lookup("resetState"))
	viper.BindPFlag(EnvDbFilePath, pflag.Lookup("dbFilePath"))
	viper.BindPFlag(EnvPreparedDb, pflag.Lookup("preparedDb"))
	viper.BindPFlag(EnvPreparedDbFilePath, pflag.Lookup("preparedDbFilePath"))
}

//nolint:errcheck
func bindHTTPFlags() {
	viper.BindPFlag(EnvHttpServerReadTimeout, pflag.Lookup("readTimeout"))
	viper.BindPFlag(EnvHttpServerWriteTimeout, pflag.Lookup("writeTimeout"))
	viper.BindPFlag(EnvHttpServerPort, pflag.Lookup("port"))
	viper.BindPFlag(EnvHttpServerUrlPrefix, pflag.Lookup("urlPrefix"))
	viper.BindPFlag(EnvHttpServerUseAuth, pflag.Lookup("useAuth"))
	viper.BindPFlag(EnvHttpServerAuthToken, pflag.Lookup("authToken"))
}

//nolint:errcheck
func bindWebsocketFlags() {
	viper.BindPFlag(EnvWebsocketHistoryMax, pflag.Lookup("historyMax"))
	viper.BindPFlag(EnvWebsocketHistoryTtl, pflag.Lookup("historyTTL"))
}

//nolint:errcheck
func bindWebhookFlags() {
	viper.BindPFlag(EnvWebhookMaxTries, pflag.Lookup("maxTries"))
}