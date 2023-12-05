package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// PulseFlagSet custom FlagSet.
type PulseFlagSet struct {
	pflag.FlagSet
}

func (fs *PulseFlagSet) pflagsMapping() {
	fs.p2pFlagsMapping()
	fs.dbFlagsMapping()
	fs.httpFlagsMapping()
	fs.websocketFlagsMapping()
	fs.webhookFlagsMapping()
	fs.merklerootFlagsMapping()
}

func (fs *PulseFlagSet) bindFlags() {
	fs.bindP2PFlags()
	fs.bindDBFlags()
	fs.bindHTTPFlags()
	fs.bindWebsocketFlags()
	fs.bindWebhookFlags()
	fs.bindMerklerootFlags()
}

func (fs *PulseFlagSet) p2pFlagsMapping() {
	fs.StringP(p2pConfigFilePath, "C", "", "path to configuration file")
	fs.String(logdirFlag, "", "directory to log output")
	fs.StringP(addpeerFlag, "a", "", "add a peer to connect with at startup")
	fs.String(connectFlag, "", "connect only to the specified peers at startup")
	fs.Bool(nolistenFlag, false, "disable listening for incoming connections -- NOTE: Listening is automatically disabled if the --connect or --proxy options are used without also specifying listen interfaces via --listen")
	fs.String(listenFlag, "", "add an interface/port to listen for connections (default all interfaces port: 8333, testnet: 18333)")
	fs.Int(maxpeersFlag, 0, "max number of inbound and outbound peers")
	fs.Int(maxpeersperipFlag, 0, "max number of inbound and outbound peers per IP")
	fs.Duration(bandurationFlag, 0, "how long to ban misbehaving peers. Valid time units are {s, m, h}. Minimum 1 second")
	fs.Uint64(minsyncpeernetworkspeedFlag, 0, "disconnect sync peers slower than this threshold in bytes/sec")
	fs.Bool(nodnsseedFlag, false, "disable DNS seeding for peers")
	fs.String(externalipFlag, "", "add an ip to the list of local addresses we claim to listen on to peers")

	fs.String(proxyFlag, "", "connect via SOCKS5 proxy (eg. 127.0.0.1:9050)")
	fs.String(proxyuserFlag, "", "username for proxy server")
	fs.String(proxypassFlag, "", "password for proxy server")

	fs.String(onionFlag, "", "connect to tor hidden services via SOCKS5 proxy (eg. 127.0.0.1:9050)")
	fs.String(onionuserFlag, "", "username for onion proxy server")
	fs.String(onionpassFlag, "", "password for onion proxy server")
	fs.Bool(noonionFlag, false, "disable connecting to tor hidden services")

	fs.Bool(torisolationFlag, false, "enable Tor stream isolation by randomizing user credentials for each connection.")

	fs.Bool(testnetFlag, false, "use the test network")
	fs.Bool(regtestFlag, false, "use the regression test network")
	fs.Bool(simnetFlag, false, "use the simulation test network")

	fs.String(addcheckpointFlag, "", "add a custom checkpoint. Format: '<height>:<hash>'")
	fs.Bool(nocheckpointsFlag, false, "disable built-in checkpoints. Don't do this unless you know what you're doing.")
	fs.StringP(loglevelFlag, "l", "info", "logging level for all subsystems {trace, debug, info, warn, error, critical}")
	fs.Bool(upnpFlag, false, "use UPnP to map our listening port outside of NAT")
	fs.Uint32(excessiveblocksizeFlag, 0, "the maximum size block (in bytes) this node will accept. Cannot be less than 32000000")
	fs.Duration(trickleintervalFlag, 0, "minimum time between attempts to send new inventory to a connected peer")
	fs.String(uacommentFlag, "", "comment to add to the user agent -- See BIP 14 for more information.")
	fs.Bool(nocfiltersFlag, false, "disable committed filtering (CF) suppor")
	fs.Uint32(targetoutboundpeersFlag, 0, "number of outbound connections to maintain")
	fs.Int(blocksforconfirmationFlag, 0, "minimum number of blocks to consider a block confirmed")
}

func (fs *PulseFlagSet) dbFlagsMapping() {
	fs.String(schemaPathFlag, "", "path to db migration files")
	fs.String(dsnFlag, "", "data source name")
	fs.Bool(resetStateFlag, false, "flag specifying wheather to clear db and start synchronization from genesis header or start from last header in db")
	fs.String(dbFilePathFlag, "", "path to db file")
	fs.Bool(preparedDbFlag, false, "flag specifying wheather to use prepared db")
	fs.String(preparedDbFilePathFlag, "", "path to prepared db file.")
}

func (fs *PulseFlagSet) websocketFlagsMapping() {
	fs.Int(historyMaxFlag, 0, "max number of published events that should be hold and send to client in case of restored lost connection")
	fs.Int(historyTTLFlag, 0, "max minutes for which published events should be hold and send to client in case of restored lost connection")
}

func (fs *PulseFlagSet) merklerootFlagsMapping() {
	fs.Int32(maxBlockHeightExcess, 0, "max block height excess over the current top height for merkleroot verification")
}

func (fs *PulseFlagSet) webhookFlagsMapping() {
	fs.Int(maxTriesFlag, 0, "max tries for webhook")
}

func (fs *PulseFlagSet) httpFlagsMapping() {
	fs.Int(readTimeoutFlag, 0, "http server read timeout")
	fs.Int(writeTimeoutFlag, 0, "http server write timeout")
	fs.Int(portFlag, 0, "http server port")
	fs.String(urlPrefixFlag, "", "http server url prefix")
	fs.Bool(useAuthFlag, false, "http server use auth")
	fs.String(authTokenFlag, "", "http server admin auth token")
}

//nolint:all
func (fs *PulseFlagSet) bindP2PFlags() {
	viper.BindPFlag(p2pConfigFilePath, fs.Lookup(p2pConfigFilePath))
	viper.BindPFlag("p2p.logdir", fs.Lookup(logdirFlag))
	viper.BindPFlag("p2p.addpeer", fs.Lookup(addpeerFlag))
	viper.BindPFlag("p2p.connect", fs.Lookup(connectFlag))
	viper.BindPFlag("p2p.nolisten", fs.Lookup(nolistenFlag))
	viper.BindPFlag("p2p.listen", fs.Lookup(listenFlag))
	viper.BindPFlag("p2p.maxpeers", fs.Lookup(maxpeersFlag))
	viper.BindPFlag("p2p.maxpeersperip", fs.Lookup(maxpeersperipFlag))
	viper.BindPFlag("p2p.banduration", fs.Lookup(bandurationFlag))
	viper.BindPFlag("p2p.minsyncpeernetworkspeed", fs.Lookup(minsyncpeernetworkspeedFlag))
	viper.BindPFlag("p2p.nodnsseed", fs.Lookup(nodnsseedFlag))
	viper.BindPFlag("p2p.externalip", fs.Lookup(externalipFlag))

	viper.BindPFlag("p2p.proxy", fs.Lookup(proxyFlag))
	viper.BindPFlag("p2p.proxyuser", fs.Lookup(proxyuserFlag))
	viper.BindPFlag("p2p.proxypass", fs.Lookup(proxypassFlag))

	viper.BindPFlag("p2p.onion", fs.Lookup(onionFlag))
	viper.BindPFlag("p2p.onionuser", fs.Lookup(onionuserFlag))
	viper.BindPFlag("p2p.onionpass", fs.Lookup(onionpassFlag))
	viper.BindPFlag("p2p.noonion", fs.Lookup(noonionFlag))

	viper.BindPFlag("p2p.torisolation", fs.Lookup(torisolationFlag))

	viper.BindPFlag("p2p.testnet", fs.Lookup(testnetFlag))
	viper.BindPFlag("p2p.regtest", fs.Lookup(regtestFlag))
	viper.BindPFlag("p2p.simnet", fs.Lookup(simnetFlag))

	viper.BindPFlag("p2p.addcheckpoint", fs.Lookup(addcheckpointFlag))
	viper.BindPFlag("p2p.nocheckpoints", fs.Lookup(nocheckpointsFlag))
	viper.BindPFlag("p2p.loglevel", fs.Lookup(loglevelFlag))
	viper.BindPFlag("p2p.upnp", fs.Lookup(upnpFlag))
	viper.BindPFlag("p2p.excessiveblocksize", fs.Lookup(excessiveblocksizeFlag))
	viper.BindPFlag("p2p.trickleinterval", fs.Lookup(trickleintervalFlag))
	viper.BindPFlag("p2p.uacomment", fs.Lookup(uacommentFlag))
	viper.BindPFlag("p2p.nocfilters", fs.Lookup(nocfiltersFlag))
	viper.BindPFlag("p2p.targetoutboundpeers", fs.Lookup(targetoutboundpeersFlag))
	viper.BindPFlag("p2p.blocksforconfirmation", fs.Lookup(blocksforconfirmationFlag))
}

//nolint:all
func (fs *PulseFlagSet) bindDBFlags() {
	viper.BindPFlag(EnvDbSchema, fs.Lookup(schemaPathFlag))
	viper.BindPFlag(EnvDbDsn, fs.Lookup(dsnFlag))
	viper.BindPFlag(EnvResetDbOnStartup, fs.Lookup(resetStateFlag))
	viper.BindPFlag(EnvDbFilePath, fs.Lookup(dbFilePathFlag))
	viper.BindPFlag(EnvPreparedDb, fs.Lookup(preparedDbFlag))
	viper.BindPFlag(EnvPreparedDbFilePath, fs.Lookup(preparedDbFilePathFlag))
}

//nolint:all
func (fs *PulseFlagSet) bindHTTPFlags() {
	viper.BindPFlag(EnvHttpServerReadTimeout, fs.Lookup(readTimeoutFlag))
	viper.BindPFlag(EnvHttpServerWriteTimeout, fs.Lookup(writeTimeoutFlag))
	viper.BindPFlag(EnvHttpServerPort, fs.Lookup(portFlag))
	viper.BindPFlag(EnvHttpServerUrlPrefix, fs.Lookup(urlPrefixFlag))
	viper.BindPFlag(EnvHttpServerUseAuth, fs.Lookup(useAuthFlag))
	viper.BindPFlag(EnvHttpServerAuthToken, fs.Lookup(authTokenFlag))
}

//nolint:all
func (fs *PulseFlagSet) bindWebsocketFlags() {
	viper.BindPFlag(EnvWebsocketHistoryMax, fs.Lookup(historyMaxFlag))
	viper.BindPFlag(EnvWebsocketHistoryTtl, fs.Lookup(historyTTLFlag))
}

//nolint:all
func (fs *PulseFlagSet) bindMerklerootFlags() {
	viper.BindPFlag(EnvMerklerootMaxBlockHeightExcess, fs.Lookup(maxBlockHeightExcess))
}

//nolint:all
func (fs *PulseFlagSet) bindWebhookFlags() {
	viper.BindPFlag(EnvWebhookMaxTries, fs.Lookup(maxTriesFlag))
}
