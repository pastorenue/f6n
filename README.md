# f6n - Serverless Function Manager

A multi-cloud beautiful terminal UI for managing serverless functions, inspired by k9s.

```
  _____  ________       
_/ ____\/  _____/ ____  
\   __\/   __  \ /    \ 
 |  |  \  |__\  \   |  \
 |__|   \_____  /___|  /
              \/     \/ 
```

## Features

- 📋 **List all Lambda/Cloud functions** in your AWS/GCP account
- 🔍 **Inspect function details** including configuration, environment variables, and metadata
- 📊 **View function metrics** and status
- 🔄 **Refresh in real-time** to see the latest changes
- 🌍 **Multi-region support** - switch between AWS regions
- 🎨 **Beautiful TUI** - clean and intuitive interface
- ⌨️  **Keyboard-driven** - fast navigation with keyboard shortcuts

## Planned Features

- 📝 View CloudWatch logs for functions
- 🔗 Get API Gateway endpoints associated with functions
- 💻 View function source code
- 🌐 Switch between environments (dev, stage, prod)
- 🔧 Support for GCP Cloud Functions (future)

## Installation

### Prerequisites

- Go 1.24+ installed
- AWS CLI configured with valid credentials
- Access to AWS Lambda in your account

### Install from source

```bash
# Clone the repository
git clone <repository-url>
cd f6n

# Build and install
make install

# Or build only
make build
```

### Binary

After building, the binary will be available at `./bin/f6n`

## Configuration

### AWS Credentials

f6n uses the standard AWS credential chain. You can configure it in several ways:

1. **AWS CLI Configuration** (recommended)
   ```bash
   aws configure
   ```

2. **Environment Variables**
   ```bash
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_REGION=us-east-1
   ```

3. **AWS Profile**
   ```bash
   f6n --profile my-profile
   ```

### Command-line Options

```bash
f6n [options]

Options:
  --region string      AWS region (default: AWS_REGION env var or us-east-1)
  --env string         Environment name (default: STAGE env var or dev)
  --profile string     AWS profile to use (default: AWS_PROFILE env var)
  --log-level string   Log level: debug, info, warn, error (default: info)
```

## Usage

### Starting f6n

```bash
# Use default configuration
f6n

# Specify region
f6n --region us-west-2

# Use specific AWS profile
f6n --profile production

# Set environment
f6n --env prod
```

### Keyboard Shortcuts

#### List View
- `↑/↓` or `j/k` - Navigate through functions
- `Enter` - View function details
- `r` - Refresh function list
- `e` - Change environment (coming soon)
- `l` - View logs (coming soon)
- `a` - View API Gateway endpoints (coming soon)
- `c` - View function code (coming soon)
- `q` or `Ctrl+C` - Quit

#### Detail View
- `↑/↓` - Scroll through details
- `Esc` - Return to list view
- `q` - Quit

## Development

### Project Structure

```
f6n/
├── cmd/
│   └── f6n/           # Application entry point
│       └── main.go
├── internal/
│   ├── aws/           # AWS Lambda client wrapper
│   │   ├── lambda.go
│   │   └── dummy.go   # Dummy data for testing
│   ├── config/        # Configuration management
│   │   └── config.go
│   └── ui/            # Terminal UI components
│       ├── model.go   # TUI model and state
│       ├── render.go  # Rendering logic
│       ├── views.go   # View types
│       └── styles/    # UI styling
│           └── styles.go
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Building

```bash
# Build the application
make build

# Run tests
make test

# Run linter
make lint

# Format code
make fmt

# Run the application
make run
```

### Testing with Dummy Data

By default, the application uses dummy data for development. To use real AWS data:

Edit `internal/aws/dummy.go` and set:
```go
var UseDummyData = false
```

### Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style

- Follow standard Go conventions
- Run `make fmt` before committing
- Ensure all tests pass with `make test`
- Add tests for new features

## Technologies Used

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2) - AWS integration

## License

MIT License - see LICENSE file for details

## Roadmap

- [ ] CloudWatch Logs integration
- [ ] API Gateway endpoint discovery
- [ ] Function code viewer
- [ ] Environment switching
- [ ] Function invocation
- [ ] Metrics and monitoring
- [ ] Multi-cloud support (GCP Cloud Functions)
- [ ] Configuration file support
- [ ] Search and filtering
- [ ] Function tagging support

## Support

For issues, questions, or contributions, please open an issue on GitHub.
