package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"f6n/internal/aws"
	"f6n/internal/config"
	"f6n/internal/logger"
	"f6n/internal/provider"
	"f6n/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"google.golang.org/api/option"
)

func main() {
	cfg := config.Load()

	// Mirror logs to stdout when verbose/debug is requested to help during local dev or inside containers.
	if cfg.Verbose || strings.EqualFold(cfg.LogLevel, "debug") {
		logger.Logger.SetOutput(io.MultiWriter(os.Stdout, logger.Logger.Writer()))
		logger.Logger.SetPrefix("[DEBUG] ")
	}

	ctx := context.Background()

	prov, err := initProvider(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize provider: %v", err)
	}

	model := ui.NewModel(prov, cfg.Environment)
	program := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		log.Fatalf("failed to start TUI: %v", err)
	}
}

// initProvider wires up the selected cloud provider implementation.
func initProvider(ctx context.Context, cfg *config.Config) (provider.Provider, error) {
	switch strings.ToLower(cfg.Provider) {
	case "aws", "":
		lambdaClient, err := aws.NewLambdaClient(ctx, cfg.Region, cfg.Profile)
		if err != nil {
			return nil, fmt.Errorf("unable to create AWS Lambda client: %w", err)
		}

		stsClient, err := aws.NewStsClient(ctx, cfg.Region, cfg.Profile)
		if err != nil {
			return nil, fmt.Errorf("unable to create AWS STS client: %w", err)
		}

		return provider.NewAWSProvider(lambdaClient, stsClient), nil

	case "gcp":
		if strings.TrimSpace(cfg.GCPProject) == "" {
			return nil, fmt.Errorf("gcp provider selected but --gcp-project / GCP_PROJECT is not set")
		}

		return provider.NewGCPProvider(cfg.GCPProject, cfg.GCPRegion, option.WithScopes(
			"https://www.googleapis.com/auth/cloud-platform",
		))

	default:
		return nil, fmt.Errorf("unknown provider %q (expected aws or gcp)", cfg.Provider)
	}
}
