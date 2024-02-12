package main

import (
	cmdnode "github.com/celestiaorg/celestia-node/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	grpcAddrFlag      = "da.grpc.address"
	grpcTokenFlag     = "da.grpc.token" // #nosec G101
	grpcNamespaceFlag = "da.grpc.namespace"
	grpcListenFlag    = "da.grpc.listen"
	grpcNetworkFlag   = "da.grpc.network"
	grpcGasPriceFlag  = "da.grpc.gasprice"
)

// WithDataAvailabilityService patches the start command to also run the gRPC Data Availability service
func WithDataAvailabilityService(flags []*pflag.FlagSet) func(*cobra.Command) {
	return func(c *cobra.Command) {
		grpcFlags := &pflag.FlagSet{}
		grpcFlags.String(grpcAddrFlag, "http://127.0.0.1:26658", "celestia-node RPC endpoint address")
		grpcFlags.String(grpcTokenFlag, "", "celestia-node RPC auth token")
		grpcFlags.String(grpcNamespaceFlag, "", "celestia namespace to use (hex encoded) [Deprecated]")
		grpcFlags.String(grpcListenFlag, "127.0.0.1:0", "gRPC service listen address")
		grpcFlags.String(grpcNetworkFlag, "tcp", "gRPC service listen network type must be \"tcp\", \"tcp4\", \"tcp6\", \"unix\" or \"unixpacket\"")
		grpcFlags.Float64(grpcGasPriceFlag, -1, "gas price for estimating fee (utia/gas) default: -1 for default fees")

		fset := append(flags, grpcFlags)

		for _, set := range fset {
			c.Flags().AddFlagSet(set)
		}

		if err := c.MarkFlagRequired(grpcNamespaceFlag); err != nil {
			log.Fatal(grpcNamespaceFlag, err)
		}

		preRun := func(cmd *cobra.Command, args []string) {
			// Extract gRPC service flags
			rpcAddress, _ := cmd.Flags().GetString(grpcAddrFlag)
			rpcToken, _ := cmd.Flags().GetString(grpcTokenFlag)
			nsString, _ := cmd.Flags().GetString(grpcNamespaceFlag)
			listenAddress, _ := cmd.Flags().GetString(grpcListenFlag)
			listenNetwork, _ := cmd.Flags().GetString(grpcNetworkFlag)
			gasPrice, _ := cmd.Flags().GetFloat64(grpcGasPriceFlag)

			if rpcToken == "" {
				token, err := authToken(cmdnode.StorePath(c.Context()))
				if err != nil {
					log.Fatal(err)
				}
				rpcToken = token
			}

			// serve the gRPC service in a goroutine
			go serve(cmd.Context(), rpcAddress, rpcToken, listenAddress, listenNetwork, nsString, gasPrice)
		}

		c.PreRun = preRun
	}
}
