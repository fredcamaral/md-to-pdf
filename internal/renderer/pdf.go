package renderer

import (
	"bytes"
	"fmt"
	"os"

	"github.com/fredcamaral/md-to-pdf/internal/plugins"
	"github.com/jung-kurt/gofpdf"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

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
	Scale     float64 // Scaling factor for mermaid diagrams
	MaxWidth  float64 // Maximum width in mm (0 = use page width)
	MaxHeight float64 // Maximum height in mm
}

type Margins struct {
	Top    float64
	Bottom float64
	Left   float64
	Right  float64
}

type PDFRenderer struct {
	config  *RenderConfig
	plugins *plugins.Manager
}

func NewPDFRenderer(config *RenderConfig, pluginManager *plugins.Manager) *PDFRenderer {
	return &PDFRenderer{
		config:  config,
		plugins: pluginManager,
	}
}

func (r *PDFRenderer) Render(node ast.Node, source []byte) (*bytes.Buffer, error) {
	pdf := gofpdf.New("P", "mm", r.config.PageSize, "")
	pdf.SetMargins(r.config.Margins.Left, r.config.Margins.Top, r.config.Margins.Right)
	pdf.SetAutoPageBreak(true, r.config.Margins.Bottom)
	pdf.AddPage()
	pdf.SetFont(r.config.FontFamily, "", r.config.FontSize)

	// Skip ContentGenerator for now to avoid duplicate image rendering

	err := r.walkAST(pdf, node, source)
	if err != nil {
		return nil, err
	}

	// Note: ContentGenerator phases would be used for TOC or other document-level content

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func (r *PDFRenderer) walkAST(pdf *gofpdf.Fpdf, node ast.Node, source []byte) error {
	// Apply AST transformers before rendering
	if r.plugins != nil {
		transformedNode, err := r.applyTransformers(node, source)
		if err != nil {
			return err
		}
		node = transformedNode
	}
	
	return ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n.Kind() {
		case ast.KindHeading:
			r.renderHeading(pdf, n.(*ast.Heading), source)
		case ast.KindParagraph:
			r.renderParagraph(pdf, n.(*ast.Paragraph), source)
		case ast.KindText:
			r.renderText(pdf, n.(*ast.Text), source)
		case ast.KindCodeBlock:
			r.renderCodeBlock(pdf, n, source)
		case ast.KindFencedCodeBlock:
			r.renderCodeBlock(pdf, n, source)
		}

		return ast.WalkContinue, nil
	})
}

func (r *PDFRenderer) applyTransformers(node ast.Node, source []byte) (ast.Node, error) {
	result := node
	
	err := ast.Walk(result, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		ctx := &plugins.TransformContext{
			CurrentNode: n,
			Parent:      n.Parent(),
			Source:      source,
			Metadata:    make(map[string]interface{}),
			Config:      make(map[string]interface{}),
		}

		transformedNode, err := r.plugins.ApplyTransformers(n, ctx)
		if err != nil {
			return ast.WalkStop, err
		}

		// If the node was transformed, replace it
		if transformedNode != n {
			if n.Parent() != nil {
				n.Parent().ReplaceChild(n.Parent(), n, transformedNode)
			}
		}

		return ast.WalkContinue, nil
	})
	
	return result, err
}

func (r *PDFRenderer) renderHeading(pdf *gofpdf.Fpdf, heading *ast.Heading, source []byte) {
	// Add space before heading
	pdf.Ln(5)
	
	fontSize := r.config.FontSize + float64(6-heading.Level)*2
	pdf.SetFont(r.config.FontFamily, "B", fontSize)
	
	// Extract heading text
	var headingText string
	for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() == ast.KindText {
			text := child.(*ast.Text)
			headingText += string(text.Segment.Value(source))
		}
	}
	
	// Render heading with proper line break
	pdf.Cell(0, fontSize*1.1, headingText)
	pdf.Ln(fontSize*1.1)
	
	// Add space after heading
	pdf.Ln(2)
}

func (r *PDFRenderer) renderParagraph(pdf *gofpdf.Fpdf, paragraph *ast.Paragraph, source []byte) {
	// Check if this is a mermaid image paragraph
	if imagePath, exists := paragraph.Attribute([]byte("data-mermaid-image")); exists {
		if pathBytes, ok := imagePath.([]byte); ok {
			r.renderMermaidImage(pdf, string(pathBytes))
			return
		}
	}
	
	pdf.SetFont(r.config.FontFamily, "", r.config.FontSize)
	
	// Extract all text from paragraph
	var paragraphText string
	for child := paragraph.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() == ast.KindText {
			text := child.(*ast.Text)
			paragraphText += string(text.Segment.Value(source))
		}
	}
	
	// Use MultiCell for proper text wrapping
	if paragraphText != "" {
		pdf.MultiCell(0, r.config.FontSize*1.2, paragraphText, "", "", false)
		pdf.Ln(2) // Space after paragraph
	}
}

func (r *PDFRenderer) renderMermaidImage(pdf *gofpdf.Fpdf, imagePath string) {
	// Read the image file
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		// Fallback to text if image can't be read
		pdf.MultiCell(0, r.config.FontSize*1.2, fmt.Sprintf("[Mermaid diagram: %s (failed to load)]", imagePath), "", "", false)
		pdf.Ln(3)
		return
	}
	
	// Add space before image
	pdf.Ln(5)
	
	// Register the image with PDF
	imageName := fmt.Sprintf("mermaid_%p", &imageData)
	imageReader := bytes.NewReader(imageData)
	
	info := pdf.RegisterImageOptionsReader(imageName, gofpdf.ImageOptions{ImageType: "PNG"}, imageReader)
	if info == nil {
		// Fallback to text if image registration fails
		pdf.MultiCell(0, r.config.FontSize*1.2, fmt.Sprintf("[Mermaid diagram: %s (failed to register)]", imagePath), "", "", false)
		pdf.Ln(3)
		return
	}
	
	// Calculate scaling using configuration
	pageWidth, _ := pdf.GetPageSize()
	leftMargin, _, rightMargin, _ := pdf.GetMargins()
	availableWidth := pageWidth - leftMargin - rightMargin - 10 // Conservative padding
	
	// Use configured max width or available page width
	maxWidth := r.config.Mermaid.MaxWidth
	if maxWidth == 0 {
		maxWidth = availableWidth
	}
	
	imgWidth, imgHeight := info.Extent()
	
	// Convert from pixels to mm using configured scaling
	baseScale := 0.2 * r.config.Mermaid.Scale // Use configured scale factor
	imgWidthMM := float64(imgWidth) * baseScale
	imgHeightMM := float64(imgHeight) * baseScale
	
	// Scale down if too wide
	if imgWidthMM > maxWidth {
		scale := maxWidth / imgWidthMM
		imgWidthMM = maxWidth
		imgHeightMM = imgHeightMM * scale
	}
	
	// Limit maximum height using configuration
	if imgHeightMM > r.config.Mermaid.MaxHeight {
		scale := r.config.Mermaid.MaxHeight / imgHeightMM
		imgHeightMM = r.config.Mermaid.MaxHeight
		imgWidthMM = imgWidthMM * scale
	}
	
	// Get current position to ensure proper placement
	x, y := pdf.GetXY()
	
	// Place the image at current position
	pdf.ImageOptions(imageName, x, y, imgWidthMM, imgHeightMM, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	
	// Move cursor to below the image with proper spacing
	pdf.SetXY(x, y+imgHeightMM+5)
}

func (r *PDFRenderer) renderText(pdf *gofpdf.Fpdf, text *ast.Text, source []byte) {
	// Text rendering is now handled in renderParagraph and renderHeading
	// This function is kept for compatibility but doesn't render directly
}

func (r *PDFRenderer) renderCodeBlock(pdf *gofpdf.Fpdf, codeBlock ast.Node, source []byte) {
	// Add space before code block
	pdf.Ln(3)
	
	pdf.SetFont("Courier", "", r.config.FontSize-1)
	
	// Add a light background for code blocks
	pdf.SetFillColor(245, 245, 245)
	
	lineHeight := float64(r.config.FontSize)
	
	var lines *text.Segments
	
	switch block := codeBlock.(type) {
	case *ast.CodeBlock:
		lines = block.Lines()
	case *ast.FencedCodeBlock:
		lines = block.Lines()
	default:
		return
	}
	
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		content := string(line.Value(source))
		// Remove trailing newlines/whitespace for cleaner display
		if len(content) > 0 && content[len(content)-1] == '\n' {
			content = content[:len(content)-1]
		}
		pdf.CellFormat(0, lineHeight, content, "", 1, "", true, 0, "")
	}
	
	// Reset background
	pdf.SetFillColor(255, 255, 255)
	pdf.SetFont(r.config.FontFamily, "", r.config.FontSize)
	
	// Add space after code block
	pdf.Ln(3)
}