# MD-to-PDF

A powerful, extensible markdown to PDF converter written in Go with plugin support.

## Why md-to-pdf?

There are several tools for converting Markdown to PDF. Here's how MD-to-PDF compares:

| Feature | MD-to-PDF | Pandoc | wkhtmltopdf |
|---------|-----------|--------|-------------|
| Binary Size | ~12MB | ~50MB | ~100MB |
| Dependencies | None | System libs | System libs |
| Plugin System | Yes | Filters | No |
| Learning Curve | Low | High | Medium |
| Configuration | Simple YAML | Complex | Command flags |

**Key differentiators:**

- **Zero dependencies**: Single binary, no runtime requirements. Download and run.
- **Plugin architecture**: Extend functionality without modifying core code. Add custom transformations, content generators, and more.
- **Simple configuration**: Human-readable YAML config with sensible defaults. No need to memorize complex command-line flags.
- **Built for Markdown**: Purpose-built for Markdown-to-PDF conversion, not a general-purpose document converter.

## Features

- **Fast & reliable**: Built in Go for performance and reliability
- **Plugin system**: Extensible architecture with custom plugin support
- **Rich formatting**: Support for tables, code blocks, images, and more
- **Mermaid diagrams**: Built-in support for Mermaid diagram generation
- **Configurable**: Extensive customization options for fonts, margins, spacing
- **Table of contents**: Automatic TOC generation
- **CLI interface**: Easy-to-use command-line interface with sensible defaults

## Quick start

### Installation

#### Using the install script (recommended):
```bash
curl -sSL https://raw.githubusercontent.com/fredcamaral/md-to-pdf/main/install.sh | bash
```

#### Manual installation:
1. Download the latest release from [GitHub Releases](https://github.com/fredcamaral/md-to-pdf/releases)
2. Extract and move the binary to your PATH

#### Build from source:
```bash
git clone https://github.com/fredcamaral/md-to-pdf.git
cd md-to-pdf
make build
```

### Basic usage

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

## Plugin system

MD-to-PDF supports two types of plugins:

### AST transformers
Modify the markdown abstract syntax tree before rendering:
```go
type ASTTransformer interface {
    Transform(node ast.Node, ctx *TransformContext) (ast.Node, error)
    Priority() int
    SupportedNodes() []ast.NodeKind
}
```

### Content generators
Generate additional content during PDF creation:
```go
type ContentGenerator interface {
    GenerateContent(ctx *GenerationContext) error
    Priority() int
}
```

### Available plugins

- **Mermaid Plugin**: Converts mermaid code blocks to PNG diagrams
- **TOC Plugin**: Generates table of contents

### Loading plugins

Place plugin `.so` files in the `plugins/` directory and md-to-pdf loads them automatically.

**[Plugin Development Guide](plugins/README.md)** - Learn how to create custom plugins

## Configuration options

| Category | Option | Default | Description |
|----------|--------|---------|-------------|
| Font | `font.family` | "Arial" | Font family name |
| Font | `font.size` | 10 | Font size in points |
| Page | `page.size` | "A4" | Page size (A4, Letter, Legal) |
| Page | `page.margins` | "20,20,20,20" | Margins (top,right,bottom,left) |
| Text | `text.lineSpacing` | 1.2 | Line spacing multiplier |
| Mermaid | `mermaid.theme` | "default" | Mermaid theme |
| Mermaid | `mermaid.scale` | 2.2 | Mermaid diagram scale |

## CLI commands

### Convert command
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

### Config commands
```bash
md-to-pdf config list                    # List all configuration
md-to-pdf config set <key> <value>      # Set configuration value
md-to-pdf config reset                  # Reset to defaults
```

## Examples

### Basic conversion
```bash
# Convert with defaults
md-to-pdf convert README.md

# Convert with custom output name
md-to-pdf convert README.md -o documentation.pdf
```

### Custom styling
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

### Document metadata
```bash
md-to-pdf convert report.md \
  --title "Monthly Report" \
  --author "Jane Smith" \
  --subject "Business Analytics" \
  --keywords "report,analytics,monthly"
```

### Mermaid diagrams
```bash
# Custom mermaid settings
md-to-pdf convert flowchart.md \
  --mermaid-theme dark \
  --mermaid-scale 3.0
```

## Supported Markdown features

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

### Project structure
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
└── plugins/               # Plugin directory and development guide
```

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and changes.

## FAQ

### How do I convert multiple files?

You can use a shell loop or find command:
```bash
# Convert all markdown files in current directory
for f in *.md; do md-to-pdf convert "$f"; done

# Convert recursively
find . -name "*.md" -exec md-to-pdf convert {} \;
```

### How do I customize styling?

You can customize styling in two ways:

1. **Command-line flags** for one-time changes:
   ```bash
   md-to-pdf convert doc.md --font-size 12 --margins "30,30,30,30"
   ```

2. **Configuration file** for persistent settings:
   ```bash
   md-to-pdf config set font.family "Times New Roman"
   md-to-pdf config set text.fontSize 12
   ```

### How do I use plugins?

1. Place `.so` plugin files in the `plugins/` directory (or specify with `--plugins-dir`)
2. md-to-pdf loads plugins automatically on conversion
3. See the [Plugin Development Guide](plugins/README.md) for creating custom plugins

### Where is the config file stored?

The configuration file is stored at:
- **Linux/macOS**: `~/.config/md-to-pdf/config.yaml`
- **Windows**: `%APPDATA%\md-to-pdf\config.yaml`

View the current config location and values with:
```bash
md-to-pdf config list
```

### How do I reset configuration to defaults?

```bash
md-to-pdf config reset
```

This removes all custom settings and restores default settings.

## Support

- [Issue Tracker](https://github.com/fredcamaral/md-to-pdf/issues)
- [Discussions](https://github.com/fredcamaral/md-to-pdf/discussions)

## Acknowledgments

- [goldmark](https://github.com/yuin/goldmark) - Markdown parser
- [gofpdf](https://github.com/jung-kurt/gofpdf) - PDF generation
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [mermaid](https://mermaid.js.org/) - Diagram generation