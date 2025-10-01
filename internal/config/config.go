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
	GCPLocation string // GCP location/region
}

// Load reads configuration from environment variables and command-line flags
func Load() *Config {
	cfg := &Config{}

	// Define command-line flags
	flag.StringVar(&cfg.Provider, "provider", "", "Cloud provider: aws or gcp (defaults to CLOUD_PROVIDER env var or aws)")
	flag.StringVar(&cfg.Region, "region", "", "AWS region (defaults to AWS_REGION env var or us-east-1)")
	flag.StringVar(&cfg.Environment, "env", "", "Environment name (defaults to STAGE env var or dev)")
	flag.StringVar(&cfg.Profile, "profile", "", "AWS profile to use (defaults to AWS_PROFILE env var)")
	flag.StringVar(&cfg.GCPProject, "gcp-project", "", "GCP project ID (defaults to GCP_PROJECT env var)")
	flag.StringVar(&cfg.GCPLocation, "gcp-location", "", "GCP location (defaults to GCP_LOCATION env var or us-central1)")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.BoolVar(&cfg.ShowVersion, "version", false, "Show version information")
	flag.BoolVar(&cfg.ShowVersion, "v", false, "Show version information (shorthand)")
	flag.Parse()

	// Handle version flag
	if cfg.ShowVersion {
		fmt.Println(version.Info())
		os.Exit(0)
	}

	// Apply defaults from environment variables
	if cfg.Provider == "" {
		cfg.Provider = os.Getenv("CLOUD_PROVIDER")
		if cfg.Provider == "" {
			cfg.Provider = "aws"
		}
	}
	
	if cfg.Region == "" {
		cfg.Region = os.Getenv("AWS_REGION")
		if cfg.Region == "" {
			cfg.Region = "us-east-1"
		}
	}

	if cfg.Environment == "" {
		cfg.Environment = os.Getenv("STAGE")
		if cfg.Environment == "" {
			cfg.Environment = "dev"
		}
	}

	if cfg.Profile == "" {
		cfg.Profile = os.Getenv("AWS_PROFILE")
	}
	
	if cfg.GCPProject == "" {
		cfg.GCPProject = os.Getenv("GCP_PROJECT")
	}
	
	if cfg.GCPLocation == "" {
		cfg.GCPLocation = os.Getenv("GCP_LOCATION")
		if cfg.GCPLocation == "" {
			cfg.GCPLocation = "us-central1"
		}
	}

	return cfg
}
