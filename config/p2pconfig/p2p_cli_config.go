package p2pconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/libsv/bitcoin-hc/version"
)

// runServiceCommand is only set to a real function on Windows.  It is used
// to parse and execute service commands specified via the -s flag.
var runServiceCommand func(string) error

// serviceOptions defines the configuration options for the daemon as a service on
// Windows.
type serviceOptions struct {
	ServiceCommand string `short:"s" long:"service" description:"Service command {install, remove, start, stop}"`
}

// newConfigParser returns a new command line flags parser.
func newConfigParser(cfg *Config, so *serviceOptions, options flags.Options) *flags.Parser {
	parser := flags.NewParser(cfg, options)
	if runtime.GOOS == "windows" {
		_, err := parser.AddGroup("Service Options", "Service Options", so)
		if err != nil {
			cfg.Logger.Error(err)
		}
	}
	return parser
}

// ParseFlags returns Config from flags.
func ParseFlags(customWorkingDirectory string) (*Config, error) {
	var workingDirectory string
	if len(customWorkingDirectory) > 0 {
		workingDirectory = customWorkingDirectory
	} else {
		workingDirectory = getWorkingDirectory()
	}
	cfg := Config{
		ConfigFile: filepath.Join(workingDirectory, Defaultp2pConfigPath),
		Logger:     useDefaultLogger(),
	}

	// Service options which are only added on Windows.
	serviceOpts := serviceOptions{}

	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified.
	flagsParser := newConfigParser(&cfg, &serviceOpts, flags.HelpFlag)
	_, err := flagsParser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}
	}

	// Show the version and exit if the version flag was specified.
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	if cfg.ShowVersion {
		fmt.Println(appName, "version", version.String())
		os.Exit(0)
	}

	// Perform service command and exit if specified.  Invalid service
	// commands show an appropriate error.  Only runs on Windows since
	// the runServiceCommand function will be nil when not on Windows.
	if serviceOpts.ServiceCommand != "" && runServiceCommand != nil {
		err := runServiceCommand(serviceOpts.ServiceCommand)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(0)
	}
	return &cfg, nil
}

// DefaultP2PConfig returns default p2pConfig.
func DefaultP2PConfig(customWorkingDirectory string) Config {
	var workingDirectory string
	if len(customWorkingDirectory) > 0 {
		workingDirectory = customWorkingDirectory
	} else {
		workingDirectory = getWorkingDirectory()
	}

	return Config{
		ConfigFile:                filepath.Join(workingDirectory, Defaultp2pConfigPath),
		LogLevel:                  DefaultLogLevel,
		MaxPeers:                  DefaultMaxPeers,
		MaxPeersPerIP:             DefaultMaxPeersPerIP,
		MinSyncPeerNetworkSpeed:   DefaultMinSyncPeerNetworkSpeed,
		BanDuration:               DefaultBanDuration,
		LogDir:                    DefaultLogDir,
		ExcessiveBlockSize:        DefaultExcessiveBlockSize,
		TrickleInterval:           DefaultTrickleInterval,
		BlocksForForkConfirmation: DefaultBlocksToConfirmFork,
		Logger:                    useDefaultLogger(),
	}
}
