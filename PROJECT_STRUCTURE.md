# f6n Project Structure

This document describes the reorganized production-ready structure of the f6n project.

## Directory Structure

```
f6n/
├── cmd/
│   └── f6n/                    # Application entry point
│       └── main.go            # Main function, initializes app
│
├── internal/                   # Private application code
│   ├── aws/                   # AWS Lambda integration
│   │   ├── lambda.go          # Lambda client wrapper
│   │   └── dummy.go           # Dummy data for development
│   │
│   ├── config/                # Configuration management
│   │   └── config.go          # Config loading from env/flags
│   │
│   ├── ui/                    # Terminal UI components
│   │   ├── model.go           # Bubble Tea model & state
│   │   ├── render.go          # View rendering logic
│   │   ├── views.go           # View type definitions
│   │   └── styles/            # UI styling
│   │       └── styles.go      # Color schemes & styles
│   │
│   └── version/               # Version information
│       └── version.go         # Version constants
│
├── bin/                       # Build output (gitignored)
│   └── f6n                    # Compiled binary
│
├── .gitignore                 # Git ignore rules
├── CONTRIBUTING.md            # Contribution guidelines
├── LICENSE                    # MIT License
├── Makefile                   # Build automation
├── README.md                  # Project documentation
├── go.mod                     # Go module definition
├── go.sum                     # Go dependencies checksum
└── main.go.backup             # Original monolithic file (backup)
```

## Package Descriptions

### `cmd/f6n`
The main application entry point. Keeps the main function minimal - just configuration loading, client initialization, and TUI startup.

**Key responsibilities:**
- Load configuration
- Initialize AWS client
- Create and start the TUI

### `internal/aws`
AWS integration layer that wraps the AWS SDK and provides a clean interface for Lambda operations.

**Key files:**
- `lambda.go`: Production Lambda client with full AWS integration
- `dummy.go`: Development data and mock client for testing without AWS

**Key responsibilities:**
- AWS SDK initialization
- Lambda function operations (list, get, configure)
- Pagination handling
- Error wrapping

### `internal/config`
Configuration management using environment variables and command-line flags.

**Key responsibilities:**
- Parse command-line flags
- Read environment variables
- Provide sensible defaults
- Configuration validation

### `internal/ui`
Terminal UI implementation using Bubble Tea framework.

**Key files:**
- `model.go`: Application state and business logic
- `render.go`: View rendering functions
- `views.go`: View type definitions
- `styles/styles.go`: Color schemes and styling

**Key responsibilities:**
- TUI state management
- User input handling
- View rendering
- Navigation logic

### `internal/version`
Application version information for the version command.

**Key responsibilities:**
- Version string management
- Build metadata (git commit, build date)

## Design Principles

### Separation of Concerns
Each package has a single, well-defined responsibility:
- AWS integration is isolated in `internal/aws`
- UI logic is contained in `internal/ui`
- Configuration is centralized in `internal/config`

### Testability
- AWS client is wrapped in an interface
- Dummy data available for testing without AWS
- UI logic separated from rendering

### Maintainability
- Small, focused files (no 500+ line files)
- Clear package boundaries
- Consistent naming conventions
- Comprehensive documentation

### Production Ready
- Proper error handling throughout
- Configuration via flags and environment variables
- Build automation with Make
- Comprehensive documentation
- MIT License included

## Build System

The Makefile provides common development tasks:

```bash
make build           # Build the application
make run             # Build and run
make test            # Run tests
make clean           # Clean build artifacts
make install         # Install to $GOPATH/bin
make fmt             # Format code
make lint            # Run linter
make tidy            # Tidy go modules
make help            # Show all targets
```

## Development Workflow

1. **Make changes** in appropriate package
2. **Format code**: `make fmt`
3. **Run tests**: `make test`
4. **Build**: `make build`
5. **Test locally**: `./bin/f6n` or `make run`
6. **Commit** with conventional commit messages

## Future Expansion

The structure supports future features:

- `internal/cloudwatch/` - For logs integration
- `internal/apigateway/` - For API Gateway endpoints
- `internal/metrics/` - For CloudWatch metrics
- `pkg/` - For any public libraries
- `docs/` - For additional documentation

## Migration Notes

The original `main.go` (502 lines) has been reorganized into:
- `cmd/f6n/main.go` - 45 lines
- `internal/ui/model.go` - 226 lines
- `internal/ui/render.go` - 153 lines
- `internal/ui/styles/styles.go` - 51 lines
- `internal/aws/lambda.go` - 99 lines
- `internal/aws/dummy.go` - 85 lines
- `internal/config/config.go` - 47 lines

Total refactored: ~700 lines (vs 502 lines), but much better organized and maintainable.

## Benefits of Reorganization

1. **Better Organization**: Code is logically grouped by functionality
2. **Easier Testing**: Each component can be tested independently
3. **Clearer Dependencies**: Package boundaries make dependencies explicit
4. **Scalability**: Easy to add new features without bloating files
5. **Team Collaboration**: Multiple developers can work without conflicts
6. **Documentation**: Each package has clear responsibility
7. **Reusability**: Components can be reused across features
