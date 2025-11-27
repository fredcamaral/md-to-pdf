# Contributing to md-to-pdf

This document helps you contribute to md-to-pdf.

## Table of contents

- [Code of conduct](#code-of-conduct)
- [Getting started](#getting-started)
- [Development setup](#development-setup)
- [Contributing guidelines](#contributing-guidelines)
- [Plugin development](#plugin-development)
- [Testing](#testing)
- [Documentation](#documentation)
- [Submitting changes](#submitting-changes)
- [Code review process](#code-review-process)

## Code of conduct

Follow our code of conduct. Be respectful, inclusive, and considerate in all interactions.

## Getting started

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
   git remote add upstream https://github.com/fredcamaral/md-to-pdf.git
   ```

## Development setup

### Build the project

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

### Project structure

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
├── plugins/               # Plugin directory and development guide
├── scripts/               # Build and deployment scripts
└── tests/                 # Test files
```

### Development workflow

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

## Contributing guidelines

### Code style

- Follow standard Go conventions and idioms
- Use `gofmt` and `golint` to format code
- Write clear, self-documenting code
- Add comments for complex logic
- Use meaningful variable and function names

### Commit messages

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

### Code organization

- Keep your functions small and focused (ideally under 50 lines)
- Use interfaces to define contracts
- Separate concerns clearly
- Follow the dependency inversion principle
- Handle errors explicitly

### Error handling

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

## Plugin development

### Plugin types

md-to-pdf supports two types of plugins:

#### AST transformers
Modify the markdown AST before rendering:

```go
type ASTTransformer interface {
    Plugin
    Transform(node ast.Node, ctx *TransformContext) (ast.Node, error)
    Priority() int
    SupportedNodes() []ast.NodeKind
}
```

#### Content generators
Generate additional content during PDF creation:

```go
type ContentGenerator interface {
    Plugin
    GenerateContent(ctx *GenerationContext) error
    Priority() int
}
```

### Creating a plugin

1. Create a new directory in `examples/plugins/`
2. Implement the required interfaces
3. Export the plugin creation function
4. Build as a shared library

Example plugin structure:
```go
package main

import (
    "github.com/fredcamaral/md-to-pdf/pkg/plugin"
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

### Plugin guidelines

- Follow the plugin interface contracts
- Handle errors gracefully
- Document plugin functionality
- Include example usage
- Test plugin functionality

## Testing

### Running tests

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

### Writing tests

- Write unit tests for all public functions you create
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

### Code documentation

- Use Go doc comments for exported functions and types
- Follow the standard Go documentation format
- Include examples in documentation when helpful

### README updates

When adding new features:
- Update the feature list
- Add usage examples
- Update configuration options if applicable

### Changelog

Update CHANGELOG.md for all changes:
- Follow the Keep a Changelog format
- Include the type of change (Added, Changed, Deprecated, Removed, Fixed, Security)
- Reference issue numbers when applicable

## Submitting changes

### Pull request process

1. **Update documentation**: Update documentation to reflect your changes
2. **Add tests**: Include tests for new functionality
3. **Update changelog**: Add entry to CHANGELOG.md
4. **Clean commit history**: Squash commits if necessary
5. **Fill PR template**: Provide clear description of changes

### Pull request template

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

### What to include

- Clear description of the problem and solution
- Screenshots for UI changes
- Performance impact assessment for significant changes
- Breaking change documentation

## Code review process

### Reviewer guidelines

As a reviewer, you should:

- Focus on code quality, maintainability, and correctness
- Suggest improvements constructively
- Test the changes locally when possible
- Check documentation and tests

### Author guidelines

As an author, you should:

- Respond to feedback promptly
- Make requested changes or discuss alternatives
- Keep discussions focused and professional
- Update PR based on feedback

### Approval process

- At least one maintainer approval required
- All CI checks must pass
- No unresolved conversations
- Up-to-date with main branch

## Development tips

### Debugging

- Use the `-v` flag for verbose output
- Add logging statements for debugging
- Use Go's debugging tools (delve, pprof)

### Performance

- Profile critical code paths
- Benchmark performance-sensitive code
- Consider memory allocation patterns
- Test with large files

### Plugin development

- Start with the example plugins
- Use the provided plugin interfaces
- Test plugins with various markdown inputs
- Document plugin configuration options

## Getting help

- **Issues**: Create an issue for bugs or feature requests
- **Discussions**: Use GitHub Discussions for questions
- **Documentation**: Check the README.md and plugins/README.md
- **Examples**: Look at example plugins and markdown files

## Recognition

Contributors will be recognized in:
- CHANGELOG.md for significant contributions
- README.md contributors section
- Release notes for major features

Thank you for contributing to md-to-pdf.