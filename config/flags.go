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
	pflag.StringP(p2pConfigFilePath, "C", "", "path to configuration file")
	pflag.String(logdirFlag, "", "directory to log output")
	pflag.StringP(addpeerFlag, "a", "", "add a peer to connect with at startup")
	pflag.String(connectFlag, "", "connect only to the specified peers at startup")
	pflag.Bool(nolistenFlag, false, "disable listening for incoming connections -- NOTE: Listening is automatically disabled if the --connect or --proxy options are used without also specifying listen interfaces via --listen")
	pflag.String(listenFlag, "", "add an interface/port to listen for connections (default all interfaces port: 8333, testnet: 18333)")
	pflag.Int(maxpeersFlag, 0, "max number of inbound and outbound peers")
	pflag.Int(maxpeersperipFlag, 0, "max number of inbound and outbound peers per IP")
	pflag.Duration(bandurationFlag, 0, "how long to ban misbehaving peers. Valid time units are {s, m, h}. Minimum 1 second")
	pflag.Uint64(minsyncpeernetworkspeedFlag, 0, "disconnect sync peers slower than this threshold in bytes/sec")
	pflag.Bool(nodnsseedFlag, false, "disable DNS seeding for peers")
	pflag.String(externalipFlag, "", "add an ip to the list of local addresses we claim to listen on to peers")

	pflag.String(proxyFlag, "", "connect via SOCKS5 proxy (eg. 127.0.0.1:9050)")
	pflag.String(proxyuserFlag, "", "username for proxy server")
	pflag.String(proxypassFlag, "", "password for proxy server")

	pflag.String(onionFlag, "", "connect to tor hidden services via SOCKS5 proxy (eg. 127.0.0.1:9050)")
	pflag.String(onionuserFlag, "", "username for onion proxy server")
	pflag.String(onionpassFlag, "", "password for onion proxy server")
	pflag.Bool(noonionFlag, false, "disable connecting to tor hidden services")

	pflag.Bool(torisolationFlag, false, "enable Tor stream isolation by randomizing user credentials for each connection.")

	pflag.Bool(testnetFlag, false, "use the test network")
	pflag.Bool(regtestFlag, false, "use the regression test network")
	pflag.Bool(simnetFlag, false, "use the simulation test network")

	pflag.String(addcheckpointFlag, "", "add a custom checkpoint. Format: '<height>:<hash>'")
	pflag.Bool(nocheckpointsFlag, false, "disable built-in checkpoints. Don't do this unless you know what you're doing.")
	pflag.StringP(loglevelFlag, "l", "info", "logging level for all subsystems {trace, debug, info, warn, error, critical}")
	pflag.Bool(upnpFlag, false, "use UPnP to map our listening port outside of NAT")
	pflag.Uint32(excessiveblocksizeFlag, 0, "the maximum size block (in bytes) this node will accept. Cannot be less than 32000000")
	pflag.Duration(trickleintervalFlag, 0, "minimum time between attempts to send new inventory to a connected peer")
	pflag.String(uacommentFlag, "", "comment to add to the user agent -- See BIP 14 for more information.")
	pflag.Bool(nocfiltersFlag, false, "disable committed filtering (CF) suppor")
	pflag.Uint32(targetoutboundpeersFlag, 0, "number of outbound connections to maintain")
	pflag.Int(blocksforconfirmationFlag, 0, "minimum number of blocks to consider a block confirmed")
}

func dbFlagsMapping() {
	pflag.String(schemaPathFlag, "", "path to db migration files")
	pflag.String(dsnFlag, "", "data source name")
	pflag.Bool(migrateFlag, false, "flag specifying wheather to run migrations")
	pflag.Bool(resetStateFlag, false, "flag specifying wheather to clear db and start synchronization from genesis header or start from last header in db")
	pflag.String(dbFilePathFlag, "", "path to db file")
	pflag.Bool(preparedDbFlag, false, "flag specifying wheather to use prepared db")
	pflag.String(preparedDbFilePathFlag, "", "path to prepared db file.")
}

func websocketFlagsMapping() {
	pflag.Int(historyMaxFlag, 0, "max number of published events that should be hold and send to client in case of restored lost connection")
	pflag.Int(historyTTLFlag, 0, "max minutes for which published events should be hold and send to client in case of restored lost connection")
}

func webhookFlagsMapping() {
	pflag.Int(maxTriesFlag, 0, "max tries for webhook")
}

func httpFlagsMapping() {
	pflag.Int(readTimeoutFlag, 0, "http server read timeout")
	pflag.Int(writeTimeoutFlag, 0, "http server write timeout")
	pflag.Int(portFlag, 0, "http server port")
	pflag.String(urlPrefixFlag, "", "http server url prefix")
	pflag.Bool(useAuthFlag, false, "http server use auth")
	pflag.String(authTokenFlag, "", "http server admin auth token")
}

//nolint:all
func bindP2PFlags() {
	viper.BindPFlag(p2pConfigFilePath, pflag.Lookup(p2pConfigFilePath))
	viper.BindPFlag("p2p.logdir", pflag.Lookup(logdirFlag))
	viper.BindPFlag("p2p.addpeer", pflag.Lookup(addpeerFlag))
	viper.BindPFlag("p2p.connect", pflag.Lookup(connectFlag))
	viper.BindPFlag("p2p.nolisten", pflag.Lookup(nolistenFlag))
	viper.BindPFlag("p2p.listen", pflag.Lookup(listenFlag))
	viper.BindPFlag("p2p.maxpeers", pflag.Lookup(maxpeersFlag))
	viper.BindPFlag("p2p.maxpeersperip", pflag.Lookup(maxpeersperipFlag))
	viper.BindPFlag("p2p.banduration", pflag.Lookup(bandurationFlag))
	viper.BindPFlag("p2p.minsyncpeernetworkspeed", pflag.Lookup(minsyncpeernetworkspeedFlag))
	viper.BindPFlag("p2p.nodnsseed", pflag.Lookup(nodnsseedFlag))
	viper.BindPFlag("p2p.externalip", pflag.Lookup(externalipFlag))

	viper.BindPFlag("p2p.proxy", pflag.Lookup(proxyFlag))
	viper.BindPFlag("p2p.proxyuser", pflag.Lookup(proxyuserFlag))
	viper.BindPFlag("p2p.proxypass", pflag.Lookup(proxypassFlag))

	viper.BindPFlag("p2p.onion", pflag.Lookup(onionFlag))
	viper.BindPFlag("p2p.onionuser", pflag.Lookup(onionuserFlag))
	viper.BindPFlag("p2p.onionpass", pflag.Lookup(onionpassFlag))
	viper.BindPFlag("p2p.noonion", pflag.Lookup(noonionFlag))

	viper.BindPFlag("p2p.torisolation", pflag.Lookup(torisolationFlag))

	viper.BindPFlag("p2p.testnet", pflag.Lookup(testnetFlag))
	viper.BindPFlag("p2p.regtest", pflag.Lookup(regtestFlag))
	viper.BindPFlag("p2p.simnet", pflag.Lookup(simnetFlag))

	viper.BindPFlag("p2p.addcheckpoint", pflag.Lookup(addcheckpointFlag))
	viper.BindPFlag("p2p.nocheckpoints", pflag.Lookup(nocheckpointsFlag))
	viper.BindPFlag("p2p.loglevel", pflag.Lookup(loglevelFlag))
	viper.BindPFlag("p2p.upnp", pflag.Lookup(upnpFlag))
	viper.BindPFlag("p2p.excessiveblocksize", pflag.Lookup(excessiveblocksizeFlag))
	viper.BindPFlag("p2p.trickleinterval", pflag.Lookup(trickleintervalFlag))
	viper.BindPFlag("p2p.uacomment", pflag.Lookup(uacommentFlag))
	viper.BindPFlag("p2p.nocfilters", pflag.Lookup(nocfiltersFlag))
	viper.BindPFlag("p2p.targetoutboundpeers", pflag.Lookup(targetoutboundpeersFlag))
	viper.BindPFlag("p2p.blocksforconfirmation", pflag.Lookup(blocksforconfirmationFlag))
}

//nolint:all
func bindDBFlags() {
	viper.BindPFlag(EnvDbSchema, pflag.Lookup(schemaPathFlag))
	viper.BindPFlag(EnvDbDsn, pflag.Lookup(dsnFlag))
	viper.BindPFlag(EnvDbMigrate, pflag.Lookup(migrateFlag))
	viper.BindPFlag(EnvResetDbOnStartup, pflag.Lookup(resetStateFlag))
	viper.BindPFlag(EnvDbFilePath, pflag.Lookup(dbFilePathFlag))
	viper.BindPFlag(EnvPreparedDb, pflag.Lookup(preparedDbFlag))
	viper.BindPFlag(EnvPreparedDbFilePath, pflag.Lookup(preparedDbFilePathFlag))
}

//nolint:all
func bindHTTPFlags() {
	viper.BindPFlag(EnvHttpServerReadTimeout, pflag.Lookup(readTimeoutFlag))
	viper.BindPFlag(EnvHttpServerWriteTimeout, pflag.Lookup(writeTimeoutFlag))
	viper.BindPFlag(EnvHttpServerPort, pflag.Lookup(portFlag))
	viper.BindPFlag(EnvHttpServerUrlPrefix, pflag.Lookup(urlPrefixFlag))
	viper.BindPFlag(EnvHttpServerUseAuth, pflag.Lookup(useAuthFlag))
	viper.BindPFlag(EnvHttpServerAuthToken, pflag.Lookup(authTokenFlag))
}

//nolint:all
func bindWebsocketFlags() {
	viper.BindPFlag(EnvWebsocketHistoryMax, pflag.Lookup(historyMaxFlag))
	viper.BindPFlag(EnvWebsocketHistoryTtl, pflag.Lookup(historyTTLFlag))
}

//nolint:all
func bindWebhookFlags() {
	viper.BindPFlag(EnvWebhookMaxTries, pflag.Lookup(maxTriesFlag))
}
