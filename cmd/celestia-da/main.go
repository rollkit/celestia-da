package main

import (
	"context"
	"errors"
	rpc "github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/celestiaorg/celestia-node/share"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"

	"github.com/rollkit/celestia-da"
	"github.com/rollkit/go-da/proxy"
)

func main() {
	// TODO: potential configuration options:
	//  - options to configure Celestia Client
	//    - address
	//    - token
	//    - namespace
	//    - height (?)
	//  - options to configure gRPC server (encryption, any other)
	//  - options for listener (network, unix sokcet?)

	rpcAddress := "127.0.0.1:"
	rpcToken := ""
	rpcNamespace := []byte("test")

	ctx := context.Background()
	client, err := rpc.NewClient(ctx, rpcAddress, rpcToken)
	if err != nil {
		log.Fatalln("failed to create celestia-node RPC client:", err)
	}
	namespace, err := share.NewBlobNamespaceV0(rpcNamespace)
	if err != nil {
		log.Fatalln("invalid namespace:", err)
	}

	da := celestia.NewCelestiaDA(client, namespace, ctx)
	srv := proxy.NewServer(da, grpc.Creds(insecure.NewCredentials()))

	lis, err := net.Listen("tcp", "")
	if err != nil {
		log.Fatalln("failed to create network listener:", err)
	}
	err = srv.Serve(lis)
	if !errors.Is(err, grpc.ErrServerStopped) {
		log.Fatalln("gRPC server stopped with error:", err)
	}
}
