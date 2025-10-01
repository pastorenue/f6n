# GCP Cloud Functions Support

f6n now supports both AWS Lambda and GCP Cloud Functions!

## Quick Start

### Using AWS Lambda (default)
```bash
# AWS is the default provider
f6n

# Or explicitly specify AWS
f6n --provider aws --region us-east-1
```

### Using GCP Cloud Functions
```bash
# Specify GCP as provider with project ID
f6n --provider gcp --gcp-project my-project-id --gcp-location us-central1

# Using environment variables
export CLOUD_PROVIDER=gcp
export GCP_PROJECT=my-project-id
export GCP_LOCATION=us-central1
f6n
```

## Configuration

### Command-Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--provider` | Cloud provider (aws or gcp) | aws |
| `--gcp-project` | GCP project ID | (from GCP_PROJECT env) |
| `--gcp-location` | GCP location/region | us-central1 |
| `--region` | AWS region | us-east-1 |
| `--profile` | AWS profile | (default profile) |
| `--env` | Environment name | dev |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CLOUD_PROVIDER` | aws or gcp |
| `GCP_PROJECT` | GCP project ID |
| `GCP_LOCATION` | GCP location (us-central1, europe-west1, etc.) |
| `AWS_REGION` | AWS region |
| `AWS_PROFILE` | AWS profile name |
| `STAGE` | Environment name (dev, stage, prod) |

## Authentication

### AWS Authentication
```bash
# Using AWS CLI
aws configure

# Or use profile
f6n --provider aws --profile my-profile

# Or set environment variables
export AWS_ACCESS_KEY_ID=your_key
export AWS_SECRET_ACCESS_KEY=your_secret
```

### GCP Authentication
```bash
# Using gcloud CLI (recommended)
gcloud auth application-default login

# Or set credentials file
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json

# Must also provide project ID
f6n --provider gcp --gcp-project my-project-id
```

## Features Supported

### ✅ Currently Working
- **List functions** - View all functions from AWS Lambda or GCP Cloud Functions
- **Inspect functions** - View detailed configuration
- **Provider info** - See which provider you're connected to
- **Multi-region** - Works with any AWS region or GCP location
- **Environment switching** - Switch between dev/stage/prod environments

### 🚧 Coming Soon
- **View logs** - CloudWatch Logs (AWS) and Cloud Logging (GCP)
- **API endpoints** - API Gateway (AWS) and Cloud Functions URLs (GCP)
- **Function code** - Download and view source code
- **Provider switching** - Switch between AWS and GCP without restarting (press 'p')

## Info Display

The info section now shows:
```
Provider: AWS  │  Region: us-east-1      │  Environment: dev
Functions: 5   │  CPU: arm64             │  MEM: 8 cores
OS: darwin     │  User: emmanuelpastor
```

or for GCP:
```
Provider: GCP  │  Region: us-central1    │  Environment: prod
Functions: 3   │  CPU: arm64             │  MEM: 8 cores
OS: darwin     │  User: emmanuelpastor
```

## Examples

### AWS Lambda Examples
```bash
# List Lambda functions in us-west-2
f6n --provider aws --region us-west-2

# Use production AWS profile
f6n --provider aws --profile production --env prod

# Default AWS (region from AWS_REGION or us-east-1)
f6n
```

### GCP Cloud Functions Examples
```bash
# List Cloud Functions in a project
f6n --provider gcp --gcp-project my-gcp-project

# Specific location
f6n --provider gcp --gcp-project my-project --gcp-location europe-west1

# Using environment variables
export CLOUD_PROVIDER=gcp
export GCP_PROJECT=my-production-project
export GCP_LOCATION=us-east1
f6n --env prod
```

## Provider Architecture

f6n uses a provider abstraction pattern:

```
Provider Interface
├── AWS Provider (wraps AWS Lambda SDK)
└── GCP Provider (wraps GCP Cloud Functions API)
```

This makes it easy to:
- Add support for new cloud providers
- Maintain consistent behavior across providers
- Test with dummy data before implementing real APIs

## Current Implementation Status

### AWS Provider
- ✅ Full AWS Lambda SDK integration
- ✅ Authentication via AWS credentials
- ✅ List, inspect, get configuration
- ✅ Real-time data from your AWS account
- 🚧 CloudWatch Logs (coming soon)
- 🚧 API Gateway endpoints (coming soon)

### GCP Provider
- ✅ Provider interface implemented
- ⚠️  Currently using dummy data for development
- 🚧 GCP Cloud Functions SDK integration (coming soon)
- 🚧 Real GCP API calls (coming soon)
- 🚧 Cloud Logging (coming soon)

## Troubleshooting

### AWS Issues
```bash
# Check AWS credentials
aws sts get-caller-identity

# Verify Lambda permissions
aws lambda list-functions --max-items 1
```

### GCP Issues
```bash
# Check GCP authentication
gcloud auth application-default print-access-token

# Verify Cloud Functions access
gcloud functions list --project=YOUR_PROJECT_ID
```

### Common Errors

**"GCP project ID is required"**
- Solution: Provide `--gcp-project` flag or set `GCP_PROJECT` env var

**"Failed to initialize AWS client"**
- Solution: Run `aws configure` or set AWS credentials

**"unsupported provider"**
- Solution: Use `--provider aws` or `--provider gcp` (lowercase)

## Development

### Testing with Dummy Data

Currently, GCP provider uses dummy data (similar to how AWS started):
```go
// In internal/provider/gcp_provider.go
// The ListFunctions method returns dummy data for now
```

To test:
```bash
# This will show dummy GCP functions
f6n --provider gcp --gcp-project test-project
```

### Adding Real GCP Integration

To implement real GCP Cloud Functions API:
1. Add GCP Cloud Functions SDK to `go.mod`
2. Update `internal/provider/gcp_provider.go`
3. Implement real API calls in place of dummy data
4. Test with actual GCP project

## Next Steps

1. **Implement real GCP Cloud Functions API** calls
2. **Add CloudWatch Logs** support for AWS
3. **Add Cloud Logging** support for GCP
4. **Add API Gateway** endpoint discovery (AWS)
5. **Add Cloud Functions URLs** discovery (GCP)
6. **Add provider switching** ('p' key to toggle)
7. **Add function code viewer** for both providers

## Contributing

Want to help implement GCP Cloud Functions integration? Check CONTRIBUTING.md for guidelines!
