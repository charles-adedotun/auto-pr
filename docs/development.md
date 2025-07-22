# Development Guide

This guide covers the development workflow and quality processes for Auto PR.

## Quick Setup

1. **Clone and setup**:
   ```bash
   git clone https://github.com/charles-adedotun/auto-pr.git
   cd auto-pr
   make dev-setup  # Installs hooks and dependencies
   ```

2. **Development workflow**:
   ```bash
   make check      # Run all quality checks before committing
   make build      # Build the binary
   make test       # Run tests
   ```

## Quality Checks

We maintain code quality through automated checks that run:
- **Before commit**: Via git pre-commit hooks
- **On PR**: Via GitHub Actions CI
- **Locally**: Via make commands

### Pre-commit Hooks

Pre-commit hooks automatically run quality checks to catch issues early:

```bash
# Install hooks (done automatically by make dev-setup)
make install-hooks

# Run pre-commit checks manually
make pre-commit
```

The pre-commit hook checks:
- ✅ **Code formatting** (`go fmt`)
- ✅ **Code issues** (`go vet`) 
- ✅ **Tests pass** (`go test`)
- ✅ **Linting** (`golangci-lint`) - if available

### Manual Quality Checks

```bash
# Run all checks (recommended before committing)
make check

# Individual checks
make format     # Format code with go fmt
make lint       # Run golangci-lint
make test       # Run tests
make vet        # Run go vet
```

## Development Commands

### Building

```bash
# Build for development
make build

# Build for all platforms (release)
make build-release
```

### Testing

```bash
# Run all tests
make test

# Run tests with race detector
make test-race

# Run tests for specific package
go test ./internal/config -v
```

### Linting

```bash
# Run linter
make lint

# Auto-fix some linting issues
golangci-lint run --fix
```

## Quality Standards

### Code Formatting
- All Go code must be formatted with `go fmt`
- Pre-commit hooks enforce formatting automatically

### Error Handling
- All error return values must be checked
- Use `_` to explicitly ignore errors when safe
- Provide meaningful error messages

### Testing
- All new functionality should include tests
- Tests must pass before committing
- Use table-driven tests where appropriate

### Linting
- Code must pass `golangci-lint` checks
- Fix all linting issues before committing
- Use `//nolint` comments sparingly and with justification

## Bypassing Checks

⚠️  **Not recommended** - Only use in emergency situations:

```bash
# Skip pre-commit hooks
git commit --no-verify -m "emergency fix"

# Skip specific checks
SKIP_TESTS=1 git commit -m "docs only change"
```

## CI/CD Pipeline

Our GitHub Actions pipeline runs:

1. **Format Check**: Ensures code is properly formatted
2. **Lint**: Runs golangci-lint with strict settings  
3. **Test**: Runs tests on Go 1.21, 1.22, and 1.23
4. **Build**: Ensures code compiles for multiple platforms

## Installing Development Tools

### golangci-lint (Required for linting)

```bash
# macOS
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

# Windows
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
```

### Other Tools

```bash
# Install all Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/tools/cmd/godoc@latest
```

## IDE Integration

### VS Code
Add to your `settings.json`:
```json
{
    "go.formatTool": "gofmt",
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "package",
    "editor.formatOnSave": true
}
```

### GoLand/IntelliJ
1. Go to **Settings** → **Tools** → **Go Linters**
2. Enable **golangci-lint**
3. Enable **Format on save**

## Troubleshooting

### Pre-commit hooks not running
```bash
# Ensure hooks are executable
chmod +x .git/hooks/pre-commit

# Reinstall hooks
make install-hooks
```

### Linting errors
```bash
# See detailed linting errors
golangci-lint run --verbose

# Auto-fix some issues
golangci-lint run --fix
```

### Tests failing
```bash
# Run specific test
go test ./internal/config -v -run TestValidateConfig

# Run tests with more detail
go test ./... -v -count=1
```

This development process ensures high code quality while maintaining development velocity.