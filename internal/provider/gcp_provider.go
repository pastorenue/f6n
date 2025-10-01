package provider

import (
	"context"
	"f6n/internal/logger"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/logging/logadmin"
	"google.golang.org/api/cloudfunctions/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GCPProvider implements the Provider interface for GCP Cloud Functions
type GCPProvider struct {
	projectID string
	region    string
	client    *cloudfunctions.Service
}

// NewGCPProvider creates a new GCP provider
func NewGCPProvider(projectID, region string, opts ...option.ClientOption) (*GCPProvider, error) {
	if region == "" {
		region = "us-central1"
	}

	ctx := context.Background()
	client, err := cloudfunctions.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Functions client: %w", err)
	}

	return &GCPProvider{
		projectID: projectID,
		region:    region,
		client:    client,
	}, nil
}

// GetProviderName returns "gcp"
func (p *GCPProvider) GetProviderName() CloudProvider {
	return GCP
}

// GetRegion returns the GCP location
func (p *GCPProvider) GetRegion() string {
	return p.region
}

func (p *GCPProvider) GetAccountID(ctx context.Context) (string, error) {
	return p.projectID, nil
}

// ListFunctions lists all Cloud Functions (using dummy data for now)
func (p *GCPProvider) ListFunctions(ctx context.Context) ([]FunctionInfo, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", p.projectID, p.region)
	resp, err := p.client.Projects.Locations.Functions.List(parent).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list Cloud Functions: %w", err)
	}

	var functions []FunctionInfo
	for _, f := range resp.Functions {
		// UpdateTime is in RFC3339 format
		lastModified, err := time.Parse(time.RFC3339, f.UpdateTime)
		if err != nil {
			lastModified = time.Time{}
		}

			timeout, err := time.ParseDuration(f.Timeout)
			if err != nil {
				timeout = 0
			}

			functions = append(functions, FunctionInfo{
				Name:         f.Name[strings.LastIndex(f.Name, "/")+1:],
				Runtime:      f.Runtime,
				Memory:       int32(f.AvailableMemoryMb),
				Timeout:      int32(timeout.Seconds()),
				LastModified: lastModified.Format("2006-01-02 15:04:05"),
				ARN:          f.Name,
				Description:  f.Description,
				Region:       p.region,
			})
		}

		return functions, nil
}

// GetFunction gets details about a specific function
func (p *GCPProvider) GetFunction(ctx context.Context, name string) (*FunctionInfo, error) {
	// TODO: Implement real GCP Cloud Functions API integration
	functions, err := p.ListFunctions(ctx)
	if err != nil {
		return nil, err
	}

	for _, fn := range functions {
		if fn.Name == name {
			return &fn, nil
		}
	}

	return nil, fmt.Errorf("function %s not found", name)
}

// GetFunctionCode gets the code/source for a function
func (p *GCPProvider) GetFunctionCode(ctx context.Context, name string) (string, error) {
	logger.Logger.Printf("Getting function code info for: %s", name)
	
	// Get the function details
	fullName := fmt.Sprintf("projects/%s/locations/%s/functions/%s", p.projectID, p.region, name)
	function, err := p.client.Projects.Locations.Functions.Get(fullName).Do()
	if err != nil {
		logger.Logger.Printf("Error getting function details: %v", err)
		return "", fmt.Errorf("failed to get function details: %w", err)
	}

	logger.Logger.Printf("Function retrieved: %s", function.Name)

	var info strings.Builder
	info.WriteString("━━━ Code Information ━━━\n\n")
	
	info.WriteString(fmt.Sprintf("Runtime: %s\n", function.Runtime))
	info.WriteString(fmt.Sprintf("Entry Point: %s\n\n", function.EntryPoint))
	
	// Source code location
	if function.SourceArchiveUrl != "" {
		info.WriteString("Source Type: Cloud Storage Archive\n")
		info.WriteString(fmt.Sprintf("Archive URL: %s\n\n", function.SourceArchiveUrl))
		
		// Parse bucket and object
		gsURL := function.SourceArchiveUrl
		if strings.HasPrefix(gsURL, "gs://") {
			parts := strings.Split(strings.TrimPrefix(gsURL, "gs://"), "/")
			if len(parts) >= 2 {
				bucket := parts[0]
				object := strings.Join(parts[1:], "/")
				info.WriteString(fmt.Sprintf("Bucket: %s\n", bucket))
				info.WriteString(fmt.Sprintf("Object: %s\n\n", object))
			}
		}
	} else if function.SourceRepository != nil {
		info.WriteString("Source Type: Cloud Source Repository\n")
		info.WriteString(fmt.Sprintf("Repository URL: %s\n\n", function.SourceRepository.Url))
		
		if function.SourceRepository.DeployedUrl != "" {
			info.WriteString(fmt.Sprintf("Deployed URL: %s\n\n", function.SourceRepository.DeployedUrl))
		}
	} else if function.SourceUploadUrl != "" {
		info.WriteString("Source Type: Upload URL\n")
		info.WriteString(fmt.Sprintf("Upload URL: %s\n\n", function.SourceUploadUrl))
	} else {
		info.WriteString("Source: Not available\n\n")
	}
	
	// Service configuration
	info.WriteString("Configuration:\n")
	info.WriteString(fmt.Sprintf("  Memory: %d MB\n", function.AvailableMemoryMb))
	info.WriteString(fmt.Sprintf("  Timeout: %ds\n", function.Timeout))
	if function.MaxInstances > 0 {
		info.WriteString(fmt.Sprintf("  Max Instances: %d\n", function.MaxInstances))
	}
	if function.MinInstances > 0 {
		info.WriteString(fmt.Sprintf("  Min Instances: %d\n", function.MinInstances))
	}
	info.WriteString("\n")
	
	// Environment variables
	if len(function.EnvironmentVariables) > 0 {
		info.WriteString("Environment Variables:\n")
		for k, v := range function.EnvironmentVariables {
			info.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
		info.WriteString("\n")
	}
	
	// VPC connector if available
	if function.VpcConnector != "" {
		info.WriteString(fmt.Sprintf("VPC Connector: %s\n\n", function.VpcConnector))
	}
	
	info.WriteString("To download source code:\n")
	if function.SourceArchiveUrl != "" {
		info.WriteString(fmt.Sprintf("1. Use gsutil: gsutil cp %s .\n", function.SourceArchiveUrl))
	}
	info.WriteString("2. Use gcloud CLI: gcloud functions describe " + name + " --region=" + p.region + "\n")
	info.WriteString("3. Download from GCP Console > Cloud Functions\n")
	if function.SourceRepository != nil {
		info.WriteString("4. Clone from Cloud Source Repository using the URL above\n")
	}
	
	logger.Logger.Printf("Successfully retrieved code information")
	return info.String(), nil
}

// GetFunctionLogs gets logs for a function
func (p *GCPProvider) GetFunctionLogs(ctx context.Context, functionName string, limit int) ([]string, error) {
	// Create logging client
	adminClient, err := logadmin.NewClient(ctx, p.projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create logging client: %w", err)
	}
	defer adminClient.Close()

	// Build filter for Cloud Function logs
	// Cloud Functions log to resources.type="cloud_function" with function_name label
	filter := fmt.Sprintf(`resource.type="cloud_function"
resource.labels.function_name="%s"
timestamp>="%s"`, 
		functionName,
		time.Now().Add(-24*time.Hour).Format(time.RFC3339), // Last 24 hours
	)

	// Query logs
	iter := adminClient.Entries(ctx,
		logadmin.Filter(filter),
		logadmin.NewestFirst(),
	)

	var logs []string
	count := 0
	
	for count < limit {
		entry, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch log entry: %w", err)
		}

		// Format log entry
		timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
		severity := entry.Severity.String()
		
		var message string
		switch payload := entry.Payload.(type) {
		case string:
			message = payload
		default:
			message = fmt.Sprintf("%v", payload)
		}

		logLine := fmt.Sprintf("[%s] %s: %s", timestamp, severity, message)
		logs = append(logs, logLine)
		count++
	}

	if len(logs) == 0 {
		return []string{fmt.Sprintf("No logs found for function: %s (last 24 hours)", functionName)}, nil
	}

	return logs, nil
}

// GetEndpoints gets endpoints associated with a function
func (p *GCPProvider) GetEndpoints(ctx context.Context, name string) ([]string, error) {
	// TODO: Implement real endpoint discovery
	return []string{
		fmt.Sprintf("https://%s-%s.cloudfunctions.net/%s", p.region, p.projectID, name),
		"Note: This is a generated URL. Verify in GCP Console for actual trigger URLs.",
	}, nil
}
