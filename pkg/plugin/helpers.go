package plugin

import (
	"strings"

	"github.com/yuin/goldmark/ast"
)

// Helper functions for common plugin development tasks

// IsCodeBlock checks if a node is a code block with the specified language
func IsCodeBlock(node ast.Node, language string, source []byte) bool {
	if node.Kind() != ast.KindFencedCodeBlock {
		return false
	}

	codeBlock := node.(*ast.FencedCodeBlock)
	if codeBlock.Info == nil {
		return language == ""
	}

	info := string(codeBlock.Info.Segment.Value(source))
	return strings.TrimSpace(info) == language
}

// GetCodeBlockContent extracts the content from a code block
func GetCodeBlockContent(node ast.Node, source []byte) string {
	if node.Kind() != ast.KindFencedCodeBlock && node.Kind() != ast.KindCodeBlock {
		return ""
	}

	var content strings.Builder

	if fencedBlock, ok := node.(*ast.FencedCodeBlock); ok {
		lines := fencedBlock.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			content.Write(line.Value(source))
		}
	} else if codeBlock, ok := node.(*ast.CodeBlock); ok {
		lines := codeBlock.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			content.Write(line.Value(source))
		}
	}

	return content.String()
}

// GetCodeBlockLanguage returns the language specified in a fenced code block
func GetCodeBlockLanguage(node ast.Node, source []byte) string {
	if node.Kind() != ast.KindFencedCodeBlock {
		return ""
	}

	fencedBlock := node.(*ast.FencedCodeBlock)
	if fencedBlock.Info == nil {
		return ""
	}

	info := string(fencedBlock.Info.Segment.Value(source))
	return strings.TrimSpace(strings.Split(info, " ")[0]) // Get first word only
}

// ReplaceNode replaces a node in the AST with a new node
func ReplaceNode(oldNode, newNode ast.Node) {
	parent := oldNode.Parent()
	if parent == nil {
		return
	}

	// Find the position of the old node
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		if child == oldNode {
			parent.ReplaceChild(parent, oldNode, newNode)
			break
		}
	}
}

// CreateParagraphWithText creates a new paragraph node with text content
func CreateParagraphWithText(text string) ast.Node {
	paragraph := ast.NewParagraph()
	// Note: Creating text nodes with actual content requires more complex setup
	// This is a simplified version - full implementation would need text segments
	return paragraph
}

// WalkChildNodes walks through all child nodes of a given node
func WalkChildNodes(node ast.Node, fn func(ast.Node) bool) {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if !fn(child) {
			break
		}
		WalkChildNodes(child, fn)
	}
}
