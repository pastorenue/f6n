package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// StsClient wraps the AWS STS client
type StsClient struct {
	client *sts.Client
}

// NewStsClient creates a new STS client
func NewStsClient(ctx context.Context, region, profile string) (*StsClient, error) {
	var opts []func(*config.LoadOptions) error

	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &StsClient{
		client: sts.NewFromConfig(cfg),
	}, nil
}

// GetAccountID gets the AWS account ID
func (c *StsClient) GetAccountID(ctx context.Context) (string, error) {
	input := &sts.GetCallerIdentityInput{}

	result, err := c.client.GetCallerIdentity(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get caller identity: %w", err)
	}

	return *result.Account, nil
}
