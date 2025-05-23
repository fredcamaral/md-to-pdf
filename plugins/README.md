# Plugin Development Guide

This directory contains plugins for MD-to-PDF. This guide will help you develop custom plugins to extend the functionality of the markdown to PDF converter.

## Quick Start

1. **Check out existing plugins** in `examples/plugins/` for reference
2. **Choose plugin type** - AST Transformer or Content Generator
3. **Create plugin directory** with your plugin code
4. **Build plugin** using `make build-plugins` or manually
5. **Test plugin** by placing `.so` file in this directory

## Plugin Types

MD-to-PDF supports two types of plugins:

### AST Transformers
Modify the markdown Abstract Syntax Tree before PDF rendering.

**Use cases:**
- Transform custom markdown syntax
- Modify existing elements (e.g., convert code blocks to diagrams)
- Add metadata to nodes
- Filter or replace content

**Interface:**
```go
type ASTTransformer interface {
    Plugin
    Transform(node ast.Node, ctx *TransformContext) (ast.Node, error)
    Priority() int
    SupportedNodes() []ast.NodeKind
}
```

### Content Generators
Generate additional content during PDF creation.

**Use cases:**
- Add table of contents
- Insert headers/footers
- Generate appendices
- Add watermarks or stamps

**Interface:**
```go
type ContentGenerator interface {
    Plugin
    GenerateContent(ctx *GenerationContext) error
    Priority() int
}
```

## Plugin Development

### 1. Project Structure

Create a new directory for your plugin:
```
plugins/
└── myplugin/
    ├── myplugin.go      # Main plugin implementation
    ├── go.mod           # Module definition
    └── README.md        # Plugin documentation
```

### 2. Basic Plugin Template

Every plugin must implement the base `Plugin` interface:

```go
package main

import (
    "github.com/fredcamaral/md-to-pdf/pkg/plugin"
)

type MyPlugin struct{}

// Required Plugin interface methods
func (p *MyPlugin) Name() string {
    return "myplugin"
}

func (p *MyPlugin) Version() string {
    return "1.0.0"
}

func (p *MyPlugin) Description() string {
    return "My custom plugin description"
}

// Required plugin creation function
func NewPlugin() plugin.Plugin {
    return &MyPlugin{}
}
```

### 3. AST Transformer Example

```go
package main

import (
    "github.com/yuin/goldmark/ast"
    "github.com/fredcamaral/md-to-pdf/pkg/plugin"
)

type CustomTransformer struct{}

func (t *CustomTransformer) Name() string { return "custom-transformer" }
func (t *CustomTransformer) Version() string { return "1.0.0" }
func (t *CustomTransformer) Description() string { return "Transforms custom elements" }

// AST Transformer specific methods
func (t *CustomTransformer) Transform(node ast.Node, ctx *plugin.TransformContext) (ast.Node, error) {
    // Implement your transformation logic here
    // Example: convert specific code blocks to custom elements
    
    if codeBlock, ok := node.(*ast.FencedCodeBlock); ok {
        language := string(codeBlock.Language(ctx.Source))
        if language == "custom" {
            // Transform this code block
            // Return modified node or new node
        }
    }
    
    return node, nil
}

func (t *CustomTransformer) Priority() int {
    return 100 // Higher number = higher priority
}

func (t *CustomTransformer) SupportedNodes() []ast.NodeKind {
    return []ast.NodeKind{
        ast.KindFencedCodeBlock,
        ast.KindCodeBlock,
    }
}

func NewPlugin() plugin.Plugin {
    return &CustomTransformer{}
}
```

### 4. Content Generator Example

```go
package main

import (
    "github.com/fredcamaral/md-to-pdf/pkg/plugin"
)

type HeaderFooterGenerator struct{}

func (g *HeaderFooterGenerator) Name() string { return "header-footer" }
func (g *HeaderFooterGenerator) Version() string { return "1.0.0" }
func (g *HeaderFooterGenerator) Description() string { return "Adds headers and footers" }

// Content Generator specific methods
func (g *HeaderFooterGenerator) GenerateContent(ctx *plugin.GenerationContext) error {
    // Access PDF instance
    pdf := ctx.PDF
    
    // Add header
    pdf.SetHeaderFunc(func() {
        pdf.SetY(15)
        pdf.SetFont("Arial", "B", 15)
        pdf.Cell(0, 10, "Document Header")
        pdf.Ln(20)
    })
    
    // Add footer
    pdf.SetFooterFunc(func() {
        pdf.SetY(-15)
        pdf.SetFont("Arial", "I", 8)
        pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pdf.PageNo()),
            "", 0, "C", false, 0, "")
    })
    
    return nil
}

func (g *HeaderFooterGenerator) Priority() int {
    return 50
}

func NewPlugin() plugin.Plugin {
    return &HeaderFooterGenerator{}
}
```

### 5. Building Plugins

#### Using Make (Recommended)
```bash
# Build all plugins
make build-plugins

# Build specific plugin
cd examples/plugins/myplugin
go build -buildmode=plugin -o ../../../plugins/myplugin.so .
```

#### Manual Build
```bash
cd your-plugin-directory
go build -buildmode=plugin -o ../../plugins/myplugin.so .
```

### 6. Plugin Configuration

Plugins can access configuration through the context:

```go
func (t *MyTransformer) Transform(node ast.Node, ctx *plugin.TransformContext) (ast.Node, error) {
    // Access configuration
    config := ctx.Config
    
    // Check for plugin-specific config
    if customSetting, ok := config.PluginSettings["myplugin.setting"]; ok {
        // Use custom setting
    }
    
    return node, nil
}
```

## Available Context

### TransformContext (AST Transformers)
```go
type TransformContext struct {
    Config   *core.Config    // Application configuration
    Source   []byte          // Original markdown source
    Document ast.Node        // Root document node
    Metadata map[string]any  // Document metadata
}
```

### GenerationContext (Content Generators)
```go
type GenerationContext struct {
    Config   *core.Config    // Application configuration
    PDF      *gofpdf.Fpdf    // PDF instance
    Document ast.Node        // Parsed document
    Metadata map[string]any  // Document metadata
}
```

## Best Practices

### Error Handling
```go
func (t *MyTransformer) Transform(node ast.Node, ctx *plugin.TransformContext) (ast.Node, error) {
    if err := validateInput(node); err != nil {
        return nil, fmt.Errorf("myplugin: invalid input: %w", err)
    }
    
    // ... transformation logic
    
    return node, nil
}
```

### Resource Management
```go
func (g *MyGenerator) GenerateContent(ctx *plugin.GenerationContext) error {
    // Open resources
    file, err := os.Open("resource.txt")
    if err != nil {
        return fmt.Errorf("myplugin: failed to open resource: %w", err)
    }
    defer file.Close() // Always clean up
    
    // ... generation logic
    
    return nil
}
```

### Configuration
```go
type PluginConfig struct {
    Enabled   bool   `yaml:"enabled"`
    Theme     string `yaml:"theme"`
    Scale     float64 `yaml:"scale"`
}

func (p *MyPlugin) loadConfig(config *core.Config) *PluginConfig {
    // Load plugin-specific configuration
    return &PluginConfig{
        Enabled: true,
        Theme:   "default",
        Scale:   1.0,
    }
}
```

## Testing Plugins

### Unit Testing
```go
func TestMyTransformer_Transform(t *testing.T) {
    transformer := &MyTransformer{}
    
    // Create test node
    node := ast.NewFencedCodeBlock(nil)
    
    // Create test context
    ctx := &plugin.TransformContext{
        Config: &core.Config{},
        Source: []byte("test markdown"),
    }
    
    // Test transformation
    result, err := transformer.Transform(node, ctx)
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Integration Testing
```bash
# Build plugin
go build -buildmode=plugin -o plugins/myplugin.so .

# Test with sample markdown
echo "# Test\n\`\`\`custom\ntest content\n\`\`\`" | md-to-pdf convert - --plugins-dir ./plugins
```

## Example Plugins

### Mermaid Plugin
Converts mermaid code blocks to PNG diagrams.

**Location:** `examples/plugins/mermaid/`
**Type:** AST Transformer
**Features:**
- Detects mermaid code blocks
- Generates PNG diagrams using mermaid CLI
- Embeds images in PDF

### TOC Plugin  
Generates table of contents from headers.

**Location:** `examples/plugins/toc/`
**Type:** Content Generator
**Features:**
- Scans document for headers
- Generates formatted TOC
- Adds page references

## Plugin API Reference

### Core Interfaces

```go
// Base plugin interface
type Plugin interface {
    Name() string
    Version() string
    Description() string
}

// AST transformation plugin
type ASTTransformer interface {
    Plugin
    Transform(node ast.Node, ctx *TransformContext) (ast.Node, error)
    Priority() int
    SupportedNodes() []ast.NodeKind
}

// Content generation plugin
type ContentGenerator interface {
    Plugin
    GenerateContent(ctx *GenerationContext) error
    Priority() int
}
```

### Helper Functions

```go
// Check if node has specific attribute
func HasAttribute(node ast.Node, name []byte) bool

// Get attribute value
func GetAttribute(node ast.Node, name []byte) []byte

// Set attribute on node
func SetAttribute(node ast.Node, name, value []byte)

// Create new paragraph with text
func NewParagraph(text string) *ast.Paragraph

// Create new image node
func NewImage(src, alt string) *ast.Image
```

## Debugging Plugins

### Enable Verbose Logging
```bash
md-to-pdf convert document.md -v
```

### Debug Output
```go
func (p *MyPlugin) Transform(node ast.Node, ctx *plugin.TransformContext) (ast.Node, error) {
    fmt.Printf("[DEBUG] %s: processing node %T\n", p.Name(), node)
    
    // ... transformation logic
    
    return node, nil
}
```

### Common Issues

1. **Plugin not loading**
   - Check file permissions
   - Verify `.so` extension
   - Ensure `NewPlugin()` function exists

2. **Build errors**
   - Use correct Go version (1.21+)
   - Check import paths
   - Verify `buildmode=plugin`

3. **Runtime errors**
   - Check plugin interface implementation
   - Validate input parameters
   - Handle errors gracefully

## Distribution

### Package Structure
```
myplugin-v1.0.0/
├── myplugin.so          # Compiled plugin
├── README.md            # Plugin documentation
├── examples/            # Usage examples
└── config.yaml          # Default configuration
```

### Installation
Users can install plugins by:
1. Downloading `.so` file to `plugins/` directory
2. Building from source
3. Using package managers (future)

## Contributing

When contributing plugins to the main repository:

1. **Follow coding standards** from CONTRIBUTING.md
2. **Add comprehensive tests** for your plugin
3. **Document functionality** with examples
4. **Update this README** if adding new patterns

## Support

- **Examples:** Check `examples/plugins/` for working examples
- **API Docs:** See `pkg/plugin/` for interface definitions
- **Issues:** Report bugs in the main repository
- **Discussions:** Use GitHub Discussions for questions

---

Happy plugin development! 🔌