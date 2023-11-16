package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	cmdnode "github.com/celestiaorg/celestia-node/cmd"
	"github.com/celestiaorg/celestia-node/nodebuilder/core"
	"github.com/celestiaorg/celestia-node/nodebuilder/gateway"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/celestiaorg/celestia-node/nodebuilder/rpc"
	"github.com/celestiaorg/celestia-node/nodebuilder/state"
)

// NOTE: We should always ensure that the added Flags below are parsed somewhere, like in the
// PersistentPreRun func on parent command.

func init() {
	flags := []*pflag.FlagSet{
		cmdnode.NodeFlags(),
		p2p.Flags(),
		core.Flags(),
		cmdnode.MiscFlags(),
		rpc.Flags(),
		gateway.Flags(),
		state.Flags(),
	}

	startCmd := cmdnode.Start(append(flags, grpcFlags())...)
	if err := startCmd.MarkFlagRequired("grpc.token"); err != nil {
		log.Fatal("grpc.token:", err)
	}

	if err := startCmd.MarkFlagRequired("grpc.namespace"); err != nil {
		log.Fatal("grpc.namespace:", err)
	}
	startRunE := startCmd.RunE
	startCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Extract gRPC service flags
		rpcAddress, _ := cmd.Flags().GetString("grpc.address")
		rpcToken, _ := cmd.Flags().GetString("grpc.token")
		nsString, _ := cmd.Flags().GetString("grpc.namespace")
		listenAddress, _ := cmd.Flags().GetString("grpc.listen")
		listenNetwork, _ := cmd.Flags().GetString("grpc.network")

		// serve the gRPC service in a goroutine
		go serve(cmd.Context(), rpcAddress, rpcToken, listenAddress, listenNetwork, nsString)

		// Continue with the original start command execution
		return startRunE(cmd, args)
	}

	bridgeCmd.AddCommand(
		cmdnode.Init(flags...),
		startCmd,
		cmdnode.AuthCmd(flags...),
		cmdnode.ResetStore(flags...),
		cmdnode.RemoveConfigCmd(flags...),
		cmdnode.UpdateConfigCmd(flags...),
	)
}

var bridgeCmd = &cobra.Command{
	Use:   "bridge [subcommand]",
	Args:  cobra.NoArgs,
	Short: "Manage your Bridge node",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return cmdnode.PersistentPreRunEnv(cmd, node.Bridge, args)
	},
}
