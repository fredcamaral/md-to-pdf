package plugins

import (
	"github.com/jung-kurt/gofpdf"
	"github.com/yuin/goldmark/ast"
)

// Base plugin interface
type Plugin interface {
	Name() string
	Version() string
	Description() string
	Init(config map[string]interface{}) error
	Cleanup() error
}

// AST transformation capability
type ASTTransformer interface {
	Plugin
	Transform(node ast.Node, ctx *TransformContext) (ast.Node, error)
	Priority() int
	SupportedNodes() []ast.NodeKind
}

// PDF content generation capability
type ContentGenerator interface {
	Plugin
	Generate(ctx *RenderContext) ([]PDFElement, error)
	GenerationPhase() GenerationPhase
}

// Plugin metadata
type PluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	License     string `json:"license"`
}

// Generation phases for content generators
type GenerationPhase int

const (
	BeforeContent GenerationPhase = iota
	AfterContent
	BeforeEachPage
	AfterEachPage
)

// Transform context for AST transformers
type TransformContext struct {
	Document    *Document
	CurrentNode ast.Node
	Parent      ast.Node
	Source      []byte
	Metadata    map[string]interface{}
	Config      map[string]interface{}
}

// RenderMargins represents page margins for rendering
type RenderMargins struct {
	Top    float64
	Bottom float64
	Left   float64
	Right  float64
}

// Render context for content generators
type RenderContext struct {
	Document    *Document
	CurrentPage int
	PDF         *gofpdf.Fpdf
	Source      []byte
	PageWidth   float64
	PageHeight  float64
	Margins     RenderMargins
	Metadata    map[string]interface{}
	Config      map[string]interface{}
}

// Document metadata
type Document struct {
	Title      string
	Author     string
	Subject    string
	Keywords   []string
	Metadata   map[string]interface{}
	SourceFile string
}

// PDF element interface for plugin-generated content
type PDFElement interface {
	Render(pdf *gofpdf.Fpdf, ctx *RenderContext) error
	Height() float64
	Width() float64
}
