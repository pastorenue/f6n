package provider

import (
	"context"
)

// CloudProvider represents the cloud provider type
type CloudProvider string

const (
	AWS CloudProvider = "aws"
	GCP CloudProvider = "gcp"
)

// FunctionInfo represents generic function information across providers
type FunctionInfo struct {
	Name         string
	Runtime      string
	Memory       int32
	Timeout      int32
	Handler      string
	LastModified string
	ARN          string // AWS ARN or GCP resource name
	Description  string
	Role         string
	Environment  map[string]string
	Region       string // AWS region or GCP location
}

// Provider defines the interface for cloud function providers
type Provider interface {
	GetProviderName() CloudProvider
	GetRegion() string
	GetAccountID(ctx context.Context) (string, error)
	ListFunctions(ctx context.Context) ([]FunctionInfo, error)
	GetFunction(ctx context.Context, name string) (*FunctionInfo, error)
	GetFunctionCode(ctx context.Context, name string) (string, error)
	GetFunctionLogs(ctx context.Context, name string, limit int) ([]string, error)
	GetEndpoints(ctx context.Context, name string) ([]string, error)
}
