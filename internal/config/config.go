package config

import (
	"flag"
	"fmt"
	"os"

	"f6n/internal/version"
)

// Config holds the application configuration
type Config struct {
	Region      string
	Environment string
	Profile     string
	LogLevel    string
	ShowVersion bool
	Provider    string // aws or gcp
	GCPProject  string // GCP project ID
	GCPRegion   string // GCP region
	Verbose     bool   // shorthand for --log-level=debug
}

// Load reads configuration from environment variables and command-line flags
func Load() *Config {
	cfg := &Config{}

	// Define command-line flags
	flag.StringVar(&cfg.Provider, "provider", "aws", "Cloud provider: aws or gcp (defaults to CLOUD_PROVIDER env var or aws)")
	flag.StringVar(&cfg.Region, "region", "", "AWS region (defaults to AWS_REGION env var or us-east-1)")
	flag.StringVar(&cfg.Environment, "env", "dev", "Environment name (defaults to STAGE env var or dev)")
	flag.StringVar(&cfg.Profile, "profile", "", "AWS profile to use (defaults to AWS_PROFILE env var)")
	flag.StringVar(&cfg.GCPProject, "gcp-project", "", "GCP project ID (defaults to GCP_PROJECT env var)")
	flag.StringVar(&cfg.GCPRegion, "gcp-region", "", "GCP region (defaults to GCP_REGION env var or us-central1)")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.BoolVar(&cfg.ShowVersion, "v", false, "Show version information (shorthand)")
	flag.BoolVar(&cfg.ShowVersion, "version", false, "Show version information")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose logging (shorthand for --log-level=debug)")
	flag.Parse()

	// Handle version flag
	if cfg.ShowVersion {
		fmt.Println(version.Info())
		os.Exit(0)
	}

	cfg.Provider = getWithEnvDefault(cfg.Provider, "CLOUD_PROVIDER", "aws")
	cfg.Region = getWithEnvDefault(cfg.Region, "AWS_REGION", "us-east-1")
	cfg.Environment = getWithEnvDefault(cfg.Environment, "STAGE", "dev")
	cfg.Profile = getWithEnvDefault(cfg.Profile, "AWS_PROFILE", "")
	cfg.GCPProject = getWithEnvDefault(cfg.GCPProject, "GCP_PROJECT", "")
	cfg.GCPRegion = getWithEnvDefault(cfg.GCPRegion, "GCP_REGION", "us-central1")

	return cfg
}

// getWithEnvDefault returns the value if not empty, otherwise checks the environment variable, otherwise returns the default
func getWithEnvDefault(value, envVar, defaultValue string) string {
	if value != "" {
		return value
	}
	if envValue := os.Getenv(envVar); envValue != "" {
		return envValue
	}
	return defaultValue
}
