# Contributing to f6n

Thank you for your interest in contributing to f6n! This document provides guidelines and instructions for contributing.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Respect different viewpoints and experiences

## How to Contribute

### Reporting Bugs

When reporting bugs, please include:

1. **Clear title and description** of the issue
2. **Steps to reproduce** the problem
3. **Expected behavior** vs **actual behavior**
4. **Environment details**: OS, Go version, AWS region
5. **Logs or error messages** if applicable
6. **Screenshots** if relevant (especially for UI issues)

### Suggesting Enhancements

Enhancement suggestions are welcome! Please:

1. **Check existing issues** to avoid duplicates
2. **Clearly describe** the feature and its use case
3. **Explain why** this enhancement would be useful
4. **Provide examples** of how it would work

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Follow the coding standards** (see below)
3. **Add tests** for new functionality
4. **Update documentation** as needed
5. **Ensure tests pass** with `make test`
6. **Format code** with `make fmt`
7. **Write clear commit messages**

#### Pull Request Process

1. Update the README.md with details of changes if needed
2. Update the roadmap in README.md if completing a feature
3. The PR will be merged once reviewed and approved

## Development Setup

### Prerequisites

- Go 1.24 or later
- AWS CLI configured (optional, for testing)
- Make

### Setup Steps

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/f6n.git
cd f6n

# Install dependencies
make deps

# Build the project
make build

# Run tests
make test

# Run the application
make run
```

### Development Workflow

```bash
# Create a feature branch
git checkout -b feature/my-new-feature

# Make your changes
# ... edit files ...

# Format code
make fmt

# Run tests
make test

# Build and test locally
make build
./bin/f6n

# Commit your changes
git add .
git commit -m "Add my new feature"

# Push to your fork
git push origin feature/my-new-feature

# Open a pull request on GitHub
```

## Coding Standards

### Go Style Guide

- Follow the [official Go style guide](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting (automatically done by `make fmt`)
- Keep functions small and focused
- Use meaningful variable and function names
- Add comments for exported functions and types

### Project Structure

```
f6n/
├── cmd/f6n/          # Main application entry point
├── internal/         # Private application code
│   ├── aws/          # AWS integration
│   ├── config/       # Configuration management
│   └── ui/           # Terminal UI components
├── pkg/              # Public libraries (if any)
└── docs/             # Additional documentation
```

### Code Organization

- **cmd/**: Application entry points
- **internal/**: Private application code (cannot be imported by other projects)
- **pkg/**: Public libraries (can be imported by other projects)
- Keep packages focused on a single responsibility
- Use interfaces for testability

### Naming Conventions

- **Packages**: lowercase, single word when possible
- **Files**: lowercase with underscores (e.g., `lambda_client.go`)
- **Types**: PascalCase (e.g., `LambdaClient`)
- **Functions**: camelCase for private, PascalCase for exported
- **Constants**: PascalCase or ALL_CAPS depending on context

### Error Handling

```go
// Always check errors
result, err := someFunction()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Use error wrapping with %w
return fmt.Errorf("context: %w", err)

// Return errors, don't panic unless truly unrecoverable
```

### Testing

- Write tests for new functionality
- Use table-driven tests when appropriate
- Mock external dependencies
- Aim for good coverage, but focus on critical paths

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case 1", "input1", "expected1"},
        {"case 2", "input2", "expected2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := doSomething(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Comments

- Use complete sentences
- Start with the name of the thing being described
- Exported items must have doc comments

```go
// LambdaClient wraps the AWS Lambda client with custom methods.
// It provides a simplified interface for common Lambda operations.
type LambdaClient struct {
    client *lambda.Client
    region string
}

// Region returns the AWS region this client is configured for.
func (c *LambdaClient) Region() string {
    return c.region
}
```

## Commit Messages

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc.)
- **refactor**: Code refactoring
- **test**: Adding or updating tests
- **chore**: Maintenance tasks

### Examples

```
feat(ui): add keyboard shortcut for environment switching

Add 'e' key to switch between environments. This allows users
to quickly toggle between dev, stage, and prod environments.

Closes #42
```

```
fix(aws): handle pagination in ListFunctions

The previous implementation didn't handle paginated results,
causing some functions to be missed when listing.
```

## Testing with AWS

### Using Dummy Data

By default, the application uses dummy data for development. This allows testing without AWS credentials.

### Testing with Real AWS

To test with real AWS resources:

1. Configure AWS credentials
2. Edit `internal/aws/dummy.go` and set `UseDummyData = false`
3. Build and run the application

**Important**: Be careful when testing with production AWS accounts!

## Documentation

- Keep README.md up to date
- Update inline documentation for code changes
- Add examples for new features
- Update the roadmap when completing features

## Getting Help

- Open an issue for questions
- Check existing issues and PRs
- Be patient and respectful

## License

By contributing to f6n, you agree that your contributions will be licensed under the MIT License.
