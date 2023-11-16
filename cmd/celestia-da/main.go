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

func main() {
	// TODO(tzdybal): extract configuration and mainCmd from main func
	// TODO(tzdybal): read configuration from file (with viper)
	rpcAddress := "http://127.0.0.1:26658"
	rpcToken := ""
	namespace := ""

	listenAddress := "0.0.0.0:0"
	listenNetwork := "tcp"

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

	var rootCmd = &cobra.Command{
		Use:     "light [subcommand]",
		Args:    cobra.NoArgs,
		Short:   "Manage your Light node",
		PostRun: func(_ *cobra.Command, _ []string) { serve(rpcAddress, rpcToken, namespace) },
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdnode.PersistentPreRunEnv(cmd, node.Light, args)
		},
	}

	rootCmd.AddCommand(
		cmdnode.Init(flags...),
		cmdnode.Start(flags...),
		cmdnode.AuthCmd(flags...),
		cmdnode.ResetStore(flags...),
		cmdnode.RemoveConfigCmd(flags...),
		cmdnode.UpdateConfigCmd(flags...),
	)

	rootCmd.Flags().StringVar(&rpcAddress, "rpc.address", rpcAddress, "celestia-node RPC endpoint address")
	rootCmd.Flags().StringVar(&rpcToken, "rpc.token", "", "celestia-node RPC auth token")
	rootCmd.Flags().StringVar(&namespace, "namespace", "", "celestia namespace to use (hex encoded)")
	rootCmd.Flags().StringVar(&listenAddress, "listen.address", listenAddress,
		"Listen address")
	rootCmd.Flags().StringVar(&listenNetwork, "listen.network", listenNetwork,
		"Listen network type must be \"tcp\", \"tcp4\", \"tcp6\", \"unix\" or \"unixpacket\"")

	if err := rootCmd.MarkFlagRequired("rpc.token"); err != nil {
		log.Fatal("token:", err)
	}

	if err := rootCmd.MarkFlagRequired("namespace"); err != nil {
		log.Fatal("namespace:", err)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func serve(rpcAddress string, rpcToken string, nsString string) {
	ctx := context.Background()
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

	lis, err := net.Listen("tcp", "")
	if err != nil {
		log.Fatalln("failed to create network listener:", err)
	}
	log.Infoln("serving celestia-da over gRPC on:", lis.Addr())
	err = srv.Serve(lis)
	if !errors.Is(err, grpc.ErrServerStopped) {
		log.Fatalln("gRPC server stopped with error:", err)
	}
}
