package p2pconfig

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/btcsuite/go-socks/socks"
	"github.com/libsv/bitcoin-hc/internal/chaincfg"
	"github.com/libsv/bitcoin-hc/transports/p2p/connmgr"
)

// Config struct for p2pconfig.
type Config struct {
	LogDir                    string        `mapstructure:"logdir" description:"Directory to log output."`
	AddPeers                  []string      `mapstructure:"addpeer" description:"Add a peer to connect with at startup"`
	ConnectPeers              []string      `mapstructure:"connect" description:"Connect only to the specified peers at startup"`
	DisableListen             bool          `mapstructure:"nolisten" description:"Disable listening for incoming connections -- NOTE: Listening is automatically disabled if the --connect or --proxy options are used without also specifying listen interfaces via --listen"`
	Listeners                 []string      `mapstructure:"listen" description:"Add an interface/port to listen for connections (default all interfaces port: 8333, testnet: 18333)"`
	MaxPeers                  int           `mapstructure:"maxpeers" description:"Max number of inbound and outbound peers"`
	MaxPeersPerIP             int           `mapstructure:"maxpeersperip" description:"Max number of inbound and outbound peers per IP"`
	BanDuration               time.Duration `mapstructure:"banduration" description:"How long to ban misbehaving peers.  Valid time units are {s, m, h}.  Minimum 1 second"`
	MinSyncPeerNetworkSpeed   uint64        `mapstructure:"minsyncpeernetworkspeed" description:"Disconnect sync peers slower than this threshold in bytes/sec"`
	DisableDNSSeed            bool          `mapstructure:"nodnsseed" description:"Disable DNS seeding for peers"`
	ExternalIPs               []string      `mapstructure:"externalip" description:"Add an ip to the list of local addresses we claim to listen on to peers"`
	Proxy                     string        `mapstructure:"proxy" description:"Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	ProxyUser                 string        `mapstructure:"proxyuser" description:"Username for proxy server"`
	ProxyPass                 string        `mapstructure:"proxypass" default-mask:"-" description:"Password for proxy server"`
	OnionProxy                string        `mapstructure:"onion" description:"Connect to tor hidden services via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	OnionProxyUser            string        `mapstructure:"onionuser" description:"Username for onion proxy server"`
	OnionProxyPass            string        `mapstructure:"onionpass" default-mask:"-" description:"Password for onion proxy server"`
	NoOnion                   bool          `mapstructure:"noonion" description:"Disable connecting to tor hidden services"`
	TorIsolation              bool          `mapstructure:"torisolation" description:"Enable Tor stream isolation by randomizing user credentials for each connection."`
	TestNet3                  bool          `mapstructure:"testnet" description:"Use the test network"`
	RegressionTest            bool          `mapstructure:"regtest" description:"Use the regression test network"`
	SimNet                    bool          `mapstructure:"simnet" description:"Use the simulation test network"`
	AddCheckpoints            []string      `mapstructure:"addcheckpoint" description:"Add a custom checkpoint.  Format: '<height>:<hash>'"`
	DisableCheckpoints        bool          `mapstructure:"nocheckpoints" description:"Disable built-in checkpoints.  Don't do this unless you know what you're doing."`
	LogLevel                  string        `mapstructure:"loglevel" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical} -- You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set the log level for individual subsystems -- Use show to list available subsystems"`
	Upnp                      bool          `mapstructure:"upnp" description:"Use UPnP to map our listening port outside of NAT"`
	ExcessiveBlockSize        uint32        `mapstructure:"excessiveblocksize" description:"The maximum size block (in bytes) this node will accept. Cannot be less than 32000000."`
	TrickleInterval           time.Duration `mapstructure:"trickleinterval" description:"Minimum time between attempts to send new inventory to a connected peer"`
	UserAgentComments         []string      `mapstructure:"uacomment" description:"Comment to add to the user agent -- See BIP 14 for more information."`
	NoCFilters                bool          `mapstructure:"nocfilters" description:"Disable committed filtering (CF) support"`
	TargetOutboundPeers       uint32        `mapstructure:"targetoutboundpeers" description:"number of outbound connections to maintain"`
	BlocksForForkConfirmation int           `mapstructure:"blocksforconfirmation" description:"Minimum number of blocks to consider a block confirmed"`
	lookup                    func(string) ([]net.IP, error)
	oniondial                 func(string, string, time.Duration) (net.Conn, error)
	dial                      func(string, string, time.Duration) (net.Conn, error)
	AddCheckpointsSlice       []chaincfg.Checkpoint
	Checkpoints               []chaincfg.Checkpoint
	TimeSource                MedianTimeSource
}

// Override overrides config with non-zero values from second config.
// Used for overriding file config with values from flag config.
func (c *Config) Override(cfg *Config) {
	reflectCfgToOverride := reflect.ValueOf(c).Elem()
	overrideReflectCfg := reflect.ValueOf(cfg).Elem()

	numFields := overrideReflectCfg.NumField()

	for i := 0; i < numFields; i++ {
		newField := overrideReflectCfg.Field(i)

		if !newField.IsZero() {
			name := overrideReflectCfg.Type().Field(i).Name
			fieldToOverride := reflectCfgToOverride.FieldByName(name)
			if fieldToOverride.CanSet() {
				fieldToOverride.Set(newField)
			}
		}
	}
}

// Validate validates config and sets some parameters based on the config provided.
func (c *Config) Validate() (err error) {
	funcName := "validateConfig"

	// Don't add peers from the config file when in regression test mode.
	if c.RegressionTest && len(c.AddPeers) > 0 {
		c.AddPeers = nil
	}

	// Multiple networks can't be selected simultaneously.
	numNets := 0
	// Count number of network flags passed; assign active network params
	// while we're at it
	if c.TestNet3 {
		numNets++
		ActiveNetParams = &testNet3Params
	}
	if c.RegressionTest {
		numNets++
		ActiveNetParams = &regressionNetParams
	}
	if c.SimNet {
		numNets++
		// Also disable dns seeding on the simulation test network.
		ActiveNetParams = &simNetParams
		c.DisableDNSSeed = true
	}
	if numNets > 1 {
		str := "%s: The testnet, regtest, segnet, and simnet params " +
			"can't be used together -- choose one of the four"
		err := fmt.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	// Append the network type to the log directory so it is "namespaced"
	// per network in the same fashion as the data directory.
	c.LogDir = cleanAndExpandPath(c.LogDir)
	c.LogDir = filepath.Join(c.LogDir, netName(ActiveNetParams))

	// Don't allow ban durations that are too short.
	if c.BanDuration < time.Second {
		str := "%s: The banduration option may not be less than 1s -- parsed [%v]"
		err := fmt.Errorf(str, funcName, c.BanDuration)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	// --addPeer and --connect do not mix.
	if len(c.AddPeers) > 0 && len(c.ConnectPeers) > 0 {
		str := "%s: the --addpeer and --connect options can not be " +
			"mixed"
		err := fmt.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	// --proxy or --connect without --listen disables listening.
	if (c.Proxy != "" || len(c.ConnectPeers) > 0) &&
		len(c.Listeners) == 0 {
		c.DisableListen = true
	}

	// Connect means no DNS seeding.
	if len(c.ConnectPeers) > 0 {
		c.DisableDNSSeed = true
	}

	// Add the default listener if none were specified. The default
	// listener is all addresses on the listen port for the network
	// we are to connect to.
	if len(c.Listeners) == 0 {
		c.Listeners = []string{
			net.JoinHostPort("", ActiveNetParams.DefaultPort),
		}
	}

	// Excessive blocksize cannot be set less than the default but it can be higher.
	c.ExcessiveBlockSize = maxUint32(c.ExcessiveBlockSize, DefaultExcessiveBlockSize)

	// Prepend ExcessiveBlockSize signaling to the UserAgentComments
	c.UserAgentComments = append([]string{fmt.Sprintf("EB%.1f", float64(c.ExcessiveBlockSize)/1000000)}, c.UserAgentComments...)

	// Look for illegal characters in the user agent comments.
	for _, uaComment := range c.UserAgentComments {
		if strings.ContainsAny(uaComment, "/:()") {
			err := fmt.Errorf("%s: The following characters must not "+
				"appear in user agent comments: '/', ':', '(', ')'",
				funcName)
			fmt.Fprintln(os.Stderr, err)
			return err
		}
	}

	// Add default port to all listener addresses if needed and remove
	// duplicate addresses.
	c.Listeners = normalizeAddresses(c.Listeners,
		ActiveNetParams.DefaultPort)

	// Add default port to all added peer addresses if needed and remove
	// duplicate addresses.
	c.AddPeers = normalizeAddresses(c.AddPeers,
		ActiveNetParams.DefaultPort)
	c.ConnectPeers = normalizeAddresses(c.ConnectPeers,
		ActiveNetParams.DefaultPort)

	// --noonion and --onion do not mix.
	if c.NoOnion && c.OnionProxy != "" {
		err := fmt.Errorf("%s: the --noonion and --onion options may "+
			"not be activated at the same time", funcName)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	c.AddCheckpointsSlice, err = parseCheckpoints(c.AddCheckpoints)
	if err != nil {
		str := "%s: Error parsing checkpoints: %v"
		err := fmt.Errorf(str, funcName, err)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	if !c.DisableCheckpoints {
		c.Checkpoints = mergeCheckpoints(ActiveNetParams.Checkpoints, c.AddCheckpointsSlice)
	}

	// Tor stream isolation requires either proxy or onion proxy to be set.
	if c.TorIsolation && c.Proxy == "" && c.OnionProxy == "" {
		str := "%s: Tor stream isolation requires either proxy or " +
			"onionproxy to be set"
		err := fmt.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	// Setup dial and DNS resolution (lookup) functions depending on the
	// specified options.  The default is to use the standard
	// net.DialTimeout function as well as the system DNS resolver.  When a
	// proxy is specified, the dial function is set to the proxy specific
	// dial function and the lookup is set to use tor (unless --noonion is
	// specified in which case the system DNS resolver is used).
	c.dial = net.DialTimeout
	c.lookup = net.LookupIP
	if c.Proxy != "" {
		_, _, err := net.SplitHostPort(c.Proxy)
		if err != nil {
			str := "%s: Proxy address '%s' is invalid: %v"
			err := fmt.Errorf(str, funcName, c.Proxy, err)
			fmt.Fprintln(os.Stderr, err)
			return err
		}

		// Tor isolation flag means proxy credentials will be overridden
		// unless there is also an onion proxy configured in which case
		// that one will be overridden.
		torIsolation := false
		if c.TorIsolation && c.OnionProxy == "" &&
			(c.ProxyUser != "" || c.ProxyPass != "") {

			torIsolation = true
			fmt.Fprintln(os.Stderr, "Tor isolation set -- "+
				"overriding specified proxy user credentials")
		}

		proxy := &socks.Proxy{
			Addr:         c.Proxy,
			Username:     c.ProxyUser,
			Password:     c.ProxyPass,
			TorIsolation: torIsolation,
		}
		c.dial = proxy.DialTimeout

		// Treat the proxy as tor and perform DNS resolution through it
		// unless the --noonion flag is set or there is an
		// onion-specific proxy configured.
		if !c.NoOnion && c.OnionProxy == "" {
			c.lookup = func(host string) ([]net.IP, error) {
				return connmgr.TorLookupIP(host, c.Proxy)
			}
		}
	}

	// Setup onion address dial function depending on the specified options.
	// The default is to use the same dial function selected above.  However,
	// when an onion-specific proxy is specified, the onion address dial
	// function is set to use the onion-specific proxy while leaving the
	// normal dial function as selected above.  This allows .onion address
	// traffic to be routed through a different proxy than normal traffic.
	if c.OnionProxy != "" {
		_, _, err := net.SplitHostPort(c.OnionProxy)
		if err != nil {
			str := "%s: Onion proxy address '%s' is invalid: %v"
			err := fmt.Errorf(str, funcName, c.OnionProxy, err)
			fmt.Fprintln(os.Stderr, err)
			return err
		}

		// Tor isolation flag means onion proxy credentials will be
		// overridden.
		if c.TorIsolation &&
			(c.OnionProxyUser != "" || c.OnionProxyPass != "") {
			fmt.Fprintln(os.Stderr, "Tor isolation set -- "+
				"overriding specified onionproxy user "+
				"credentials ")
		}

		c.oniondial = func(network, addr string, timeout time.Duration) (net.Conn, error) {
			proxy := &socks.Proxy{
				Addr:         c.OnionProxy,
				Username:     c.OnionProxyUser,
				Password:     c.OnionProxyPass,
				TorIsolation: c.TorIsolation,
			}
			return proxy.DialTimeout(network, addr, timeout)
		}

		// When configured in bridge mode (both --onion and --proxy are
		// configured), it means that the proxy configured by --proxy is
		// not a tor proxy, so override the DNS resolution to use the
		// onion-specific proxy.
		if c.Proxy != "" {
			c.lookup = func(host string) ([]net.IP, error) {
				return connmgr.TorLookupIP(host, c.OnionProxy)
			}
		}
	} else {
		c.oniondial = c.dial
	}

	// Specifying --noonion means the onion address dial function results in
	// an error.
	if c.NoOnion {
		c.oniondial = func(a, b string, t time.Duration) (net.Conn, error) {
			return nil, errors.New("tor has been disabled")
		}
	}

	return nil
}
