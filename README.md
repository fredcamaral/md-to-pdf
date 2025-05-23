# MD-to-PDF

A powerful, extensible markdown to PDF converter written in Go with plugin support.

## Features

- 🚀 **Fast & Reliable**: Built in Go for performance and reliability
- 🔌 **Plugin System**: Extensible architecture with custom plugin support
- 🎨 **Rich Formatting**: Support for tables, code blocks, images, and more
- 📊 **Mermaid Diagrams**: Built-in support for Mermaid diagram generation
- ⚙️ **Configurable**: Extensive customization options for fonts, margins, spacing
- 📖 **Table of Contents**: Automatic TOC generation
- 🖥️ **CLI Interface**: Easy-to-use command-line interface with sensible defaults

## Quick Start

### Installation

#### Using the install script (recommended):
```bash
curl -sSL https://raw.githubusercontent.com/your-username/md-to-pdf/main/install.sh | bash
```

#### Manual installation:
1. Download the latest release from [GitHub Releases](https://github.com/your-username/md-to-pdf/releases)
2. Extract and move the binary to your PATH

#### Build from source:
```bash
git clone https://github.com/your-username/md-to-pdf.git
cd md-to-pdf
make build
```

### Basic Usage

Convert a markdown file to PDF:
```bash
md-to-pdf convert document.md
```

Convert with custom options:
```bash
md-to-pdf convert document.md \
  --output custom-name.pdf \
  --title "My Document" \
  --author "John Doe" \
  --font-size 12 \
  --margins "20,20,20,20"
```

### Configuration

View current configuration:
```bash
md-to-pdf config list
```

Set configuration values:
```bash
md-to-pdf config set font.family "Times New Roman"
md-to-pdf config set page.margins "25,25,25,25"
md-to-pdf config set text.fontSize 11
```

Reset configuration to defaults:
```bash
md-to-pdf config reset
```

## Plugin System

MD-to-PDF supports two types of plugins:

### AST Transformers
Modify the markdown abstract syntax tree before rendering:
```go
type ASTTransformer interface {
    Transform(node ast.Node, ctx *TransformContext) (ast.Node, error)
    Priority() int
    SupportedNodes() []ast.NodeKind
}
```

### Content Generators
Generate additional content during PDF creation:
```go
type ContentGenerator interface {
    GenerateContent(ctx *GenerationContext) error
    Priority() int
}
```

### Available Plugins

- **Mermaid Plugin**: Converts mermaid code blocks to PNG diagrams
- **TOC Plugin**: Generates table of contents

### Loading Plugins

Place plugin `.so` files in the `plugins/` directory and they will be loaded automatically.

📖 **[Plugin Development Guide](plugins/README.md)** - Learn how to create custom plugins

## Configuration Options

| Category | Option | Default | Description |
|----------|--------|---------|-------------|
| Font | `font.family` | "Arial" | Font family name |
| Font | `font.size` | 10 | Font size in points |
| Page | `page.size` | "A4" | Page size (A4, Letter, Legal) |
| Page | `page.margins` | "20,20,20,20" | Margins (top,right,bottom,left) |
| Text | `text.lineSpacing` | 1.2 | Line spacing multiplier |
| Mermaid | `mermaid.theme` | "default" | Mermaid theme |
| Mermaid | `mermaid.scale` | 2.2 | Mermaid diagram scale |

## CLI Commands

### Convert Command
```bash
md-to-pdf convert [file] [flags]
```

#### Flags:
- `--output, -o`: Output PDF file name
- `--title`: Document title
- `--author`: Document author
- `--subject`: Document subject
- `--keywords`: Document keywords
- `--font-family`: Font family
- `--font-size`: Font size
- `--page-size`: Page size (A4, Letter, Legal)
- `--margins`: Page margins "top,right,bottom,left"
- `--line-spacing`: Text line spacing
- `--mermaid-theme`: Mermaid theme
- `--mermaid-scale`: Mermaid scale factor
- `--plugins-dir`: Plugins directory
- `--verbose, -v`: Verbose output

### Config Commands
```bash
md-to-pdf config list                    # List all configuration
md-to-pdf config set <key> <value>      # Set configuration value
md-to-pdf config reset                  # Reset to defaults
```

## Examples

### Basic Conversion
```bash
# Convert with defaults
md-to-pdf convert README.md

# Convert with custom output name
md-to-pdf convert README.md -o documentation.pdf
```

### Custom Styling
```bash
# Larger font and margins
md-to-pdf convert document.md \
  --font-size 12 \
  --margins "30,30,30,30" \
  --line-spacing 1.5

# Different font family
md-to-pdf convert document.md \
  --font-family "Times New Roman"
```

### Document Metadata
```bash
md-to-pdf convert report.md \
  --title "Monthly Report" \
  --author "Jane Smith" \
  --subject "Business Analytics" \
  --keywords "report,analytics,monthly"
```

### Mermaid Diagrams
```bash
# Custom mermaid settings
md-to-pdf convert flowchart.md \
  --mermaid-theme dark \
  --mermaid-scale 3.0
```

## Supported Markdown Features

- **Headers** (H1-H6)
- **Emphasis** (bold, italic, strikethrough)
- **Lists** (ordered, unordered, nested)
- **Links** (inline, reference)
- **Images** (local files, embedded)
- **Code blocks** (syntax highlighting)
- **Tables** (with alignment)
- **Blockquotes**
- **Horizontal rules**
- **Mermaid diagrams** (via plugin)

## Development

### Prerequisites
- Go 1.21 or later
- Make
- Git

### Building
```bash
make build          # Build binary
make build-plugins  # Build plugins
make test           # Run tests
make clean          # Clean build artifacts
```

### Project Structure
```
md-to-pdf/
├── cmd/                    # CLI commands
├── internal/
│   ├── core/              # Core conversion engine
│   ├── parser/            # Markdown parsing
│   ├── renderer/          # PDF rendering
│   ├── plugins/           # Plugin system
│   └── config/            # Configuration management
├── examples/
│   └── plugins/           # Example plugins
├── pkg/
│   └── plugin/            # Public plugin API
└── docs/                  # Documentation
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and changes.

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/your-username/md-to-pdf/issues)
- 💬 [Discussions](https://github.com/your-username/md-to-pdf/discussions)

## Acknowledgments

- [goldmark](https://github.com/yuin/goldmark) - Markdown parser
- [gofpdf](https://github.com/jung-kurt/gofpdf) - PDF generation
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [mermaid](https://mermaid.js.org/) - Diagram generation