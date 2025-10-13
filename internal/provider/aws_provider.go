package provider

import (
	"context"
	"fmt"
	"time"

	"f6n/internal/aws"

	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// AWSProvider implements the Provider interface for AWS Lambda
type AWSProvider struct {
	client    *aws.LambdaClient
	stsClient *aws.StsClient
}

// NewAWSProvider creates a new AWS provider
func NewAWSProvider(client *aws.LambdaClient, stsClient *aws.StsClient) *AWSProvider {
	return &AWSProvider{
		client:    client,
		stsClient: stsClient,
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

func (p *AWSProvider) GetAccountID(ctx context.Context) (string, error) {
	return p.stsClient.GetAccountID(ctx)
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

// StreamFunctionLogs streams logs for a function in real-time (placeholder)
func (p *AWSProvider) StreamFunctionLogs(ctx context.Context, functionName string) (<-chan LogEntry, <-chan error) {
	logChan := make(chan LogEntry, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(logChan)
		defer close(errChan)

		// Send a placeholder message
		logChan <- LogEntry{
			Timestamp: time.Now(),
			Severity:  "INFO",
			Message:   "CloudWatch Logs streaming coming soon... Use AWS Console for real-time logs",
			Labels:    map[string]string{"function": functionName},
		}
	}()

	return logChan, errChan
}

// GetFunctionMetrics retrieves metrics for a Lambda function (placeholder)
func (p *AWSProvider) GetFunctionMetrics(ctx context.Context, functionName string, startTime, endTime time.Time) (*FunctionMetrics, error) {
	// TODO: Implement CloudWatch metrics integration

	// Create sample metrics data for now
	metrics := &FunctionMetrics{
		FunctionName: functionName,
		TimeRange: struct {
			Start time.Time
			End   time.Time
		}{Start: startTime, End: endTime},
	}

	// Generate sample data points
	now := time.Now()
	samplePoints := []MetricDataPoint{
		{Timestamp: now.Add(-1 * time.Hour), Value: 10},
		{Timestamp: now.Add(-45 * time.Minute), Value: 15},
		{Timestamp: now.Add(-30 * time.Minute), Value: 8},
		{Timestamp: now.Add(-15 * time.Minute), Value: 12},
		{Timestamp: now, Value: 6},
	}

	metrics.Invocations = MetricData{
		MetricName:  "Invocations",
		Unit:        "count",
		Description: "Number of function invocations (sample data)",
		DataPoints:  samplePoints,
	}

	durationPoints := []MetricDataPoint{
		{Timestamp: now.Add(-1 * time.Hour), Value: 250.5},
		{Timestamp: now.Add(-45 * time.Minute), Value: 180.2},
		{Timestamp: now.Add(-30 * time.Minute), Value: 320.8},
		{Timestamp: now.Add(-15 * time.Minute), Value: 195.3},
		{Timestamp: now, Value: 275.1},
	}

	metrics.Duration = MetricData{
		MetricName:  "Duration",
		Unit:        "ms",
		Description: "Average function execution duration (sample data)",
		DataPoints:  durationPoints,
	}

	return metrics, nil
}

// GetEndpoints gets API Gateway endpoints associated with a function (placeholder)
func (p *AWSProvider) GetEndpoints(ctx context.Context, name string) ([]string, error) {
	// TODO: Implement API Gateway integration
	return []string{
		"API Gateway integration coming soon...",
		"Use AWS Console to view endpoints for now",
	}, nil
}

// DownloadFunctionCode downloads the function code to a local path (placeholder)
func (p *AWSProvider) DownloadFunctionCode(ctx context.Context, name, destination string) error {
	// Add logging to track AWS download attempts
	fmt.Printf("AWS DownloadFunctionCode called - function: %s, destination: %s\n", name, destination)
	// TODO: Implement AWS Lambda function code download
	// This could use the GetFunction API with Code location to download from S3
	return fmt.Errorf("AWS Lambda code download not yet implemented")
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
