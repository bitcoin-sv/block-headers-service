package cli

import (
	"fmt"
	"os"

	"github.com/bitcoin-sv/pulse/app/logger"
	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/database"
	"github.com/bitcoin-sv/pulse/version"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type cliFlags struct {
	showVersion   bool `mapstructure:"showVersion"`
	showHelp      bool `mapstructure:"showHelp"`
	exportHeaders bool `mapstructure:"exportHeaders"`
	dumpConfig    bool `mapstructure:"dumpConfig"`
}

func LoadFlags(cfg *config.AppConfig) error {
	if !anyFlagsPassed() {
		return nil
	}

	cli := &cliFlags{}
	pulseFlags := pflag.NewFlagSet("pulseFlags", pflag.ContinueOnError)

	initFlags(pulseFlags, cli)
	err := pulseFlags.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("error while parsing flags: %v", err.Error())
		os.Exit(1)
	}

	err = viper.BindPFlags(pulseFlags)
	if err != nil {
		fmt.Printf("error while binding flags: %v", err.Error())
		os.Exit(1)
	}

	parseCliFlags(cli, cfg)

	return nil
}

func anyFlagsPassed() bool {
	return len(os.Args) > 1
}

func initFlags(fs *pflag.FlagSet, cliFlags *cliFlags) {
	fs.StringP(config.ConfigFilePathKey, "C", "", "custom config file path")

	fs.BoolVar(&cliFlags.exportHeaders, "exportHeaders", false, "export headers from database to CSV file")
	fs.BoolVarP(&cliFlags.showHelp, "help", "h", false, "show help")
	fs.BoolVarP(&cliFlags.showVersion, "version", "v", false, "show version")
	fs.BoolVarP(&cliFlags.dumpConfig, "dump_config", "d", false, "dump config to file, specified by config_file flag")
}

func parseCliFlags(cli *cliFlags, cfg *config.AppConfig) {
	lf := logger.DefaultLoggerFactory()
	log := lf.NewLogger("flags")

	if cli.showHelp {
		pflag.Usage()
		os.Exit(0)
	}

	if cli.showVersion {
		fmt.Println("pulse", "version", version.String())
		os.Exit(0)
	}

	if cli.exportHeaders {
		if err := database.ExportHeaders(cfg, log); err != nil {
			fmt.Printf("\nError: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if cli.dumpConfig {
		configPath := viper.GetString(config.ConfigFilePathKey)
		if configPath == "" {
			configPath = config.DefaultConfigFilePath
		}

		err := viper.SafeWriteConfigAs(configPath)
		if err != nil {
			fmt.Printf("error while dumping config: %v", err.Error())
		}
		os.Exit(0)
	}
}
