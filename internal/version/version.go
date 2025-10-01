package version

import "fmt"

var (
	// Version is the current version of the application
	Version = "0.1.0"

	// GitCommit is the git commit hash (set during build)
	GitCommit = "dev"

	// BuildDate is the build date (set during build)
	BuildDate = "unknown"
)

// Info returns formatted version information
func Info() string {
	return fmt.Sprintf("f6n version %s (commit: %s, built: %s)", Version, GitCommit, BuildDate)
}

// Short returns the short version string
func Short() string {
	return Version
}
