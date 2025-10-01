package provider

import (
	"context"
	"fmt"

	"f6n/internal/aws"

	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// AWSProvider implements the Provider interface for AWS Lambda
type AWSProvider struct {
	client *aws.LambdaClient
}

// NewAWSProvider creates a new AWS provider
func NewAWSProvider(client *aws.LambdaClient) *AWSProvider {
	return &AWSProvider{
		client: client,
	}
}

// GetProviderName returns "aws"
func (p *AWSProvider) GetProviderName() CloudProvider {
	return AWS
}

// GetRegion returns the AWS region
func (p *AWSProvider) GetRegion() string {
	return p.client.Region()
}

// ListFunctions lists all Lambda functions
func (p *AWSProvider) ListFunctions(ctx context.Context) ([]FunctionInfo, error) {
	functions, err := p.client.ListFunctionsWithFallback(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]FunctionInfo, 0, len(functions))
	for _, fn := range functions {
		result = append(result, convertAWSFunction(fn, p.client.Region()))
	}

	return result, nil
}

// GetFunction gets details about a specific function
func (p *AWSProvider) GetFunction(ctx context.Context, name string) (*FunctionInfo, error) {
	output, err := p.client.GetFunctionConfiguration(ctx, name)
	if err != nil {
		return nil, err
	}

	info := &FunctionInfo{
		Name:         getString(output.FunctionName),
		Runtime:      string(output.Runtime),
		Memory:       getInt32(output.MemorySize),
		Timeout:      getInt32(output.Timeout),
		Handler:      getString(output.Handler),
		LastModified: getString(output.LastModified),
		ARN:          getString(output.FunctionArn),
		Description:  getString(output.Description),
		Role:         getString(output.Role),
		Region:       p.client.Region(),
	}

	if output.Environment != nil {
		info.Environment = output.Environment.Variables
	}

	return info, nil
}

// GetFunctionCode gets the code/source for a function
func (p *AWSProvider) GetFunctionCode(ctx context.Context, name string) (string, error) {
	output, err := p.client.GetFunction(ctx, name)
	if err != nil {
		return "", err
	}

	if output.Code != nil && output.Code.Location != nil {
		return fmt.Sprintf("Code location: %s\\n\\nNote: Download the code from the S3 location above to view it.", *output.Code.Location), nil
	}

	return "Code location not available", nil
}

// GetFunctionLogs gets logs for a function (placeholder)
func (p *AWSProvider) GetFunctionLogs(ctx context.Context, name string, limit int) ([]string, error) {
	// TODO: Implement CloudWatch Logs integration
	return []string{
		"CloudWatch Logs integration coming soon...",
		fmt.Sprintf("Function: %s", name),
		"Use AWS Console or AWS CLI to view logs for now",
	}, nil
}

// GetEndpoints gets API Gateway endpoints associated with a function (placeholder)
func (p *AWSProvider) GetEndpoints(ctx context.Context, name string) ([]string, error) {
	// TODO: Implement API Gateway integration
	return []string{
		"API Gateway integration coming soon...",
		"Use AWS Console to view endpoints for now",
	}, nil
}

// Helper functions

func convertAWSFunction(fn awstypes.FunctionConfiguration, region string) FunctionInfo {
	info := FunctionInfo{
		Name:         getString(fn.FunctionName),
		Runtime:      string(fn.Runtime),
		Memory:       getInt32(fn.MemorySize),
		Timeout:      getInt32(fn.Timeout),
		Handler:      getString(fn.Handler),
		LastModified: getString(fn.LastModified),
		ARN:          getString(fn.FunctionArn),
		Description:  getString(fn.Description),
		Role:         getString(fn.Role),
		Region:       region,
	}

	if fn.Environment != nil {
		info.Environment = fn.Environment.Variables
	}

	return info
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}
