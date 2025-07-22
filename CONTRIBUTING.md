# Contributing to Auto PR

Thank you for your interest in contributing to Auto PR! We welcome contributions from the community and are grateful for any help you can provide.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct:
- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive criticism
- Respect differing viewpoints and experiences

## How to Contribute

### Reporting Issues

1. Check if the issue already exists in the [issue tracker](https://github.com/charles-adedotun/auto-pr/issues)
2. If not, create a new issue with:
   - Clear, descriptive title
   - Detailed description of the problem
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, etc.)

### Suggesting Features

1. Check existing [issues](https://github.com/charles-adedotun/auto-pr/issues) and [discussions](https://github.com/charles-adedotun/auto-pr/discussions)
2. Open a new discussion or issue with:
   - Clear description of the feature
   - Use cases and benefits
   - Potential implementation approach

### Contributing Code

#### Setup Development Environment

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/auto-pr.git
   cd auto-pr
   ```
3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/charles-adedotun/auto-pr.git
   ```
4. Install dependencies:
   ```bash
   go mod download
   ```

#### Development Workflow

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following our coding standards

3. Write or update tests:
   ```bash
   go test ./...
   ```

4. Run linting and formatting:
   ```bash
   go fmt ./...
   go vet ./...
   golangci-lint run  # if installed
   ```

5. Commit your changes:
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   ```
   
   Follow [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` new feature
   - `fix:` bug fix
   - `docs:` documentation changes
   - `style:` formatting, missing semi-colons, etc.
   - `refactor:` code restructuring
   - `test:` adding tests
   - `chore:` maintenance tasks

6. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

7. Create a Pull Request

#### Pull Request Guidelines

1. **Title**: Use a clear, descriptive title following conventional commits
2. **Description**: Include:
   - What changes were made and why
   - Related issue numbers
   - Screenshots for UI changes
   - Breaking changes (if any)
3. **Tests**: Ensure all tests pass
4. **Documentation**: Update relevant documentation
5. **Sign-off**: Include `Signed-off-by` in commits if required

### Code Standards

#### Go Code Style

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for exported functions
- Keep functions focused and small
- Handle errors appropriately
- Write tests for new functionality

Example:
```go
// AnalyzeRepository analyzes git repository and returns status
func AnalyzeRepository(path string) (*RepoStatus, error) {
    if path == "" {
        return nil, fmt.Errorf("path cannot be empty")
    }
    
    // Implementation...
    return status, nil
}
```

#### Testing

- Write unit tests for new functions
- Maintain or improve code coverage
- Use table-driven tests where appropriate
- Mock external dependencies

Example:
```go
func TestAnalyzeRepository(t *testing.T) {
    tests := []struct {
        name    string
        path    string
        want    *RepoStatus
        wantErr bool
    }{
        {
            name:    "empty path",
            path:    "",
            wantErr: true,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := AnalyzeRepository(tt.path)
            if (err != nil) != tt.wantErr {
                t.Errorf("AnalyzeRepository() error = %v, wantErr %v", err, tt.wantErr)
            }
            // More assertions...
        })
    }
}
```

### Documentation

- Update README.md for user-facing changes
- Add/update code comments
- Update configuration examples
- Document new environment variables
- Add examples for new features

### Review Process

1. Maintainers will review your PR
2. Address any feedback or requested changes
3. Once approved, your PR will be merged
4. Your contribution will be included in the next release

## Development Tips

### Running Locally

```bash
# Build
go build -o auto-pr .

# Run
./auto-pr --help

# Run with verbose output
./auto-pr create --verbose --dry-run
```

### Debugging

```bash
# Enable debug logging
export AUTO_PR_DEBUG=true

# Run with race detector
go run -race . create
```

### Common Issues

1. **Import errors**: Run `go mod tidy`
2. **Formatting issues**: Run `go fmt ./...`
3. **Test failures**: Check test output and update as needed

## Release Process

1. Maintainers create release branches
2. Version bumps follow semantic versioning
3. Releases are automated via GitHub Actions
4. Release notes are generated from commit messages

## Getting Help

- Check the [documentation](https://github.com/charles-adedotun/auto-pr/tree/main/docs)
- Ask in [discussions](https://github.com/charles-adedotun/auto-pr/discussions)
- Join our community chat (if available)

## Recognition

Contributors will be recognized in:
- Release notes
- Contributors list
- Project documentation

Thank you for contributing to Auto PR!