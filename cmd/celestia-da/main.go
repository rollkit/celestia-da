package main

import (
	"context"
	"os"

	cmdnode "github.com/celestiaorg/celestia-node/cmd"

	logging "github.com/ipfs/go-log/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var log = logging.Logger("cmd")

// WithSubcommands returns the node command where the start subcommand also starts the Data Availability gRPC service.
func WithSubcommands() func(*cobra.Command, []*pflag.FlagSet) {
	return func(c *cobra.Command, flags []*pflag.FlagSet) {
		c.AddCommand(
			cmdnode.Init(flags...),
			cmdnode.Start(WithDataAvailabilityService(flags)),
			cmdnode.AuthCmd(flags...),
			cmdnode.ResetStore(flags...),
			cmdnode.RemoveConfigCmd(flags...),
			cmdnode.UpdateConfigCmd(flags...),
		)
	}
}

func init() {
	bridgeCmd := cmdnode.NewBridge(WithSubcommands())
	lightCmd := cmdnode.NewLight(WithSubcommands())
	fullCmd := cmdnode.NewFull(WithSubcommands())
	rootCmd.AddCommand(lightCmd, bridgeCmd, fullCmd, versionCmd)
}

func main() {
	err := run()
	if err != nil {
		log.Errorln("application exited with error:", err)
		os.Exit(1)
	}
}

func run() error {
	return rootCmd.ExecuteContext(context.Background())
}

var rootCmd = &cobra.Command{
	Use: "celestia-da [  bridge  ||  full ||  light  ] [subcommand]",
	Short: `
	    ____      __          __  _
	  / ____/__  / /__  _____/ /_(_)___ _
	 / /   / _ \/ / _ \/ ___/ __/ / __  /
	/ /___/  __/ /  __(__  ) /_/ / /_/ /
	\____/\___/_/\___/____/\__/_/\__,_/
	`,
	Args: cobra.NoArgs,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: false,
	},
}
