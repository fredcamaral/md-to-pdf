package parser

import (
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type MarkdownParser struct {
	goldmark goldmark.Markdown
}

func NewMarkdownParser() *MarkdownParser {
	md := goldmark.New(
		goldmark.WithExtensions(),
	)

	return &MarkdownParser{
		goldmark: md,
	}
}

func (p *MarkdownParser) Parse(content []byte) (ast.Node, error) {
	reader := text.NewReader(content)
	return p.goldmark.Parser().Parse(reader), nil
}

func (p *MarkdownParser) ParseFile(path string) (ast.Node, error) {
	content, err := os.ReadFile(path) // #nosec G304 - file path comes from user CLI input
	if err != nil {
		return nil, err
	}
	return p.Parse(content)
}
