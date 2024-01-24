package main

import (
	"fmt"

	"github.com/rollkit/celestia-da/celestia"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show information about the current binary build",
	Args:  cobra.NoArgs,
	Run:   printBuildInfo,
}

func printBuildInfo(_ *cobra.Command, _ []string) {
	buildInfo := celestia.ReadBuildInfo()
	fmt.Printf("Semantic version: %s\n", buildInfo.SemanticVersion)
	fmt.Printf("Build Time: %s\n", buildInfo.BuildTime)
	fmt.Printf("Last Commit: %s\n", buildInfo.LastCommit)
	fmt.Printf("Golang version: %s\n", buildInfo.GoVersion)
	fmt.Printf("Celestia Node version: %s\n", buildInfo.NodeVersion)
}
