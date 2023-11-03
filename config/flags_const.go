package config

// P2P consts.
const (
	p2pConfigFilePath           = "p2pconfig"
	logdirFlag                  = "logdir"
	addpeerFlag                 = "addpeer"
	connectFlag                 = "connect"
	nolistenFlag                = "nolisten"
	listenFlag                  = "listen"
	maxpeersFlag                = "maxpeers"
	maxpeersperipFlag           = "maxpeersperip"
	bandurationFlag             = "banduration"
	minsyncpeernetworkspeedFlag = "minsyncpeernetworkspeed"
	nodnsseedFlag               = "nodnsseed"
	externalipFlag              = "externalip"
	proxyFlag                   = "proxy"
	proxyuserFlag               = "proxyuser"
	proxypassFlag               = "proxypass"
	onionFlag                   = "onion"
	onionuserFlag               = "onionuser"
	onionpassFlag               = "onionpass"
	noonionFlag                 = "noonion"
	torisolationFlag            = "torisolation"
	testnetFlag                 = "testnet"
	regtestFlag                 = "regtest"
	simnetFlag                  = "simnet"
	addcheckpointFlag           = "addcheckpoint"
	nocheckpointsFlag           = "nocheckpoints"
	loglevelFlag                = "loglevel"
	upnpFlag                    = "upnp"
	excessiveblocksizeFlag      = "excessiveblocksize"
	trickleintervalFlag         = "trickleinterval"
	uacommentFlag               = "uacomment"
	nocfiltersFlag              = "nocfilters"
	targetoutboundpeersFlag     = "targetoutboundpeers"
	blocksforconfirmationFlag   = "blocksforconfirmation"
)

// DB consts.
const (
	schemaPathFlag         = "schemaPath"
	dsnFlag                = "dsn"
	migrateFlag            = "migrate"
	resetStateFlag         = "resetState"
	dbFilePathFlag         = "dbFilePath"
	preparedDbFlag         = "preparedDb"
	preparedDbFilePathFlag = "preparedDbFilePath"
)

// Merkleroots consts.
const maxBlockHeightExcess = "maxblockheightexcess"

// HTTP consts.
const (
	readTimeoutFlag  = "readTimeout"
	writeTimeoutFlag = "writeTimeout"
	portFlag         = "port"
	urlPrefixFlag    = "urlPrefix"
	useAuthFlag      = "useAuth"
	authTokenFlag    = "authToken"
)

// Websockets and Webhooks consts.
const (
	historyMaxFlag = "historyMax"
	historyTTLFlag = "historyTTL"
	maxTriesFlag   = "maxTries"
)
