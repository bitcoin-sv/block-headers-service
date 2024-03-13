package cli

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/database"
	"github.com/bitcoin-sv/block-headers-service/logging"
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
	appFlags := pflag.NewFlagSet("appFlags", pflag.ContinueOnError)

	initFlags(appFlags, cli)
	err := appFlags.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("error while parsing flags: %v", err.Error())
		os.Exit(1)
	}

	err = viper.BindPFlag(config.ConfigFilePathKey, appFlags.Lookup(config.ConfigFilePathKey))
	if err != nil {
		fmt.Printf("error while binding flags: %v", err.Error())
		os.Exit(1)
	}

	parseCliFlags(cli, cfg, appFlags)

	return nil
}

func anyFlagsPassed() bool {
	return len(os.Args) > 1
}

func initFlags(fs *pflag.FlagSet, cliFlags *cliFlags) {
	fs.StringP(config.ConfigFilePathKey, "C", "", "custom config file path")

	fs.BoolVarP(&cliFlags.exportHeaders, "export_headers", "e", false, "export headers from database to CSV file")
	fs.BoolVarP(&cliFlags.showHelp, "help", "h", false, "show help")
	fs.BoolVarP(&cliFlags.showVersion, "version", "v", false, "show version")
	fs.BoolVarP(&cliFlags.dumpConfig, "dump_config", "d", false, "dump config to file, specified by config_file flag")
}

func parseCliFlags(cli *cliFlags, cfg *config.AppConfig, appFlags *pflag.FlagSet) {
	log := logging.GetDefaultLogger().With().Str("service", "flags").Logger()

	if cli.showHelp {
		appFlags.PrintDefaults()
		os.Exit(0)
	}

	if cli.showVersion {
		fmt.Println(config.ApplicationName, config.Version())
		os.Exit(0)
	}

	if cli.exportHeaders {
		if err := database.ExportHeaders(cfg, &log); err != nil {
			log.Error().Msgf("error while exporting headers: %v", err.Error())
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
			log.Error().Msgf("error while dumping config: %v", err.Error())
		}
		os.Exit(0)
	}
}
