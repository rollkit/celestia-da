package main

import (
	"context"
	"encoding/hex"
	"errors"
	"net"
	"os"

	rpc "github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/celestiaorg/celestia-node/share"
	logging "github.com/ipfs/go-log/v2"
	"github.com/rollkit/celestia-da"
	"github.com/rollkit/go-da/proxy"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func init() {
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
