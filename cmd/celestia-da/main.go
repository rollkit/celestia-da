package main

import (
	"context"
	"encoding/hex"
	"errors"
	rpc "github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/celestiaorg/celestia-node/share"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"os"

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
		Run: func(cmd *cobra.Command, args []string) {
			// Working with OutOrStdout/OutOrStderr allows us to unit test our command easier
			serve(rpcAddress, rpcToken, namespace)
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
	log.Println("serving celestia-da over gRPC on:", lis.Addr())
	err = srv.Serve(lis)
	if !errors.Is(err, grpc.ErrServerStopped) {
		log.Fatalln("gRPC server stopped with error:", err)
	}
}
