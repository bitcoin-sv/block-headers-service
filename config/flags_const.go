package config

// p2p consts
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

// db consts
const (
	schemaPathFlag         = "schemaPath"
	dsnFlag                = "dsn"
	migrateFlag            = "migrate"
	resetStateFlag         = "resetState"
	dbFilePathFlag         = "dbFilePath"
	preparedDbFlag         = "preparedDb"
	preparedDbFilePathFlag = "preparedDbFilePath"
)

// merkleroots consts
const maxBlockHeightExcess = "maxblockheightexcess"

// http consts
const (
	readTimeoutFlag  = "readTimeout"
	writeTimeoutFlag = "writeTimeout"
	portFlag         = "port"
	urlPrefixFlag    = "urlPrefix"
	useAuthFlag      = "useAuth"
	authTokenFlag    = "authToken"
)

// websockets and webhooks consts
const (
	historyMaxFlag = "historyMax"
	historyTTLFlag = "historyTTL"
	maxTriesFlag   = "maxTries"
)
