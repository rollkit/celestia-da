package main

import (
	"context"
	"encoding/hex"
	"errors"
	"net"
	"os"

	rpc "github.com/celestiaorg/celestia-node/api/rpc/client"
	cmdnode "github.com/celestiaorg/celestia-node/cmd"
	"github.com/celestiaorg/celestia-node/cmd/celestia/bridge"
	"github.com/celestiaorg/celestia-node/cmd/celestia/full"
	"github.com/celestiaorg/celestia-node/cmd/celestia/light"
	"github.com/celestiaorg/celestia-node/share"
	logging "github.com/ipfs/go-log/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/rollkit/celestia-da"
	"github.com/rollkit/go-da/proxy"
)

var log = logging.Logger("cmd")

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

// WithDataAvailabilityService patches the start command to also run the gRPC Data Availability service
func WithDataAvailabilityService(flags []*pflag.FlagSet) func(*cobra.Command) {
	return func(c *cobra.Command) {
		grpcFlags := &pflag.FlagSet{}
		grpcFlags.String("grpc.address", "http://127.0.0.1:26658", "celestia-node RPC endpoint address")
		grpcFlags.String("grpc.token", "", "celestia-node RPC auth token")
		grpcFlags.String("grpc.namespace", "", "celestia namespace to use (hex encoded)")
		grpcFlags.String("grpc.listen", "", "gRPC service listen address")
		grpcFlags.String("grpc.network", "tcp", "gRPC service listen network type must be \"tcp\", \"tcp4\", \"tcp6\", \"unix\" or \"unixpacket\"")

		fset := append(flags, grpcFlags)

		for _, set := range fset {
			c.Flags().AddFlagSet(set)
		}

		requiredFlags := []string{"grpc.token", "grpc.namespace"}
		for _, required := range requiredFlags {
			if err := c.MarkFlagRequired(required); err != nil {
				log.Fatal(required, err)
			}
		}

		preRun := func(cmd *cobra.Command, args []string) {
			// Extract gRPC service flags
			rpcAddress, _ := cmd.Flags().GetString("grpc.address")
			rpcToken, _ := cmd.Flags().GetString("grpc.token")
			nsString, _ := cmd.Flags().GetString("grpc.namespace")
			listenAddress, _ := cmd.Flags().GetString("grpc.listen")
			listenNetwork, _ := cmd.Flags().GetString("grpc.network")

			// serve the gRPC service in a goroutine
			go serve(cmd.Context(), rpcAddress, rpcToken, listenAddress, listenNetwork, nsString)
		}

		c.PreRun = preRun
	}
}

// WithSubcommands returns the node command where the start command is patched with WithPatchStart
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
	bridgeCmd := bridge.NewCommand(WithSubcommands())
	lightCmd := light.NewCommand(WithSubcommands())
	fullCmd := full.NewCommand(WithSubcommands())
	rootCmd.AddCommand(lightCmd, bridgeCmd, fullCmd)
}

func main() {
	err := run()
	if err != nil {
		os.Exit(1)
	}
}

func run() error {
	return rootCmd.ExecuteContext(context.Background())
}

var rootCmd = &cobra.Command{
	Use: "celestia [  bridge  ||  full ||  light  ] [subcommand]",
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
