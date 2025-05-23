# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2024-01-15

### Added
- Initial release of MD-to-PDF converter
- Core markdown to PDF conversion functionality
- Plugin system with AST transformers and content generators
- Mermaid diagram support via plugin
- Table of Contents generation plugin
- Comprehensive CLI interface using Cobra
- Configuration management system with YAML persistence
- Support for custom fonts, margins, page sizes, and line spacing
- Document metadata support (title, author, subject, keywords)
- Verbose logging and error handling
- Cross-platform binary distribution
- GitHub Actions CI/CD pipeline
- Comprehensive documentation and examples

### Core Features
- **Markdown Parser**: Full CommonMark support using goldmark
- **PDF Renderer**: High-quality PDF generation using gofpdf
- **Plugin Architecture**: Dynamic plugin loading with .so files
- **Configuration System**: Three-layer config (defaults → user config → CLI flags)
- **CLI Commands**: 
  - `convert` - Convert markdown to PDF with extensive options
  - `config` - Manage user configuration (list, set, reset)

### Supported Markdown Elements
- Headers (H1-H6) with automatic styling
- Text formatting (bold, italic, strikethrough)
- Lists (ordered, unordered, nested)
- Links (inline and reference style)
- Images (local files with proper embedding)
- Code blocks with syntax highlighting
- Tables with column alignment
- Blockquotes with indentation
- Horizontal rules
- Mermaid diagrams (via plugin)

### Plugin System
- **ASTTransformer Interface**: Modify markdown AST before rendering
- **ContentGenerator Interface**: Generate additional content during PDF creation
- **Plugin Manager**: Automatic plugin discovery and loading
- **Priority System**: Control plugin execution order
- **Error Handling**: Graceful plugin failure handling

### Configuration Options
- **Font Settings**: Family, size customization
- **Page Layout**: Size (A4, Letter, Legal), margins
- **Text Formatting**: Line spacing, paragraph spacing
- **Mermaid Settings**: Theme, scale factor
- **Output Options**: File naming, metadata

### Build and Distribution
- **Makefile**: Automated build process
- **Cross-compilation**: Support for multiple platforms
- **Plugin Building**: Separate plugin compilation
- **Installation Script**: Easy one-command installation
- **GitHub Releases**: Automated binary distribution

### Documentation
- **README.md**: Comprehensive usage guide
- **CONTRIBUTING.md**: Developer contribution guidelines
- **DEMO.md**: Feature demonstration with examples
- **API Documentation**: Plugin development guide
- **Examples**: Sample markdown files and plugins

### Testing
- **Unit Tests**: Core functionality coverage
- **Integration Tests**: End-to-end conversion testing
- **Plugin Tests**: Plugin system validation
- **CI/CD**: Automated testing on multiple Go versions

### Security
- **Input Validation**: Sanitized markdown processing
- **Safe Plugin Loading**: Controlled plugin execution
- **Error Handling**: Secure error messages
- **File Permissions**: Proper output file handling

## [0.4.0] - 2024-01-12

### Added
- Polish and documentation phase completion
- Comprehensive error handling with custom error types
- Professional documentation structure
- GitHub Actions workflows for CI/CD
- Installation script for easy deployment
- DEMO.md with feature showcase

### Changed
- Improved error messages and handling
- Enhanced test coverage
- Better project organization

### Fixed
- Test reliability issues
- Documentation formatting
- Build process consistency

## [0.3.0] - 2024-01-10

### Added
- Mermaid plugin for diagram generation
- Table of Contents (TOC) plugin
- Plugin building automation
- Example plugin implementations

### Changed
- Improved mermaid diagram rendering
- Better image positioning in PDFs
- Enhanced plugin loading mechanism

### Fixed
- Mermaid image scaling issues
- Text spacing problems
- Image embedding reliability
- Plugin compilation errors

## [0.2.0] - 2024-01-08

### Added
- Plugin system architecture
- ASTTransformer and ContentGenerator interfaces
- Plugin manager with dynamic loading
- Public plugin API package
- Plugin discovery and loading

### Changed
- Refactored core engine for plugin integration
- Improved module organization
- Enhanced error handling

### Fixed
- Import cycle issues
- Plugin interface compatibility
- Go plugin system integration

## [0.1.0] - 2024-01-05

### Added
- Basic markdown to PDF conversion
- Core engine architecture
- Markdown parser using goldmark
- PDF renderer using gofpdf
- CLI interface using Cobra
- Project structure and module setup

### Core Components
- **Engine**: Central conversion orchestrator
- **Parser**: Markdown processing with goldmark
- **Renderer**: PDF generation with gofpdf
- **CLI**: Command-line interface with Cobra

### Supported Features
- Basic markdown elements (headers, text, lists)
- Simple PDF output
- Command-line conversion
- File input/output handling

---

## Version History Summary

- **v1.0.0**: Full-featured release with plugin system, configuration management, and comprehensive documentation
- **v0.4.0**: Polish and documentation phase
- **v0.3.0**: Plugin examples and improvements
- **v0.2.0**: Plugin system implementation
- **v0.1.0**: Initial core functionality

## Development Phases

### Phase 1: Core Functionality ✅
- Basic markdown to PDF conversion
- Core engine architecture
- CLI interface foundation

### Phase 2: Plugin System ✅
- Plugin interface design
- Dynamic plugin loading
- Plugin manager implementation

### Phase 3: Example Plugins ✅
- Mermaid diagram support
- Table of Contents generation
- Plugin development examples

### Phase 4: Polish & Documentation ✅
- Error handling and testing
- Comprehensive documentation
- CI/CD and distribution
- Production readiness

## Future Roadmap

### Potential v1.1.0 Features
- Additional diagram types (PlantUML, Graphviz)
- Custom CSS styling support
- Batch conversion capabilities
- Web interface option
- Additional export formats

### Plugin Ecosystem
- Community plugin repository
- Plugin template generator
- Enhanced plugin APIs
- Plugin marketplace

---

## Contributors

- Initial development and architecture
- Plugin system design
- Documentation and testing
- CI/CD implementation

## Acknowledgments

Special thanks to the open-source projects that made this possible:
- [goldmark](https://github.com/yuin/goldmark) - Markdown parsing
- [gofpdf](https://github.com/jung-kurt/gofpdf) - PDF generation  
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [mermaid](https://mermaid.js.org/) - Diagram generation