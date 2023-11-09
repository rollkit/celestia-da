package main

import (
	"context"
	"encoding/hex"
	"net"

	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"

	rpc "github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/celestiaorg/celestia-node/nodebuilder"
	"github.com/celestiaorg/celestia-node/share"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/rollkit/celestia-da"
	"github.com/rollkit/go-da/proxy"
)

func main() {
	// TODO(tzdybal): extract configuration and mainCmd from main func
	// TODO(tzdybal): read configuration from file (with viper)
	rpcAddress := "http://127.0.0.1:26658"
	rpcToken := ""
	namespace := ""

	listenAddress := "0.0.0.0:0"
	listenNetwork := "tcp"

	rootCmd := &cobra.Command{
		Use:   "celestia-da",
		Short: "Celesia DA layer gRPC server for rollkit",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// override config with all modifiers passed on start
			cfg := NodeConfig(ctx)

			storePath := StorePath(ctx)
			keysPath := filepath.Join(storePath, "keys")

			// construct ring
			// TODO @renaynay: Include option for setting custom `userInput` parameter with
			//  implementation of https://github.com/celestiaorg/celestia-node/issues/415.
			encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)
			ring, err := keyring.New(app.Name, cfg.State.KeyringBackend, keysPath, os.Stdin, encConf.Codec)
			if err != nil {
				return err
			}

			store, err := nodebuilder.OpenStore(storePath, ring)
			if err != nil {
				return err
			}
			defer func() {
				err = errors.Join(err, store.Close())
			}()

			nd, err := nodebuilder.NewWithConfig(NodeType(ctx), Network(ctx), store, &cfg, NodeOptions(ctx)...)
			if err != nil {
				return err
			}

			ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()
			err = nd.Start(ctx)
			if err != nil {
				return err
			}

			<-ctx.Done()
			cancel() // ensure we stop reading more signals for start context

			ctx, cancel = signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			if err := nd.Stop(ctx); err != nil {
				return err
			}

			// Working with OutOrStdout/OutOrStderr allows us to unit test our command easier
			serve(rpcAddress, rpcToken, namespace)
			return nil
		},
	}

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
