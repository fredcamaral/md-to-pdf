package core

import (
	"time"

	"github.com/yuin/goldmark/ast"
)

type Document struct {
	Title      string
	Author     string
	Subject    string
	Keywords   []string
	Content    ast.Node
	Metadata   map[string]interface{}
	SourceFile string
	CreatedAt  time.Time
}

type Config struct {
	Parser   ParserConfig
	Renderer RenderConfig
	Plugins  PluginConfig
	Output   OutputConfig
	Document DocumentConfig
}

type ParserConfig struct {
	Extensions []string
}

type RenderConfig struct {
	PageSize     string
	Margins      Margins
	FontFamily   string
	FontSize     float64
	HeadingScale float64
	LineSpacing  float64
	CodeFont     string
	CodeSize     float64
	Mermaid      MermaidConfig
}

type MermaidConfig struct {
	Scale     float64 // Scaling factor for mermaid diagrams (1.0 = normal, 1.4 = 40% bigger)
	MaxWidth  float64 // Maximum width in mm (0 = use page width)
	MaxHeight float64 // Maximum height in mm
}

type PluginConfig struct {
	Directory string
	Enabled   bool
}

type OutputConfig struct {
	Path    string
	Quality string
}

type DocumentConfig struct {
	Title   string
	Author  string
	Subject string
}

type Margins struct {
	Top    float64
	Bottom float64
	Left   float64
	Right  float64
}

type ConversionOptions struct {
	InputFiles []string
	OutputPath string
	PluginDir  string
	Verbose    bool
}