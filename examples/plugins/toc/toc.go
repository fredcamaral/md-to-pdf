package main

import (
	"fmt"

	"github.com/fredcamaral/md-to-pdf/pkg/plugin"
	"github.com/yuin/goldmark/ast"
)

// TOCPlugin generates a table of contents
type TOCPlugin struct {
	*plugin.BasePlugin
	headings []string
}

// NewPlugin is the required entry point for plugins
func NewPlugin() plugin.Plugin {
	return &TOCPlugin{
		BasePlugin: plugin.NewBasePlugin(
			"toc",
			"1.0.0",
			"Generates table of contents from headings",
		),
		headings: make([]string, 0),
	}
}

// Implement ASTTransformer interface
func (p *TOCPlugin) Transform(node ast.Node, ctx *plugin.TransformContext) (ast.Node, error) {
	// Collect headings for TOC generation
	if node.Kind() == ast.KindHeading {
		heading := node.(*ast.Heading)
		text := plugin.ExtractText(node, ctx.Source)
		p.headings = append(p.headings, fmt.Sprintf("%s %s",
			getHeadingPrefix(heading.Level), text))
	}

	return node, nil
}

func (p *TOCPlugin) Priority() int {
	return 10 // Lower priority, runs after other transformers
}

func (p *TOCPlugin) SupportedNodes() []ast.NodeKind {
	return []ast.NodeKind{ast.KindHeading}
}

// Implement ContentGenerator interface
func (p *TOCPlugin) Generate(ctx *plugin.RenderContext) ([]plugin.PDFElement, error) {
	if len(p.headings) == 0 {
		return nil, nil
	}

	var elements []plugin.PDFElement

	// Add TOC title
	elements = append(elements, plugin.CreateTextElement("Table of Contents", 16, "B"))
	elements = append(elements, plugin.CreateLineElement(0, 0, 100, 0, 0.5))

	// Add headings
	for _, heading := range p.headings {
		elements = append(elements, plugin.CreateTextElement(heading, 12, ""))
	}

	// Add spacing after TOC
	elements = append(elements, plugin.CreateTextElement("", 12, ""))

	return elements, nil
}

func (p *TOCPlugin) GenerationPhase() plugin.GenerationPhase {
	return plugin.BeforeContent
}

func getHeadingPrefix(level int) string {
	prefix := ""
	for i := 1; i < level; i++ {
		prefix += "  "
	}
	return prefix + "•"
}

// main is required for Go plugin compilation with -buildmode=plugin
func main() {}
