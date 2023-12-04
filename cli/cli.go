package cli

import (
	"fmt"
	"os"

	"github.com/bitcoin-sv/pulse/app/logger"
	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/database"
	"github.com/bitcoin-sv/pulse/version"
	"github.com/spf13/pflag"
)

func ParseCliFlags(cli *config.CLI, cfg *config.Config) {

	lf := logger.DefaultLoggerFactory()
	log := lf.NewLogger("cli")

	if cli.ShowHelp {
		pflag.Usage()
		os.Exit(0)
	}

	if cli.ShowVersion {
		fmt.Println("pulse", "version", version.String())
		os.Exit(0)
	}

	if cli.ExportHeaders {
		if err := database.ExportHeaders(cfg, log); err != nil {
			fmt.Printf("\nError: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}
