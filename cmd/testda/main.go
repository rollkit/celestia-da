package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	logging "github.com/ipfs/go-log/v2"
	"github.com/rollkit/celestia-da/celestia"
	"github.com/spf13/cobra"
)

var log = logging.Logger("cmd")

var rootCmd = &cobra.Command{
	Use:  "test-da",
	Args: cobra.NoArgs,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: false,
	},
	Run: func(cmd *cobra.Command, args []string) {
		m := celestia.NewMockService()
		defer m.Close()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		go func() {
			s := <-sig
			fmt.Printf("Received signal: %v\n", s)
			os.Exit(0)
		}()
		<-sig
	},
}

func run() error {
	return rootCmd.ExecuteContext(context.Background())
}

func main() {
	err := run()
	if err != nil {
		log.Errorln("application exited with error:", err)
		os.Exit(1)
	}
}
