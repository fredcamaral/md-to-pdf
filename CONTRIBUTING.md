# Contributing to MD-to-PDF

Thank you for your interest in contributing to MD-to-PDF! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Plugin Development](#plugin-development)
- [Testing](#testing)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)
- [Code Review Process](#code-review-process)

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please be respectful, inclusive, and considerate in all interactions.

## Getting Started

### Prerequisites

- **Go**: Version 1.21 or later
- **Git**: For version control
- **Make**: For build automation
- **mermaid-cli**: For mermaid plugin development (`npm install -g @mermaid-js/mermaid-cli`)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/md-to-pdf.git
   cd md-to-pdf
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/original-owner/md-to-pdf.git
   ```

## Development Setup

### Build the Project

```bash
# Install dependencies
go mod download

# Build the main binary
make build

# Build plugins
make build-plugins

# Run tests
make test
```

### Project Structure

```
md-to-pdf/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command setup
│   ├── convert.go         # Convert command
│   └── config.go          # Config commands
├── internal/              # Internal packages
│   ├── core/              # Core conversion engine
│   ├── parser/            # Markdown parsing
│   ├── renderer/          # PDF rendering
│   ├── plugins/           # Plugin system
│   └── config/            # Configuration management
├── examples/              # Example files and plugins
│   ├── plugins/           # Example plugin implementations
│   └── markdown/          # Sample markdown files
├── pkg/                   # Public API packages
│   └── plugin/            # Plugin development API
├── docs/                  # Documentation
├── scripts/               # Build and deployment scripts
└── tests/                 # Test files
```

### Development Workflow

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following the coding standards

3. Run tests and ensure they pass:
   ```bash
   make test
   make lint
   ```

4. Commit your changes with a descriptive message

5. Push to your fork and create a pull request

## Contributing Guidelines

### Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` and `golint` to format code
- Write clear, self-documenting code
- Add comments for complex logic
- Use meaningful variable and function names

### Commit Messages

Follow the conventional commit format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code refactoring
- `style`: Code style changes
- `test`: Adding or updating tests
- `docs`: Documentation changes
- `build`: Build system changes
- `breaking`: Breaking changes

Examples:
```
feat(plugins): add new content generator interface
fix(renderer): resolve image positioning issue
docs(readme): update installation instructions
```

### Code Organization

- Keep functions small and focused (ideally under 50 lines)
- Use interfaces to define contracts
- Separate concerns clearly
- Follow the dependency inversion principle
- Handle errors explicitly

### Error Handling

- Use Go's idiomatic error handling
- Create custom error types when appropriate
- Provide meaningful error messages
- Don't ignore errors

Example:
```go
func (e *Engine) Convert(opts ConversionOptions) error {
    if err := e.validateOptions(opts); err != nil {
        return fmt.Errorf("invalid options: %w", err)
    }
    
    // ... rest of function
    
    if err := e.render(doc); err != nil {
        return fmt.Errorf("rendering failed: %w", err)
    }
    
    return nil
}
```

## Plugin Development

### Plugin Types

MD-to-PDF supports two types of plugins:

#### AST Transformers
Modify the markdown AST before rendering:

```go
type ASTTransformer interface {
    Plugin
    Transform(node ast.Node, ctx *TransformContext) (ast.Node, error)
    Priority() int
    SupportedNodes() []ast.NodeKind
}
```

#### Content Generators
Generate additional content during PDF creation:

```go
type ContentGenerator interface {
    Plugin
    GenerateContent(ctx *GenerationContext) error
    Priority() int
}
```

### Creating a Plugin

1. Create a new directory in `examples/plugins/`
2. Implement the required interfaces
3. Export the plugin creation function
4. Build as a shared library

Example plugin structure:
```go
package main

import (
    "github.com/your-username/md-to-pdf/pkg/plugin"
)

type MyPlugin struct{}

func (p *MyPlugin) Name() string { return "myplugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }
func (p *MyPlugin) Description() string { return "My custom plugin" }

// Plugin creation function (required)
func NewPlugin() plugin.Plugin {
    return &MyPlugin{}
}
```

### Plugin Guidelines

- Follow the plugin interface contracts
- Handle errors gracefully
- Document plugin functionality
- Include example usage
- Test plugin functionality

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test ./internal/core -v

# Run tests with race detection
go test -race ./...
```

### Writing Tests

- Write unit tests for all public functions
- Use table-driven tests when appropriate
- Mock external dependencies
- Test error conditions
- Aim for high test coverage

Example test structure:
```go
func TestEngine_Convert(t *testing.T) {
    tests := []struct {
        name    string
        opts    ConversionOptions
        wantErr bool
    }{
        {
            name: "valid conversion",
            opts: ConversionOptions{
                InputFile:  "test.md",
                OutputFile: "test.pdf",
            },
            wantErr: false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            engine := NewEngine()
            err := engine.Convert(tt.opts)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Documentation

### Code Documentation

- Use Go doc comments for exported functions and types
- Follow the standard Go documentation format
- Include examples in documentation when helpful

### README Updates

When adding new features:
- Update the feature list
- Add usage examples
- Update configuration options if applicable

### Changelog

Update CHANGELOG.md for all changes:
- Follow the Keep a Changelog format
- Include the type of change (Added, Changed, Deprecated, Removed, Fixed, Security)
- Reference issue numbers when applicable

## Submitting Changes

### Pull Request Process

1. **Update Documentation**: Ensure documentation reflects your changes
2. **Add Tests**: Include tests for new functionality
3. **Update Changelog**: Add entry to CHANGELOG.md
4. **Clean Commit History**: Squash commits if necessary
5. **Fill PR Template**: Provide clear description of changes

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass
- [ ] New tests added
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Changelog updated
```

### What to Include

- Clear description of the problem and solution
- Screenshots for UI changes
- Performance impact assessment for significant changes
- Breaking change documentation

## Code Review Process

### Reviewer Guidelines

- Focus on code quality, maintainability, and correctness
- Suggest improvements constructively
- Test the changes locally when possible
- Check documentation and tests

### Author Guidelines

- Respond to feedback promptly
- Make requested changes or discuss alternatives
- Keep discussions focused and professional
- Update PR based on feedback

### Approval Process

- At least one maintainer approval required
- All CI checks must pass
- No unresolved conversations
- Up-to-date with main branch

## Development Tips

### Debugging

- Use the `-v` flag for verbose output
- Add logging statements for debugging
- Use Go's debugging tools (delve, pprof)

### Performance

- Profile critical code paths
- Benchmark performance-sensitive code
- Consider memory allocation patterns
- Test with large files

### Plugin Development

- Start with the example plugins
- Use the provided plugin interfaces
- Test plugins with various markdown inputs
- Document plugin configuration options

## Getting Help

- **Issues**: Create an issue for bugs or feature requests
- **Discussions**: Use GitHub Discussions for questions
- **Documentation**: Check the docs/ directory
- **Examples**: Look at example plugins and markdown files

## Recognition

Contributors will be recognized in:
- CHANGELOG.md for significant contributions
- README.md contributors section
- Release notes for major features

Thank you for contributing to MD-to-PDF!