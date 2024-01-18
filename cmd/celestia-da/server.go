package main

import (
	"context"
	"encoding/hex"
	"errors"
	"net"
	"net/http"

	rpc "github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/celestiaorg/celestia-node/share"
	"github.com/filecoin-project/go-jsonrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/rollkit/celestia-da/celestia"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/rollkit/go-da/proxy"
)

var serviceName = "celestia-da"

func serve(ctx context.Context, rpcAddress, rpcToken, listenAddress, listenNetwork, nsString string, gasPrice float64, metrics bool) {
	var client *rpc.Client
	var err error
	var m *sdkmetric.MeterProvider
	if metrics {
		m, err = setupMetrics(ctx, serviceName)
		if err != nil {
			log.Fatalln("failed to setup metrics:", err)
		}
		httpTransport := NewMetricTransport(nil, m.Meter("rpc"), "rpc", "celestia-node json-rpc client", nil)
		httpClient := &http.Client{
			Transport: httpTransport,
		}
		client, err = rpc.NewClient(ctx, rpcAddress, rpcToken, jsonrpc.WithHTTPClient(httpClient))
		if err != nil {
			log.Fatalln("failed to create celestia-node RPC client:", err)
		}
	} else {
		client, err = rpc.NewClient(ctx, rpcAddress, rpcToken)
		if err != nil {
			log.Fatalln("failed to create celestia-node RPC client:", err)
		}
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

	da := celestia.NewCelestiaDA(client, namespace, gasPrice, ctx, m)
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
