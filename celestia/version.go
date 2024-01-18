package celestia

import (
	"runtime/debug"
	"strings"
)

// BuildInfo populated in ldflags by Makefile
type BuildInfo struct {
	BuildTime       string
	LastCommit      string
	SemanticVersion string
	NodeVersion     string
	GoVersion       string
}

// ReadBuildInfo *BuildInfo containing build information
func ReadBuildInfo() *BuildInfo {
	debugInfo, _ := debug.ReadBuildInfo()
	var ldflags string
	for _, kv := range debugInfo.Settings {
		switch kv.Key {
		case "-ldflags":
			ldflags = kv.Value
		}
	}
	var buildInfo BuildInfo

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
	buildInfo.GoVersion = debugInfo.GoVersion
	return &buildInfo
}
