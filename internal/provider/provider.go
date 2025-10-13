package provider

import (
	"context"
	"time"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time
	Severity  string
	Message   string
	Labels    map[string]string
}

// MetricDataPoint represents a single metric data point
type MetricDataPoint struct {
	Timestamp time.Time
	Value     float64
}

// MetricData represents metrics for a function
type MetricData struct {
	MetricName  string
	Unit        string
	DataPoints  []MetricDataPoint
	Description string
}

// FunctionMetrics represents comprehensive metrics for a function
type FunctionMetrics struct {
	FunctionName string
	TimeRange    struct {
		Start time.Time
		End   time.Time
	}
	Invocations          MetricData
	Duration             MetricData
	Errors               MetricData
	Throttles            MetricData
	Memory               MetricData
	ConcurrentExecutions MetricData
}

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
	DownloadFunctionCode(ctx context.Context, name, destination string) error
	GetFunctionLogs(ctx context.Context, name string, limit int) ([]string, error)
	StreamFunctionLogs(ctx context.Context, name string) (<-chan LogEntry, <-chan error)
	GetFunctionMetrics(ctx context.Context, name string, startTime, endTime time.Time) (*FunctionMetrics, error)
	GetEndpoints(ctx context.Context, name string) ([]string, error)
}
