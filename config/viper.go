package config

import (
	"fmt"
	"log"

	// "net"
	"os"
	"strings"

	"github.com/libsv/bitcoin-hc/config/p2pconfig"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// NewViperConfig creates and returns new viper config.
func Load(appname string) *Config {
	// Use env vars
	viper.AutomaticEnv()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setP2PDefaults()
	p2pCfg := ParseP2PConfig(nil)
	p2pCfg.Override(nil)
	err := p2pCfg.Validate()
	if err != nil {
		log.Fatal(err)
	}

	setHttpServerDefaults()
	setWebhookDefaults()
	setWebsocketDefaults()
	return &Config{
		P2P: p2pCfg,
	}
}

// WithDb edits and returns database-based viper config.
func (c *Config) WithDb() *Config {
	viper.SetDefault(EnvDb, "sqlite")
	viper.SetDefault(EnvResetDbOnStartup, false)
	viper.SetDefault(EnvDbFilePath, "./data/blockheaders.db")
	viper.SetDefault(EnvDbDsn, "file:./data/blockheaders.db?_foreign_keys=true&pooling=true")
	viper.SetDefault(EnvDbSchema, "./data/sql/migrations")
	viper.SetDefault(EnvDbMigrate, true)
	viper.SetDefault(EnvPreparedDb, false)
	viper.SetDefault(EnvPreparedDbFilePath, "./data/blockheaders.xz")
	c.Db = &Db{
		Type:       DbType(viper.GetString(EnvDb)),
		Dsn:        viper.GetString(EnvDbDsn),
		SchemaPath: viper.GetString(EnvDbSchema),
		MigrateDb:  viper.GetBool(EnvDbMigrate),
	}
	return c
}

// WithAuthorization edits and returns authorization-based viper config.
func (c *Config) WithAuthorization() *Config {
	viper.SetDefault(EnvHttpServerUseAuth, true)
	viper.SetDefault(EnvHttpServerAuthToken, "mQZQ6WmxURxWz5ch")
	return c
}

// WithoutAuthorization edits and returns viper configuration with disabled authorization.
func (c *Config) WithoutAuthorization() *Config {
	viper.SetDefault(EnvHttpServerUseAuth, false)
	return c
}

// setHttpServerDefaults sets default values for http server.
func setHttpServerDefaults() {
	viper.SetDefault(EnvHttpServerReadTimeout, 10)
	viper.SetDefault(EnvHttpServerWriteTimeout, 10)
	viper.SetDefault(EnvHttpServerPort, 8080)
	viper.SetDefault(EnvHttpServerUrlPrefix, "/api/v1")
}

// setWebhookDefaults sets default values for webhook.
func setWebhookDefaults() {
	viper.SetDefault(EnvWebhookMaxTries, 10)
}

// setWebhookDefaults sets default values for websocket.
func setWebsocketDefaults() {
	viper.SetDefault(EnvWebsocketHistoryMax, 300)
	viper.SetDefault(EnvWebsocketHistoryTtl, 10)
}
func setP2PDefaults() {
	viper.SetDefault(EnvP2PLogLevel, p2pconfig.DefaultLogLevel)
	viper.SetDefault(EnvP2PMaxPeers, p2pconfig.DefaultMaxPeers)
	viper.SetDefault(EnvP2PMaxPeersPerIP, p2pconfig.DefaultMaxPeersPerIP)
	viper.SetDefault(EnvP2PMinSyncPeerNetworkSpeed, p2pconfig.DefaultMinSyncPeerNetworkSpeed)
	viper.SetDefault(EnvP2PBanDuration, p2pconfig.DefaultBanDuration)
	viper.SetDefault(EnvP2PLogDir, p2pconfig.DefaultLogDir)
	viper.SetDefault(EnvP2PExcessiveBlockSize, p2pconfig.DefaultExcessiveBlockSize)
	viper.SetDefault(EnvP2PTrickleInterval, p2pconfig.DefaultTrickleInterval)
	viper.SetDefault(EnvP2PBlocksForForkConfirmation, p2pconfig.DefaultBlocksToConfirmFork)
}

// ParseP2PConfig init p2p viper config based from specific config file.
func ParseP2PConfig(cliCfg *p2pconfig.Config) *p2pconfig.Config {
	f := initFlags()

	if f.ShowHelp {
		pflag.Usage()
		os.Exit(0)
	}

	flags := new(p2pconfig.Config)

	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(f)

	if err := viper.Unmarshal(&flags); err != nil {
		err = fmt.Errorf("config can't be unmarshaled %s", err.Error())
		log.Fatal(err)
	}
	fmt.Println(viper.GetString("logdir"))
	fmt.Println(viper.GetString("p2p_logdir"))
	fmt.Println(viper.AllKeys())
	fmt.Println(flags)

	viper.SetConfigFile(cliCfg.ConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	p2pCfg := new(p2pconfig.Config)

	if err := viper.Unmarshal(&p2pCfg); err != nil {
		err = fmt.Errorf("config can't be unmarshaled %s", err.Error())
		log.Fatal(err)
	}
	return p2pCfg
}

func initFlags() p2pconfig.Config {
	cfg := p2pconfig.Config{}

	pflag.BoolVarP(&cfg.ShowHelp, "help", "H", false, "show help")
	pflag.BoolVarP(&cfg.ShowVersion, "version", "V", false, "print the version")
	pflag.StringVarP(&cfg.ConfigFile, "config", "C", "", "path to configuration file")
	pflag.StringVar(&cfg.LogDir, "logdir", "", "directory to log output")
	pflag.StringArrayVarP(&cfg.AddPeers, "addpeer", "a", nil, "add a peer to connect with at startup")
	pflag.StringArrayVar(&cfg.ConnectPeers, "connect", nil, "connect only to the specified peers at startup")
	pflag.BoolVar(&cfg.DisableListen, "nolisten", false, "disable listening for incoming connections -- NOTE: Listening is automatically disabled if the --connect or --proxy options are used without also specifying listen interfaces via --listen")
	pflag.StringArrayVar(&cfg.Listeners, "listen", nil, "add an interface/port to listen for connections (default all interfaces port: 8333, testnet: 18333)")
	pflag.IntVar(&cfg.MaxPeers, "maxpeers", 0, "max number of inbound and outbound peers")
	pflag.IntVar(&cfg.MaxPeersPerIP, "maxpeersperip", 0, "max number of inbound and outbound peers per IP")
	pflag.DurationVar(&cfg.BanDuration, "banduration", 0, "how long to ban misbehaving peers. Valid time units are {s, m, h}. Minimum 1 second")
	pflag.Uint64Var(&cfg.MinSyncPeerNetworkSpeed, "minsyncpeernetworkspeed", 0, "disconnect sync peers slower than this threshold in bytes/sec")
	pflag.BoolVar(&cfg.DisableDNSSeed, "nodnsseed", false, "disable DNS seeding for peers")
	pflag.StringArrayVar(&cfg.ExternalIPs, "externalip", nil, "add an ip to the list of local addresses we claim to listen on to peers")

	pflag.StringVar(&cfg.Proxy, "proxy", "", "connect via SOCKS5 proxy (eg. 127.0.0.1:9050)")
	pflag.StringVar(&cfg.ProxyUser, "proxyuser", "", "username for proxy server")
	pflag.StringVar(&cfg.ProxyPass, "proxypass", "", "password for proxy server")

	pflag.StringVar(&cfg.OnionProxy, "onion", "", "connect to tor hidden services via SOCKS5 proxy (eg. 127.0.0.1:9050)")
	pflag.StringVar(&cfg.OnionProxyUser, "onionuser", "", "username for onion proxy server")
	pflag.StringVar(&cfg.OnionProxyPass, "onionpass", "", "password for onion proxy server")
	pflag.BoolVar(&cfg.NoOnion, "noonion", false, "disable connecting to tor hidden services")

	pflag.BoolVar(&cfg.TorIsolation, "torisolation", false, "enable Tor stream isolation by randomizing user credentials for each connection.")

	pflag.BoolVar(&cfg.TestNet3, "testnet", false, "use the test network")
	pflag.BoolVar(&cfg.RegressionTest, "regtest", false, "use the regression test network")
	pflag.BoolVar(&cfg.SimNet, "simnet", false, "use the simulation test network")

	pflag.StringArrayVar(&cfg.AddCheckpoints, "addcheckpoint", nil, "add a custom checkpoint. Format: '<height>:<hash>'")
	pflag.BoolVar(&cfg.DisableCheckpoints, "nocheckpoints", false, "disable built-in checkpoints. Don't do this unless you know what you're doing.")
	pflag.StringVarP(&cfg.LogLevel, "debuglevel", "d", "info", "logging level for all subsystems {trace, debug, info, warn, error, critical}")
	pflag.BoolVar(&cfg.Upnp, "upnp", false, "use UPnP to map our listening port outside of NAT")
	pflag.Uint32Var(&cfg.ExcessiveBlockSize, "excessiveblocksize", 0, "the maximum size block (in bytes) this node will accept. Cannot be less than 32000000")
	pflag.DurationVar(&cfg.TrickleInterval, "trickleinterval", 0, "minimum time between attempts to send new inventory to a connected peer")
	pflag.StringArrayVar(&cfg.UserAgentComments, "uacomment", nil, "comment to add to the user agent -- See BIP 14 for more information.")
	pflag.BoolVar(&cfg.NoCFilters, "nocfilters", false, "disable committed filtering (CF) suppor")
	pflag.Uint32Var(&cfg.TargetOutboundPeers, "targetoutboundpeers", 0, "number of outbound connections to maintain")
	pflag.IntVar(&cfg.BlocksForForkConfirmation, "blocksforconfirmation", 0, "minimum number of blocks to consider a block confirmed")

	pflag.Parse()

	return cfg
}
