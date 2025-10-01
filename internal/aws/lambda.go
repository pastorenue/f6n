package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// LambdaClient wraps the AWS Lambda client with our custom methods
type LambdaClient struct {
	client *lambda.Client
	region string
}

// NewLambdaClient creates a new Lambda client for the specified region
func NewLambdaClient(ctx context.Context, region, profile string) (*LambdaClient, error) {
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

	return &LambdaClient{
		client: lambda.NewFromConfig(cfg),
		region: cfg.Region,
	}, nil
}

// ListFunctions retrieves all Lambda functions in the region
func (c *LambdaClient) ListFunctions(ctx context.Context) ([]types.FunctionConfiguration, error) {
	var functions []types.FunctionConfiguration
	var marker *string

	for {
		input := &lambda.ListFunctionsInput{
			Marker: marker,
		}

		result, err := c.client.ListFunctions(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list functions: %w", err)
		}

		functions = append(functions, result.Functions...)

		if result.NextMarker == nil {
			break
		}
		marker = result.NextMarker
	}

	return functions, nil
}

// GetFunction retrieves detailed information about a specific function
func (c *LambdaClient) GetFunction(ctx context.Context, functionName string) (*lambda.GetFunctionOutput, error) {
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	}

	result, err := c.client.GetFunction(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get function %s: %w", functionName, err)
	}

	return result, nil
}

// GetFunctionConfiguration retrieves configuration for a specific function
func (c *LambdaClient) GetFunctionConfiguration(ctx context.Context, functionName string) (*lambda.GetFunctionConfigurationOutput, error) {
	input := &lambda.GetFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
	}

	result, err := c.client.GetFunctionConfiguration(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get function configuration %s: %w", functionName, err)
	}

	return result, nil
}

// Region returns the AWS region this client is configured for
func (c *LambdaClient) Region() string {
	return c.region
}
