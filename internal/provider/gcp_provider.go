package provider

import (
	"archive/zip"
	"context"
	"f6n/internal/logger"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/logging/logadmin"
	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"cloud.google.com/go/storage"
	"google.golang.org/api/cloudfunctions/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GCPProvider implements the Provider interface for GCP Cloud Functions
type GCPProvider struct {
	projectID  string
	region     string
	client     *cloudfunctions.Service
	clientOpts []option.ClientOption
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
		projectID:  projectID,
		region:     region,
		client:     client,
		clientOpts: opts,
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

	writeSourceInfo(&info, function)
	writeConfigInfo(&info, function)
	writeEnvVars(&info, function)
	writeVPCInfo(&info, function)
	writeDownloadInstructions(&info, function, name, p.region)

	logger.Logger.Printf("Successfully retrieved code information")
	return info.String(), nil
}

// DownloadFunctionCode downloads the function code to a local path
func (p *GCPProvider) DownloadFunctionCode(ctx context.Context, name, destination string) error {
	logger.Logger.Printf("DownloadFunctionCode called - function: %s, destination: %s", name, destination)

	fullName := fmt.Sprintf("projects/%s/locations/%s/functions/%s", p.projectID, p.region, name)
	logger.Logger.Printf("Getting function details from GCP: %s", fullName)

	function, err := p.client.Projects.Locations.Functions.Get(fullName).Do()
	if err != nil {
		logger.Logger.Printf("Error getting function details from GCP: %v", err)
		return fmt.Errorf("failed to get function details: %w", err)
	}

	logger.Logger.Printf("Function retrieved successfully: %s", function.Name)
	logger.Logger.Printf("Function source details:")
	logger.Logger.Printf("  - SourceArchiveUrl: %s", function.SourceArchiveUrl)
	if function.SourceRepository != nil {
		logger.Logger.Printf("  - SourceRepository.Url: %s", function.SourceRepository.Url)
		logger.Logger.Printf("  - SourceRepository.DeployedUrl: %s", function.SourceRepository.DeployedUrl)
	} else {
		logger.Logger.Printf("  - SourceRepository: nil")
	}
	logger.Logger.Printf("  - SourceUploadUrl: %s", function.SourceUploadUrl)

	// Create destination directory if it doesn't exist
	logger.Logger.Printf("Creating destination directory: %s", destination)
	if err := os.MkdirAll(destination, 0755); err != nil {
		logger.Logger.Printf("Error creating destination directory: %v", err)
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if function.SourceArchiveUrl != "" {
		logger.Logger.Printf("Found SourceArchiveUrl: %s", function.SourceArchiveUrl)
		// Download from Cloud Storage
		return p.downloadFromGCS(ctx, function.SourceArchiveUrl, destination)
	} else if function.SourceRepository != nil {
		logger.Logger.Printf("Found SourceRepository: %s", function.SourceRepository.Url)
		// For source repositories, we can't directly download but provide instructions
		instructionsFile := filepath.Join(destination, "clone_instructions.txt")
		instructions := fmt.Sprintf(`To clone this Cloud Source Repository:

Repository URL: %s
`, function.SourceRepository.Url)

		if function.SourceRepository.DeployedUrl != "" {
			instructions += fmt.Sprintf("Deployed URL: %s\n", function.SourceRepository.DeployedUrl)
		}

		instructions += `
Commands to clone:
1. Install Google Cloud SDK if not already installed
2. Authenticate: gcloud auth login
3. Clone: gcloud source repos clone [REPO_NAME] --project=` + p.projectID

		logger.Logger.Printf("Writing clone instructions to: %s", instructionsFile)
		return os.WriteFile(instructionsFile, []byte(instructions), 0644)
	} else if function.SourceUploadUrl != "" {
		logger.Logger.Printf("Found SourceUploadUrl but not supported: %s", function.SourceUploadUrl)
		return fmt.Errorf("source upload URL type not supported for direct download")
	}

	logger.Logger.Printf("No downloadable source found for function %s", name)
	return fmt.Errorf("no downloadable source found for function %s", name)
}

// downloadFromGCS downloads and extracts a ZIP file from Google Cloud Storage
func (p *GCPProvider) downloadFromGCS(ctx context.Context, gsURL, destination string) error {
	logger.Logger.Printf("downloadFromGCS called with URL: %s, destination: %s", gsURL, destination)

	if !strings.HasPrefix(gsURL, "gs://") {
		logger.Logger.Printf("Invalid GCS URL - doesn't start with gs://: %s", gsURL)
		return fmt.Errorf("invalid GCS URL: %s", gsURL)
	}

	// Parse the GCS URL
	urlParts := strings.TrimPrefix(gsURL, "gs://")
	parts := strings.SplitN(urlParts, "/", 2)
	if len(parts) != 2 {
		logger.Logger.Printf("Invalid GCS URL format - couldn't parse bucket/object: %s", gsURL)
		return fmt.Errorf("invalid GCS URL format: %s", gsURL)
	}

	bucket := parts[0]
	object := parts[1]

	logger.Logger.Printf("Downloading from GCS bucket: %s, object: %s", bucket, object)

	// Create Cloud Storage client
	logger.Logger.Printf("Creating GCS client with authentication options...")
	client, err := storage.NewClient(ctx, p.clientOpts...)
	if err != nil {
		logger.Logger.Printf("Failed to create GCS client: %v", err)
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	// Download the file
	logger.Logger.Printf("Creating object reader for bucket: %s, object: %s", bucket, object)
	reader, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		logger.Logger.Printf("Failed to create object reader: %v", err)
		return fmt.Errorf("failed to create object reader: %w", err)
	}
	defer reader.Close()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destination, 0755); err != nil {
		logger.Logger.Printf("Failed to create destination directory %s: %v", destination, err)
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create temporary file for the ZIP
	tempFile := filepath.Join(destination, "source.zip")
	logger.Logger.Printf("Creating temporary file: %s", tempFile)
	outFile, err := os.Create(tempFile)
	if err != nil {
		logger.Logger.Printf("Failed to create temp file %s: %v", tempFile, err)
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer outFile.Close()

	// Copy the content
	logger.Logger.Printf("Copying content from GCS object to local file...")
	bytesWritten, err := io.Copy(outFile, reader)
	if err != nil {
		logger.Logger.Printf("Failed to download file: %v", err)
		return fmt.Errorf("failed to download file: %w", err)
	}

	logger.Logger.Printf("ZIP file downloaded successfully to: %s (%d bytes)", tempFile, bytesWritten)

	// Extract the ZIP file
	if err := p.extractZip(tempFile, destination); err != nil {
		return fmt.Errorf("failed to extract ZIP: %w", err)
	}

	// Remove the ZIP file after extraction
	os.Remove(tempFile)

	logger.Logger.Printf("Function code successfully downloaded and extracted to: %s", destination)
	return nil
}

// extractZip extracts a ZIP file to the specified destination
func (p *GCPProvider) extractZip(src, dest string) error {
	logger.Logger.Printf("Extracting ZIP file from %s to %s", src, dest)

	reader, err := zip.OpenReader(src)
	if err != nil {
		logger.Logger.Printf("Failed to open ZIP file %s: %v", src, err)
		return err
	}
	defer reader.Close()

	logger.Logger.Printf("ZIP file opened successfully, found %d files", len(reader.File))

	// Extract files
	for i, file := range reader.File {
		logger.Logger.Printf("Extracting file %d/%d: %s", i+1, len(reader.File), file.Name)
		path := filepath.Join(dest, file.Name)

		// Check that the file path is within destination (security check)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.FileInfo().Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", path, err)
			}
			continue
		}

		// Create directories if needed
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// Extract file
		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		_, err = io.Copy(targetFile, fileReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeSourceInfo(info *strings.Builder, function *cloudfunctions.CloudFunction) {
	if function.SourceArchiveUrl != "" {
		info.WriteString("Source Type: Cloud Storage Archive\n")
		info.WriteString(fmt.Sprintf("Archive URL: %s\n\n", function.SourceArchiveUrl))
		writeBucketObjectInfo(info, function.SourceArchiveUrl)
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
}

func writeBucketObjectInfo(info *strings.Builder, gsURL string) {
	if strings.HasPrefix(gsURL, "gs://") {
		parts := strings.Split(strings.TrimPrefix(gsURL, "gs://"), "/")
		if len(parts) >= 2 {
			bucket := parts[0]
			object := strings.Join(parts[1:], "/")
			info.WriteString(fmt.Sprintf("Bucket: %s\n", bucket))
			info.WriteString(fmt.Sprintf("Object: %s\n\n", object))
		}
	}
}

func writeConfigInfo(info *strings.Builder, function *cloudfunctions.CloudFunction) {
	info.WriteString("Configuration:\n")
	info.WriteString(fmt.Sprintf("  Memory: %d MB\n", function.AvailableMemoryMb))
	info.WriteString(fmt.Sprintf("  Timeout: %s\n", function.Timeout))
	if function.MaxInstances > 0 {
		info.WriteString(fmt.Sprintf("  Max Instances: %d\n", function.MaxInstances))
	}
	if function.MinInstances > 0 {
		info.WriteString(fmt.Sprintf("  Min Instances: %d\n", function.MinInstances))
	}
	info.WriteString("\n")
}

func writeEnvVars(info *strings.Builder, function *cloudfunctions.CloudFunction) {
	if len(function.EnvironmentVariables) > 0 {
		info.WriteString("Environment Variables:\n")
		for k, v := range function.EnvironmentVariables {
			info.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
		info.WriteString("\n")
	}
}

func writeVPCInfo(info *strings.Builder, function *cloudfunctions.CloudFunction) {
	if function.VpcConnector != "" {
		info.WriteString(fmt.Sprintf("VPC Connector: %s\n\n", function.VpcConnector))
	}
}

func writeDownloadInstructions(info *strings.Builder, function *cloudfunctions.CloudFunction, name, region string) {
	info.WriteString("To download source code:\n")
	if function.SourceArchiveUrl != "" {
		info.WriteString(fmt.Sprintf("1. Use gsutil: gsutil cp %s .\n", function.SourceArchiveUrl))
	}
	info.WriteString("2. Use gcloud CLI: gcloud functions describe " + name + " --region=" + region + "\n")
	info.WriteString("3. Download from GCP Console > Cloud Functions\n")
	if function.SourceRepository != nil {
		info.WriteString("4. Clone from Cloud Source Repository using the URL above\n")
	}
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

// StreamFunctionLogs streams logs for a function in real-time
func (p *GCPProvider) StreamFunctionLogs(ctx context.Context, functionName string) (<-chan LogEntry, <-chan error) {
	logChan := make(chan LogEntry, 100) // Buffer to prevent blocking
	errChan := make(chan error, 1)

	go func() {
		defer close(logChan)
		defer close(errChan)

		logger.Logger.Printf("Starting log streaming for function: %s", functionName)

		// Create logging client
		adminClient, err := logadmin.NewClient(ctx, p.projectID, p.clientOpts...)
		if err != nil {
			errChan <- fmt.Errorf("failed to create logging client: %w", err)
			return
		}
		defer adminClient.Close()

		// Track the last seen timestamp to avoid duplicates
		lastTimestamp := time.Now().Add(-1 * time.Minute) // Start from 1 minute ago
		ticker := time.NewTicker(2 * time.Second)         // Poll every 2 seconds
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Logger.Printf("Log streaming cancelled for function: %s", functionName)
				return
			case <-ticker.C:
				// Build filter for Cloud Function logs since last timestamp
				filter := fmt.Sprintf(`resource.type="cloud_function"
resource.labels.function_name="%s"
timestamp>="%s"`,
					functionName,
					lastTimestamp.Format(time.RFC3339),
				)

				// Query logs
				iter := adminClient.Entries(ctx,
					logadmin.Filter(filter),
					logadmin.NewestFirst(),
				)

				var newEntries []LogEntry
				for {
					entry, err := iter.Next()
					if err == iterator.Done {
						break
					}
					if err != nil {
						logger.Logger.Printf("Error fetching log entry: %v", err)
						continue
					}

					// Skip if we've already seen this timestamp or older
					if !entry.Timestamp.After(lastTimestamp) {
						continue
					}

					// Create LogEntry
					var message string
					switch payload := entry.Payload.(type) {
					case string:
						message = payload
					default:
						message = fmt.Sprintf("%v", payload)
					}

					logEntry := LogEntry{
						Timestamp: entry.Timestamp,
						Severity:  entry.Severity.String(),
						Message:   message,
						Labels:    entry.Labels,
					}

					newEntries = append(newEntries, logEntry)
				}

				// Send new entries in chronological order (oldest first)
				for i := len(newEntries) - 1; i >= 0; i-- {
					select {
					case logChan <- newEntries[i]:
						// Update last timestamp
						if newEntries[i].Timestamp.After(lastTimestamp) {
							lastTimestamp = newEntries[i].Timestamp
						}
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return logChan, errChan
}

// GetFunctionMetrics retrieves metrics for a Cloud Function
func (p *GCPProvider) GetFunctionMetrics(ctx context.Context, functionName string, startTime, endTime time.Time) (*FunctionMetrics, error) {
	logger.Logger.Printf("Fetching metrics for function: %s", functionName)

	// Create monitoring client
	client, err := monitoring.NewMetricClient(ctx, p.clientOpts...)
	if err != nil {
		logger.Logger.Printf("Failed to create monitoring client: %v, falling back to sample data", err)
		return p.generateSampleMetrics(functionName, startTime, endTime), nil
	}
	defer client.Close()

	metrics := &FunctionMetrics{
		FunctionName: functionName,
		TimeRange: struct {
			Start time.Time
			End   time.Time
		}{Start: startTime, End: endTime},
	}

	// Define the metrics we want to fetch
	metricTypes := map[string]*MetricData{
		"cloudfunctions.googleapis.com/function/executions":        &metrics.Invocations,
		"cloudfunctions.googleapis.com/function/execution_times":   &metrics.Duration,
		"cloudfunctions.googleapis.com/function/user_memory_bytes": &metrics.Memory,
	}

	// Set metric metadata
	metrics.Invocations.MetricName = "Invocations"
	metrics.Invocations.Unit = "count"
	metrics.Invocations.Description = "Number of function invocations"

	metrics.Duration.MetricName = "Duration"
	metrics.Duration.Unit = "ms"
	metrics.Duration.Description = "Function execution duration"

	metrics.Memory.MetricName = "Memory Usage"
	metrics.Memory.Unit = "bytes"
	metrics.Memory.Description = "Memory used during execution"

	// Track if we successfully fetched any data
	hasData := false

	// Fetch each metric
	for metricType, metricData := range metricTypes {
		dataPoints, err := p.fetchMetricData(ctx, client, metricType, functionName, startTime, endTime)
		if err != nil {
			logger.Logger.Printf("Error fetching metric %s: %v", metricType, err)
			continue
		}
		metricData.DataPoints = dataPoints
		if len(dataPoints) > 0 {
			hasData = true
		}
	}

	// If no real data was found, return sample data
	if !hasData {
		logger.Logger.Printf("No real metrics data found, returning sample data")
		return p.generateSampleMetrics(functionName, startTime, endTime), nil
	}

	return metrics, nil
}

// generateSampleMetrics creates sample metrics data for demonstration
func (p *GCPProvider) generateSampleMetrics(functionName string, startTime, endTime time.Time) *FunctionMetrics {
	logger.Logger.Printf("Generating sample metrics for function: %s", functionName)

	metrics := &FunctionMetrics{
		FunctionName: functionName,
		TimeRange: struct {
			Start time.Time
			End   time.Time
		}{Start: startTime, End: endTime},
	}

	// Generate sample data points over the time range
	duration := endTime.Sub(startTime)
	interval := duration / 12 // Create 12 data points

	var invocationPoints, durationPoints, memoryPoints []MetricDataPoint

	for i := 0; i < 12; i++ {
		timestamp := startTime.Add(time.Duration(i) * interval)

		// Simulate realistic patterns
		invocationPoints = append(invocationPoints, MetricDataPoint{
			Timestamp: timestamp,
			Value:     float64(5 + i%8 + (i*3)%5), // Varying invocation count
		})

		durationPoints = append(durationPoints, MetricDataPoint{
			Timestamp: timestamp,
			Value:     200 + float64((i*37)%150), // Varying duration 200-350ms
		})

		memoryPoints = append(memoryPoints, MetricDataPoint{
			Timestamp: timestamp,
			Value:     float64(50000000 + (i*1000000)%30000000), // Varying memory usage
		})
	}

	metrics.Invocations = MetricData{
		MetricName:  "Invocations",
		Unit:        "count",
		Description: "Number of function invocations (sample data)",
		DataPoints:  invocationPoints,
	}

	metrics.Duration = MetricData{
		MetricName:  "Duration",
		Unit:        "ms",
		Description: "Average function execution duration (sample data)",
		DataPoints:  durationPoints,
	}

	metrics.Memory = MetricData{
		MetricName:  "Memory Usage",
		Unit:        "bytes",
		Description: "Memory used during execution (sample data)",
		DataPoints:  memoryPoints,
	}

	return metrics
}

// fetchMetricData fetches time series data for a specific metric
func (p *GCPProvider) fetchMetricData(ctx context.Context, client *monitoring.MetricClient, metricType, functionName string, startTime, endTime time.Time) ([]MetricDataPoint, error) {
	req := &monitoringpb.ListTimeSeriesRequest{
		Name:   fmt.Sprintf("projects/%s", p.projectID),
		Filter: fmt.Sprintf(`resource.type="cloud_function" AND resource.labels.function_name="%s" AND metric.type="%s"`, functionName, metricType),
		Interval: &monitoringpb.TimeInterval{
			StartTime: timestamppb.New(startTime),
			EndTime:   timestamppb.New(endTime),
		},
		View: monitoringpb.ListTimeSeriesRequest_FULL,
	}

	it := client.ListTimeSeries(ctx, req)
	var dataPoints []MetricDataPoint

	for {
		ts, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error iterating time series: %w", err)
		}

		// Process each point in the time series
		for _, point := range ts.Points {
			timestamp := point.Interval.EndTime.AsTime()
			var value float64

			// Extract value based on value type
			switch v := point.Value.Value.(type) {
			case *monitoringpb.TypedValue_DoubleValue:
				value = v.DoubleValue
			case *monitoringpb.TypedValue_Int64Value:
				value = float64(v.Int64Value)
			case *monitoringpb.TypedValue_DistributionValue:
				// For distributions, use the mean
				if v.DistributionValue != nil {
					value = v.DistributionValue.Mean
				}
			default:
				continue
			}

			dataPoints = append(dataPoints, MetricDataPoint{
				Timestamp: timestamp,
				Value:     value,
			})
		}
	}

	return dataPoints, nil
}

// GetEndpoints gets endpoints associated with a function
func (p *GCPProvider) GetEndpoints(ctx context.Context, name string) ([]string, error) {
	// TODO: Implement real endpoint discovery
	return []string{
		fmt.Sprintf("https://%s-%s.cloudfunctions.net/%s", p.region, p.projectID, name),
		"Note: This is a generated URL. Verify in GCP Console for actual trigger URLs.",
	}, nil
}
