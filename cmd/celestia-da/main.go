package main

import (
	"errors"
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
	//  - options to configure gRPC server (encryption, any other)
	//  - options for listener (network, unix sokcet?)

	// TODO(tzdybal): replace with a constructor
	da := &celestia.CelestiaDA{}
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
