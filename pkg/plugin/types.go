package plugin

import (
	"github.com/fredcamaral/md-to-pdf/internal/plugins"
	"github.com/jung-kurt/gofpdf"
	"github.com/yuin/goldmark/ast"
)

// Re-export types for plugin developers
type Plugin = plugins.Plugin
type ASTTransformer = plugins.ASTTransformer
type ContentGenerator = plugins.ContentGenerator
type TransformContext = plugins.TransformContext
type RenderContext = plugins.RenderContext
type Document = plugins.Document
type PDFElement = plugins.PDFElement
type GenerationPhase = plugins.GenerationPhase

// Re-export constants
const (
	BeforeContent  = plugins.BeforeContent
	AfterContent   = plugins.AfterContent
	BeforeEachPage = plugins.BeforeEachPage
	AfterEachPage  = plugins.AfterEachPage
)

// Re-export built-in elements
type TextElement = plugins.TextElement
type ImageElement = plugins.ImageElement
type LineElement = plugins.LineElement

// BasePlugin provides a basic implementation of the Plugin interface
type BasePlugin struct {
	name        string
	version     string
	description string
}

func NewBasePlugin(name, version, description string) *BasePlugin {
	return &BasePlugin{
		name:        name,
		version:     version,
		description: description,
	}
}

func (p *BasePlugin) Name() string {
	return p.name
}

func (p *BasePlugin) Version() string {
	return p.version
}

func (p *BasePlugin) Description() string {
	return p.description
}

func (p *BasePlugin) Init(config map[string]interface{}) error {
	return nil
}

func (p *BasePlugin) Cleanup() error {
	return nil
}

// Helper functions for plugin developers

// ExtractText extracts text content from an AST node
func ExtractText(node ast.Node, source []byte) string {
	if textNode, ok := node.(*ast.Text); ok {
		return string(textNode.Segment.Value(source))
	}
	
	var text string
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindText {
			if textNode, ok := n.(*ast.Text); ok {
				text += string(textNode.Segment.Value(source))
			}
		}
		return ast.WalkContinue, nil
	})
	
	return text
}

// CreateTextElement creates a new text element for PDF generation
func CreateTextElement(content string, fontSize float64, style string) PDFElement {
	return &TextElement{
		Content:  content,
		FontSize: fontSize,
		Style:    style,
	}
}

// CreateLineElement creates a new line element for PDF generation
func CreateLineElement(x1, y1, x2, y2, width float64) PDFElement {
	return &LineElement{
		X1:        x1,
		Y1:        y1,
		X2:        x2,
		Y2:        y2,
		LineWidth: width,
	}
}

// GetCurrentPosition returns the current position in the PDF
func GetCurrentPosition(pdf *gofpdf.Fpdf) (float64, float64) {
	return pdf.GetXY()
}