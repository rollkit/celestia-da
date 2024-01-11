package main

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

// buildInfo populated in ldflags by Makefile
type buildInfo struct {
	BuildTime       string
	LastCommit      string
	SemanticVersion string
	NodeVersion     string
}

// extractBuildInfo parses the ldflags string and returns buildInfo
func extractBuildInfo(ldflags string) *buildInfo {
	var buildInfo buildInfo

	// Split ldflags into individual key-value pairs
	keyValuePairs := strings.Split(ldflags, "-X ")

	// Iterate over key-value pairs
	for _, pair := range keyValuePairs {
		// Skip empty pairs
		if pair == "" {
			continue
		}

		// Remove quotes
		pair = strings.Trim(strings.TrimSpace(pair), "'")

		// Split pair into key and value
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			// Invalid pair, skip
			continue
		}

		// Trim leading and trailing spaces from key and value
		key := parts[0]
		value := strings.TrimSpace(parts[1])

		// Assign value to corresponding field in BuildInfo
		switch key {
		case ".buildTime":
			buildInfo.BuildTime = value
		case ".lastCommit":
			buildInfo.LastCommit = value
		case ".semanticVersion":
			buildInfo.SemanticVersion = value
		case ".nodeVersion":
			buildInfo.NodeVersion = value
		}
	}
	return &buildInfo
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show information about the current binary build",
	Args:  cobra.NoArgs,
	Run:   printBuildInfo,
}

func printBuildInfo(_ *cobra.Command, _ []string) {
	debugInfo, _ := debug.ReadBuildInfo()
	var ldflags string
	for _, kv := range debugInfo.Settings {
		switch kv.Key {
		case "-ldflags":
			ldflags = kv.Value
		}
	}
	buildInfo := extractBuildInfo(ldflags)
	fmt.Printf("Semantic version: %s\n", buildInfo.SemanticVersion)
	fmt.Printf("Build Time: %s\n", buildInfo.BuildTime)
	fmt.Printf("Last Commit: %s\n", buildInfo.LastCommit)
	fmt.Printf("Golang version: %s\n", debugInfo.GoVersion)
	fmt.Printf("Celestia Node version: %s\n", buildInfo.NodeVersion)
}
