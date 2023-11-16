package main

import (
	"context"
	"encoding/hex"
	"errors"
	"net"
	"os"

	logging "github.com/ipfs/go-log/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	rpc "github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/celestiaorg/celestia-node/nodebuilder/core"
	"github.com/celestiaorg/celestia-node/nodebuilder/gateway"
	"github.com/celestiaorg/celestia-node/nodebuilder/header"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/celestiaorg/celestia-node/nodebuilder/state"
	"github.com/celestiaorg/celestia-node/share"

	"github.com/rollkit/celestia-da"
	"github.com/rollkit/go-da/proxy"

	cmdnode "github.com/celestiaorg/celestia-node/cmd"
	noderpc "github.com/celestiaorg/celestia-node/nodebuilder/rpc"
)

var log = logging.Logger("cmd")

// grpcFlags contain the flags related to the gRPC DAService
func grpcFlags() *pflag.FlagSet {
	flags := &pflag.FlagSet{}

	flags.String("grpc.address", "http://127.0.0.1:26658", "celestia-node RPC endpoint address")
	flags.String("grpc.token", "", "celestia-node RPC auth token")
	flags.String("grpc.namespace", "", "celestia namespace to use (hex encoded)")
	flags.String("grpc.listen", "", "gRPC service listen address")
	flags.String("grpc.network", "tcp", "gRPC service listen network type must be \"tcp\", \"tcp4\", \"tcp6\", \"unix\" or \"unixpacket\"")

	return flags
}

func main() {
	// TODO(tzdybal): extract configuration and mainCmd from main func
	// TODO(tzdybal): read configuration from file (with viper)
	flags := []*pflag.FlagSet{
		cmdnode.NodeFlags(),
		p2p.Flags(),
		header.Flags(),
		cmdnode.MiscFlags(),
		// NOTE: for now, state-related queries can only be accessed
		// over an RPC connection with a celestia-core node.
		core.Flags(),
		noderpc.Flags(),
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

	var rootCmd = &cobra.Command{
		Use:   "light [subcommand]",
		Args:  cobra.NoArgs,
		Short: "Manage your Light node",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdnode.PersistentPreRunEnv(cmd, node.Light, args)
		},
	}

	rootCmd.AddCommand(
		cmdnode.Init(flags...),
		startCmd,
		cmdnode.AuthCmd(flags...),
		cmdnode.ResetStore(flags...),
		cmdnode.RemoveConfigCmd(flags...),
		cmdnode.UpdateConfigCmd(flags...),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func serve(ctx context.Context, rpcAddress, rpcToken, listenAddress, listenNetwork, nsString string) {
	client, err := rpc.NewClient(ctx, rpcAddress, rpcToken)
	if err != nil {
		log.Fatalln("failed to create celestia-node RPC client:", err)
	}
	nsBytes := make([]byte, len(nsString)/2)
	_, err = hex.Decode(nsBytes, []byte(nsString))
	if err != nil {
		log.Fatalln("invalid hex value of a namespace:", err)
	}
	namespace, err := share.NewBlobNamespaceV0(nsBytes)
	if err != nil {
		log.Fatalln("invalid namespace:", err)
	}

	da := celestia.NewCelestiaDA(client, namespace, ctx)
	// TODO(tzdybal): add configuration options for encryption
	srv := proxy.NewServer(da, grpc.Creds(insecure.NewCredentials()))

	lis, err := net.Listen(listenNetwork, listenAddress)
	if err != nil {
		log.Fatalln("failed to create network listener:", err)
	}
	defer func() {
		if err := lis.Close(); err != nil {
			log.Errorln("failed to close network listener:", err)
		}
	}()
	log.Infoln("serving celestia-da over gRPC on:", lis.Addr())
	err = srv.Serve(lis)
	if !errors.Is(err, grpc.ErrServerStopped) {
		log.Fatalln("gRPC server stopped with error:", err)
	}
}
