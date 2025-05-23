package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fredcamaral/md-to-pdf/pkg/plugin"
	"github.com/yuin/goldmark/ast"
)

// MermaidPlugin transforms mermaid code blocks into diagram images
type MermaidPlugin struct {
	*plugin.BasePlugin
	outputDir  string
	images     []ImageInfo // Store images to embed
}

type ImageInfo struct {
	OriginalNode ast.Node
	FilePath     string
	Content      string
}

// NewPlugin is the required entry point for plugins
func NewPlugin() plugin.Plugin {
	return &MermaidPlugin{
		BasePlugin: plugin.NewBasePlugin(
			"mermaid",
			"1.0.0",
			"Converts mermaid code blocks to diagram images",
		),
		outputDir: "./mermaid-output",
		images:    make([]ImageInfo, 0),
	}
}

func (p *MermaidPlugin) Init(config map[string]interface{}) error {
	// Create output directory for mermaid diagrams
	err := os.MkdirAll(p.outputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create mermaid output directory: %w", err)
	}
	
	// Check if mermaid CLI is available
	_, err = exec.LookPath("mmdc")
	if err != nil {
		fmt.Println("Warning: mermaid CLI (mmdc) not found. Install with: npm install -g @mermaid-js/mermaid-cli")
		fmt.Println("Mermaid blocks will be rendered as placeholders")
	}
	
	return nil
}

// Implement ASTTransformer interface
func (p *MermaidPlugin) Transform(node ast.Node, ctx *plugin.TransformContext) (ast.Node, error) {
	// Only process fenced code blocks
	if node.Kind() != ast.KindFencedCodeBlock {
		return node, nil
	}
	
	language := plugin.GetCodeBlockLanguage(node, ctx.Source)
	
	// Only process mermaid blocks
	if language != "mermaid" {
		return node, nil
	}
	
	content := plugin.GetCodeBlockContent(node, ctx.Source)
	if content == "" {
		return node, nil
	}
	
	// Generate diagram
	imagePath, err := p.generateDiagram(content)
	if err != nil {
		// If diagram generation fails, return original node with error info
		fmt.Printf("Warning: failed to generate mermaid diagram: %v\n", err)
		return node, nil
	}
	
	// Store image info for later embedding
	p.images = append(p.images, ImageInfo{
		OriginalNode: node,
		FilePath:     imagePath,
		Content:      content,
	})
	
	// Create a special marker paragraph that the renderer can recognize
	paragraph := ast.NewParagraph()
	
	fmt.Printf("Generated mermaid diagram: %s\n", imagePath)
	
	// Store the marker in the paragraph's attributes for the renderer to find
	paragraph.SetAttribute([]byte("data-mermaid-image"), []byte(imagePath))
	
	return paragraph, nil
}

func (p *MermaidPlugin) Priority() int {
	return 5 // High priority, should run early
}

func (p *MermaidPlugin) SupportedNodes() []ast.NodeKind {
	return []ast.NodeKind{ast.KindFencedCodeBlock}
}

// Note: We don't implement ContentGenerator anymore since we're embedding 
// images directly during AST transformation via paragraph attributes

func (p *MermaidPlugin) generateDiagram(content string) (string, error) {
	// Generate a unique filename based on content hash
	hash := md5.Sum([]byte(content))
	filename := fmt.Sprintf("mermaid-%x.png", hash)
	outputPath := filepath.Join(p.outputDir, filename)
	
	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		return outputPath, nil
	}
	
	// Try to use mermaid CLI if available
	if err := p.generateWithCLI(content, outputPath); err == nil {
		return outputPath, nil
	}
	
	// Fallback: create a placeholder file
	return p.createPlaceholder(content, outputPath)
}

func (p *MermaidPlugin) generateWithCLI(content, outputPath string) error {
	// Check if mmdc is available
	_, err := exec.LookPath("mmdc")
	if err != nil {
		return err
	}
	
	// Create temporary input file
	tempInput := filepath.Join(p.outputDir, "temp.mmd")
	err = os.WriteFile(tempInput, []byte(content), 0644)
	if err != nil {
		return err
	}
	defer os.Remove(tempInput)
	
	// Run mermaid CLI
	cmd := exec.Command("mmdc", "-i", tempInput, "-o", outputPath, "-b", "white")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mermaid CLI failed: %w, output: %s", err, output)
	}
	
	return nil
}

func (p *MermaidPlugin) createPlaceholder(content, outputPath string) (string, error) {
	// Create a simple text file as placeholder
	placeholderContent := fmt.Sprintf("Mermaid Diagram Placeholder\n\nContent:\n%s\n\nTo generate actual diagrams, install mermaid CLI:\n npm install -g @mermaid-js/mermaid-cli", content)
	
	placeholderPath := outputPath + ".txt"
	err := os.WriteFile(placeholderPath, []byte(placeholderContent), 0644)
	if err != nil {
		return "", err
	}
	
	return placeholderPath, nil
}

func (p *MermaidPlugin) Cleanup() error {
	// Optionally clean up temporary files
	// For now, we'll keep the generated diagrams
	return nil
}