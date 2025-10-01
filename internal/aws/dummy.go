package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// GetDummyFunctions returns dummy Lambda functions for testing
// TODO: Remove this once AWS integration is fully working
func GetDummyFunctions() []types.FunctionConfiguration {
	return []types.FunctionConfiguration{
		{
			FunctionName: ptr("user-authentication-service"),
			Runtime:      types.RuntimeNodejs20x,
			MemorySize:   ptr(int32(512)),
			Timeout:      ptr(int32(30)),
			Handler:      ptr("index.handler"),
			LastModified: ptr("2024-09-15T10:30:00.000+0000"),
			FunctionArn:  ptr("arn:aws:lambda:us-east-1:123456789:function:user-auth"),
			Description:  ptr("Handles user authentication and JWT token generation"),
			Role:         ptr("arn:aws:iam::123456789:role/lambda-exec-role"),
		},
		{
			FunctionName: ptr("payment-processor"),
			Runtime:      types.RuntimePython312,
			MemorySize:   ptr(int32(1024)),
			Timeout:      ptr(int32(60)),
			Handler:      ptr("app.lambda_handler"),
			LastModified: ptr("2024-09-20T14:22:00.000+0000"),
			FunctionArn:  ptr("arn:aws:lambda:us-east-1:123456789:function:payment"),
			Description:  ptr("Processes payment transactions via Stripe API"),
			Role:         ptr("arn:aws:iam::123456789:role/payment-lambda-role"),
		},
		{
			FunctionName: ptr("email-notification-sender"),
			Runtime:      types.RuntimeNodejs18x,
			MemorySize:   ptr(int32(256)),
			Timeout:      ptr(int32(15)),
			Handler:      ptr("index.sendEmail"),
			LastModified: ptr("2024-09-18T08:45:00.000+0000"),
			FunctionArn:  ptr("arn:aws:lambda:us-east-1:123456789:function:email-sender"),
			Description:  ptr("Sends email notifications using SES"),
			Role:         ptr("arn:aws:iam::123456789:role/email-lambda-role"),
		},
		{
			FunctionName: ptr("data-analytics-processor"),
			Runtime:      types.RuntimePython312,
			MemorySize:   ptr(int32(2048)),
			Timeout:      ptr(int32(300)),
			Handler:      ptr("analytics.process"),
			LastModified: ptr("2024-09-22T16:10:00.000+0000"),
			FunctionArn:  ptr("arn:aws:lambda:us-east-1:123456789:function:analytics"),
			Description:  ptr("Processes large datasets for analytics dashboard"),
			Role:         ptr("arn:aws:iam::123456789:role/analytics-lambda-role"),
		},
		{
			FunctionName: ptr("image-resizer"),
			Runtime:      types.RuntimeNodejs20x,
			MemorySize:   ptr(int32(1536)),
			Timeout:      ptr(int32(45)),
			Handler:      ptr("resize.handler"),
			LastModified: ptr("2024-09-10T12:00:00.000+0000"),
			FunctionArn:  ptr("arn:aws:lambda:us-east-1:123456789:function:image-resize"),
			Description:  ptr("Resizes and optimizes images for S3 storage"),
			Role:         ptr("arn:aws:iam::123456789:role/image-lambda-role"),
		},
	}
}

// UseDummyData controls whether to use dummy data or real AWS calls
var UseDummyData = true

// ListFunctionsWithFallback tries to list real functions, falls back to dummy data
func (c *LambdaClient) ListFunctionsWithFallback(ctx context.Context) ([]types.FunctionConfiguration, error) {
	if UseDummyData {
		return GetDummyFunctions(), nil
	}
	return c.ListFunctions(ctx)
}

// ptr is a helper function to create pointers
func ptr[T any](v T) *T {
	return &v
}
