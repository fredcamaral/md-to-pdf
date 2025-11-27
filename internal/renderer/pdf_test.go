package renderer

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fredcamaral/md-to-pdf/internal/plugins"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// createTestPNG creates a simple solid color image for testing
func createTestPNG(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	red := color.RGBA{255, 0, 0, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, red)
		}
	}
	return img
}

// writePNG writes an image as PNG to the given writer
func writePNG(w io.Writer, img image.Image) error {
	return png.Encode(w, img)
}

func defaultTestConfig() *RenderConfig {
	return &RenderConfig{
		PageSize:     "A4",
		FontFamily:   "Arial",
		FontSize:     12,
		HeadingScale: 1.5,
		LineSpacing:  1.2,
		CodeFont:     "Courier",
		CodeSize:     10,
		Margins: Margins{
			Top:    20,
			Bottom: 20,
			Left:   15,
			Right:  15,
		},
		Mermaid: MermaidConfig{
			Scale:     2.2,
			MaxWidth:  0,
			MaxHeight: 150.0,
		},
	}
}

func defaultTestDocumentMetadata() *DocumentMetadata {
	return &DocumentMetadata{
		Title:   "Test Document",
		Author:  "Test Author",
		Subject: "Test Subject",
	}
}

func createTestDocument(content string) (ast.Node, []byte) {
	source := []byte(content)
	reader := text.NewReader(source)

	// Simple document creation - create a basic AST structure
	doc := ast.NewDocument()

	// Parse basic structure manually for test purposes
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "# ") {
			heading := ast.NewHeading(1)
			textNode := ast.NewTextSegment(text.NewSegment(0, len(source)))
			heading.AppendChild(heading, textNode)
			doc.AppendChild(doc, heading)
		} else if strings.HasPrefix(line, "## ") {
			heading := ast.NewHeading(2)
			doc.AppendChild(doc, heading)
		} else {
			paragraph := ast.NewParagraph()
			doc.AppendChild(doc, paragraph)
		}
	}

	_ = reader // Used to create source
	return doc, source
}

func TestNewPDFRenderer(t *testing.T) {
	config := defaultTestConfig()
	document := defaultTestDocumentMetadata()
	pluginManager := plugins.NewManager("./plugins", false, nil)

	renderer := NewPDFRenderer(config, document, pluginManager)

	if renderer == nil {
		t.Fatal("NewPDFRenderer returned nil")
	}
	if renderer.config == nil {
		t.Error("renderer config should not be nil")
	}
	if renderer.plugins == nil {
		t.Error("renderer plugins should not be nil")
	}
}

func TestRender_BasicDocument(t *testing.T) {
	config := defaultTestConfig()
	document := defaultTestDocumentMetadata()
	pluginManager := plugins.NewManager("./plugins", false, nil)
	renderer := NewPDFRenderer(config, document, pluginManager)

	// Create a simple markdown document
	markdown := `# Hello World

This is a test paragraph.`

	node, source := createTestDocument(markdown)

	buf, err := renderer.Render(node, source)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if buf == nil {
		t.Fatal("Render returned nil buffer")
	}

	if buf.Len() == 0 {
		t.Error("Render returned empty buffer")
	}
}

func TestRender_CodeBlock(t *testing.T) {
	config := defaultTestConfig()
	document := defaultTestDocumentMetadata()
	pluginManager := plugins.NewManager("./plugins", false, nil)
	renderer := NewPDFRenderer(config, document, pluginManager)

	// Create document with code block
	doc := ast.NewDocument()

	// Create a fenced code block
	codeBlock := ast.NewFencedCodeBlock(nil)

	// We need to set up the segments for the code block
	source := []byte("```go\nfunc main() {}\n```")

	doc.AppendChild(doc, codeBlock)

	buf, err := renderer.Render(doc, source)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if buf == nil {
		t.Fatal("Render returned nil buffer")
	}
}

func TestRender_OutputIsValidPDF(t *testing.T) {
	config := defaultTestConfig()
	document := defaultTestDocumentMetadata()
	pluginManager := plugins.NewManager("./plugins", false, nil)
	renderer := NewPDFRenderer(config, document, pluginManager)

	// Create a simple document
	doc := ast.NewDocument()
	paragraph := ast.NewParagraph()
	doc.AppendChild(doc, paragraph)

	source := []byte("Test content")

	buf, err := renderer.Render(doc, source)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// PDF files should start with %PDF
	pdfContent := buf.Bytes()
	if len(pdfContent) < 4 {
		t.Fatal("PDF content too short")
	}

	pdfHeader := string(pdfContent[:4])
	if pdfHeader != "%PDF" {
		t.Errorf("PDF should start with %%PDF, got %q", pdfHeader)
	}
}

func TestRender_WithMermaidImage(t *testing.T) {
	// Create a temporary image file using Go's image package to generate a valid PNG
	tempDir := t.TempDir()
	imagePath := filepath.Join(tempDir, "mermaid.png")

	// Create a simple 10x10 red PNG image programmatically
	img := createTestPNG(10, 10)
	f, err := os.Create(imagePath)
	if err != nil {
		t.Fatalf("failed to create test image file: %v", err)
	}
	defer f.Close()

	if err := writePNG(f, img); err != nil {
		t.Fatalf("failed to write PNG: %v", err)
	}

	config := defaultTestConfig()
	document := defaultTestDocumentMetadata()
	pluginManager := plugins.NewManager("./plugins", false, nil)
	renderer := NewPDFRenderer(config, document, pluginManager)

	// Create document with mermaid image paragraph
	doc := ast.NewDocument()
	paragraph := ast.NewParagraph()
	paragraph.SetAttribute([]byte("data-mermaid-image"), []byte(imagePath))
	doc.AppendChild(doc, paragraph)

	source := []byte("")

	buf, err := renderer.Render(doc, source)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if buf == nil {
		t.Fatal("Render returned nil buffer")
	}

	// Verify it's a valid PDF
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
		t.Error("output should be a valid PDF")
	}
}

func TestRender_WithMermaidImage_MissingFile(t *testing.T) {
	config := defaultTestConfig()
	document := defaultTestDocumentMetadata()
	pluginManager := plugins.NewManager("./plugins", false, nil)
	renderer := NewPDFRenderer(config, document, pluginManager)

	// Create document with mermaid image paragraph pointing to nonexistent file
	doc := ast.NewDocument()
	paragraph := ast.NewParagraph()
	paragraph.SetAttribute([]byte("data-mermaid-image"), []byte("/nonexistent/image.png"))
	doc.AppendChild(doc, paragraph)

	source := []byte("")

	// Should not fail, but should render a fallback
	buf, err := renderer.Render(doc, source)
	if err != nil {
		t.Fatalf("Render should not fail for missing mermaid image: %v", err)
	}

	if buf == nil {
		t.Fatal("Render returned nil buffer")
	}
}

func TestApplyTransformers(t *testing.T) {
	config := defaultTestConfig()
	document := defaultTestDocumentMetadata()

	// Create a mock transformer for testing
	mockManager := &mockPluginManager{
		transformers: []plugins.ASTTransformer{
			&mockTransformer{
				name:     "test-transformer",
				priority: 100,
				transformFunc: func(node ast.Node, ctx *plugins.TransformContext) (ast.Node, error) {
					// Simply return the node unchanged
					return node, nil
				},
			},
		},
	}

	renderer := &PDFRenderer{
		config:   config,
		document: document,
		plugins:  createManagerWithTransformers(mockManager.transformers),
	}

	// Create a simple document
	doc := ast.NewDocument()
	paragraph := ast.NewParagraph()
	doc.AppendChild(doc, paragraph)

	source := []byte("Test")
	ctx := &plugins.TransformContext{
		CurrentNode: paragraph,
		Parent:      doc,
		Source:      source,
		Metadata:    make(map[string]interface{}),
		Config:      make(map[string]interface{}),
	}

	// Apply transformers
	result, err := renderer.plugins.ApplyTransformers(paragraph, ctx)
	if err != nil {
		t.Fatalf("ApplyTransformers failed: %v", err)
	}

	if result == nil {
		t.Error("ApplyTransformers returned nil")
	}
}

func TestRender_DifferentPageSizes(t *testing.T) {
	pageSizes := []string{"A4", "A3", "A5", "Letter", "Legal"}

	document := defaultTestDocumentMetadata()
	pluginManager := plugins.NewManager("./plugins", false, nil)

	for _, pageSize := range pageSizes {
		t.Run(pageSize, func(t *testing.T) {
			config := defaultTestConfig()
			config.PageSize = pageSize

			renderer := NewPDFRenderer(config, document, pluginManager)

			doc := ast.NewDocument()
			source := []byte("")

			buf, err := renderer.Render(doc, source)
			if err != nil {
				t.Fatalf("Render failed for page size %s: %v", pageSize, err)
			}

			if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
				t.Errorf("output should be a valid PDF for page size %s", pageSize)
			}
		})
	}
}

func TestRender_DifferentMargins(t *testing.T) {
	tests := []struct {
		name    string
		margins Margins
	}{
		{
			name:    "default_margins",
			margins: Margins{Top: 20, Bottom: 20, Left: 15, Right: 15},
		},
		{
			name:    "zero_margins",
			margins: Margins{Top: 0, Bottom: 0, Left: 0, Right: 0},
		},
		{
			name:    "large_margins",
			margins: Margins{Top: 50, Bottom: 50, Left: 50, Right: 50},
		},
		{
			name:    "asymmetric_margins",
			margins: Margins{Top: 10, Bottom: 30, Left: 20, Right: 40},
		},
	}

	document := defaultTestDocumentMetadata()
	pluginManager := plugins.NewManager("./plugins", false, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := defaultTestConfig()
			config.Margins = tt.margins

			renderer := NewPDFRenderer(config, document, pluginManager)

			doc := ast.NewDocument()
			source := []byte("")

			buf, err := renderer.Render(doc, source)
			if err != nil {
				t.Fatalf("Render failed with margins %+v: %v", tt.margins, err)
			}

			if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
				t.Errorf("output should be a valid PDF with margins %+v", tt.margins)
			}
		})
	}
}

func TestRender_DifferentFontSizes(t *testing.T) {
	fontSizes := []float64{8, 10, 12, 14, 16, 18, 24}

	document := defaultTestDocumentMetadata()
	pluginManager := plugins.NewManager("./plugins", false, nil)

	for _, fontSize := range fontSizes {
		t.Run(string(rune(int(fontSize))), func(t *testing.T) {
			config := defaultTestConfig()
			config.FontSize = fontSize

			renderer := NewPDFRenderer(config, document, pluginManager)

			doc := ast.NewDocument()
			paragraph := ast.NewParagraph()
			doc.AppendChild(doc, paragraph)
			source := []byte("Test text")

			buf, err := renderer.Render(doc, source)
			if err != nil {
				t.Fatalf("Render failed with font size %v: %v", fontSize, err)
			}

			if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
				t.Errorf("output should be a valid PDF with font size %v", fontSize)
			}
		})
	}
}

func TestRender_MultipleHeadings(t *testing.T) {
	config := defaultTestConfig()
	document := defaultTestDocumentMetadata()
	pluginManager := plugins.NewManager("./plugins", false, nil)
	renderer := NewPDFRenderer(config, document, pluginManager)

	// Create document with multiple heading levels
	doc := ast.NewDocument()

	for level := 1; level <= 6; level++ {
		heading := ast.NewHeading(level)
		doc.AppendChild(doc, heading)
	}

	source := []byte("# H1\n## H2\n### H3\n#### H4\n##### H5\n###### H6")

	buf, err := renderer.Render(doc, source)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
		t.Error("output should be a valid PDF")
	}
}

func TestRender_EmptyDocument(t *testing.T) {
	config := defaultTestConfig()
	document := defaultTestDocumentMetadata()
	pluginManager := plugins.NewManager("./plugins", false, nil)
	renderer := NewPDFRenderer(config, document, pluginManager)

	doc := ast.NewDocument()
	source := []byte("")

	buf, err := renderer.Render(doc, source)
	if err != nil {
		t.Fatalf("Render failed for empty document: %v", err)
	}

	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
		t.Error("output should be a valid PDF even for empty document")
	}
}

func TestRender_MermaidScaling(t *testing.T) {
	tests := []struct {
		name      string
		scale     float64
		maxWidth  float64
		maxHeight float64
	}{
		{
			name:      "default_scale",
			scale:     2.2,
			maxWidth:  0,
			maxHeight: 150,
		},
		{
			name:      "small_scale",
			scale:     1.0,
			maxWidth:  100,
			maxHeight: 100,
		},
		{
			name:      "large_scale",
			scale:     3.0,
			maxWidth:  200,
			maxHeight: 200,
		},
	}

	document := defaultTestDocumentMetadata()
	pluginManager := plugins.NewManager("./plugins", false, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := defaultTestConfig()
			config.Mermaid = MermaidConfig{
				Scale:     tt.scale,
				MaxWidth:  tt.maxWidth,
				MaxHeight: tt.maxHeight,
			}

			renderer := NewPDFRenderer(config, document, pluginManager)

			doc := ast.NewDocument()
			source := []byte("")

			buf, err := renderer.Render(doc, source)
			if err != nil {
				t.Fatalf("Render failed with mermaid config %+v: %v", tt, err)
			}

			if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
				t.Error("output should be a valid PDF")
			}
		})
	}
}

// Helper types for testing

type mockPluginManager struct {
	transformers []plugins.ASTTransformer
}

type mockTransformer struct {
	name           string
	version        string
	description    string
	priority       int
	supportedNodes []ast.NodeKind
	transformFunc  func(ast.Node, *plugins.TransformContext) (ast.Node, error)
}

func (m *mockTransformer) Name() string                                { return m.name }
func (m *mockTransformer) Version() string                             { return m.version }
func (m *mockTransformer) Description() string                         { return m.description }
func (m *mockTransformer) Init(config map[string]interface{}) error    { return nil }
func (m *mockTransformer) Cleanup() error                              { return nil }
func (m *mockTransformer) Priority() int                               { return m.priority }
func (m *mockTransformer) SupportedNodes() []ast.NodeKind              { return m.supportedNodes }
func (m *mockTransformer) Transform(node ast.Node, ctx *plugins.TransformContext) (ast.Node, error) {
	if m.transformFunc != nil {
		return m.transformFunc(node, ctx)
	}
	return node, nil
}

func createManagerWithTransformers(transformers []plugins.ASTTransformer) *plugins.Manager {
	manager := plugins.NewManager("./plugins", true, nil)
	// Note: We can't directly inject transformers, so we test through the public API
	return manager
}
