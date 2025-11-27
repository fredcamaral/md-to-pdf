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

// DocumentMetadata holds PDF document metadata
type DocumentMetadata struct {
	Title   string
	Author  string
	Subject string
}

type PDFRenderer struct {
	config   *RenderConfig
	document *DocumentMetadata
	plugins  *plugins.Manager
}

func NewPDFRenderer(config *RenderConfig, document *DocumentMetadata, pluginManager *plugins.Manager) *PDFRenderer {
	return &PDFRenderer{
		config:   config,
		document: document,
		plugins:  pluginManager,
	}
}

func (r *PDFRenderer) Render(node ast.Node, source []byte) (*bytes.Buffer, error) {
	pdf := gofpdf.New("P", "mm", r.config.PageSize, "")
	pdf.SetMargins(r.config.Margins.Left, r.config.Margins.Top, r.config.Margins.Right)
	pdf.SetAutoPageBreak(true, r.config.Margins.Bottom)
	pdf.AddPage()
	pdf.SetFont(r.config.FontFamily, "", r.config.FontSize)

	// Set document metadata if available
	if r.document != nil {
		pdf.SetTitle(r.document.Title, false)
		pdf.SetAuthor(r.document.Author, false)
		pdf.SetSubject(r.document.Subject, false)
	}

	// Generate BeforeContent elements (e.g., TOC, cover page)
	if r.plugins != nil {
		ctx := r.createRenderContext(pdf, source)
		elements, err := r.plugins.GenerateContent(plugins.BeforeContent, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate before content: %w", err)
		}
		for _, elem := range elements {
			if renderErr := elem.Render(pdf, ctx); renderErr != nil {
				return nil, fmt.Errorf("failed to render before content element: %w", renderErr)
			}
		}
	}

	err := r.walkAST(pdf, node, source)
	if err != nil {
		return nil, err
	}

	// Generate AfterContent elements (e.g., appendix, index)
	if r.plugins != nil {
		ctx := r.createRenderContext(pdf, source)
		elements, err := r.plugins.GenerateContent(plugins.AfterContent, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate after content: %w", err)
		}
		for _, elem := range elements {
			if renderErr := elem.Render(pdf, ctx); renderErr != nil {
				return nil, fmt.Errorf("failed to render after content element: %w", renderErr)
			}
		}
	}

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

// createRenderContext creates a render context for plugin content generation
func (r *PDFRenderer) createRenderContext(pdf *gofpdf.Fpdf, source []byte) *plugins.RenderContext {
	pageWidth, pageHeight := pdf.GetPageSize()
	return &plugins.RenderContext{
		PDF:        pdf,
		Source:     source,
		PageWidth:  pageWidth,
		PageHeight: pageHeight,
		Margins: plugins.RenderMargins{
			Top:    r.config.Margins.Top,
			Bottom: r.config.Margins.Bottom,
			Left:   r.config.Margins.Left,
			Right:  r.config.Margins.Right,
		},
		Metadata: make(map[string]interface{}),
		Config:   make(map[string]interface{}),
	}
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
		case ast.KindDocument:
			// Document node is just a container, continue walking children
		case ast.KindHeading:
			r.renderHeading(pdf, n.(*ast.Heading), source)
		case ast.KindParagraph:
			r.renderParagraph(pdf, n.(*ast.Paragraph), source)
		case ast.KindText:
			// Text nodes are handled by their parent (paragraph, heading, etc.)
			// to ensure proper text aggregation and formatting
		case ast.KindCodeBlock:
			r.renderCodeBlock(pdf, n, source)
		case ast.KindFencedCodeBlock:
			r.renderCodeBlock(pdf, n, source)
		case ast.KindList:
			r.renderList(pdf, n.(*ast.List), source)
			return ast.WalkSkipChildren, nil
		case ast.KindBlockquote:
			r.renderBlockquote(pdf, n.(*ast.Blockquote), source)
			return ast.WalkSkipChildren, nil
		case ast.KindThematicBreak:
			r.renderThematicBreak(pdf)
		case ast.KindImage:
			r.renderImage(pdf, n.(*ast.Image), source)
		case ast.KindLink:
			// Links are handled inline within text rendering
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
	pdf.Ln(fontSize * 1.1)

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
	imageData, err := os.ReadFile(imagePath) // #nosec G304 - path is generated internally by plugins
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

// renderList renders ordered and unordered lists
func (r *PDFRenderer) renderList(pdf *gofpdf.Fpdf, list *ast.List, source []byte) {
	pdf.SetFont(r.config.FontFamily, "", r.config.FontSize)
	pdf.Ln(2)

	itemNum := 1
	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() == ast.KindListItem {
			// Create bullet or number prefix
			var prefix string
			if list.IsOrdered() {
				prefix = fmt.Sprintf("%d. ", itemNum)
				itemNum++
			} else {
				prefix = "  * "
			}

			// Extract text from list item
			itemText := r.extractTextFromNode(child, source)
			pdf.MultiCell(0, r.config.FontSize*1.2, prefix+itemText, "", "", false)
		}
	}
	pdf.Ln(2)
}

// renderBlockquote renders blockquote elements with indentation
func (r *PDFRenderer) renderBlockquote(pdf *gofpdf.Fpdf, blockquote *ast.Blockquote, source []byte) {
	pdf.SetFont(r.config.FontFamily, "I", r.config.FontSize)
	pdf.Ln(2)

	// Add left margin for blockquote
	leftMargin, _, _, _ := pdf.GetMargins()
	pdf.SetLeftMargin(leftMargin + 10)

	// Extract and render blockquote content
	blockText := r.extractTextFromNode(blockquote, source)
	if blockText != "" {
		pdf.MultiCell(0, r.config.FontSize*1.2, blockText, "", "", false)
	}

	// Restore margin
	pdf.SetLeftMargin(leftMargin)
	pdf.SetFont(r.config.FontFamily, "", r.config.FontSize)
	pdf.Ln(2)
}

// renderThematicBreak renders horizontal rule (---, ***, ___)
func (r *PDFRenderer) renderThematicBreak(pdf *gofpdf.Fpdf) {
	pdf.Ln(5)

	// Draw a horizontal line
	pageWidth, _ := pdf.GetPageSize()
	leftMargin, _, rightMargin, _ := pdf.GetMargins()
	lineWidth := pageWidth - leftMargin - rightMargin

	x, y := pdf.GetXY()
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(x, y, x+lineWidth, y)
	pdf.SetDrawColor(0, 0, 0)

	pdf.Ln(5)
}

// renderImage renders image elements
func (r *PDFRenderer) renderImage(pdf *gofpdf.Fpdf, image *ast.Image, source []byte) {
	destination := string(image.Destination)
	altText := string(image.Text(source))

	// Try to load and render the image
	imageData, err := os.ReadFile(destination) // #nosec G304 - path from markdown content
	if err != nil {
		// Fallback to alt text if image can't be loaded
		pdf.SetFont(r.config.FontFamily, "I", r.config.FontSize)
		pdf.MultiCell(0, r.config.FontSize*1.2, fmt.Sprintf("[Image: %s]", altText), "", "", false)
		pdf.SetFont(r.config.FontFamily, "", r.config.FontSize)
		return
	}

	pdf.Ln(3)

	// Register and render the image
	imageName := fmt.Sprintf("img_%p", &imageData)
	imageReader := bytes.NewReader(imageData)

	// Determine image type from extension
	imageType := "PNG"
	if len(destination) > 4 {
		ext := destination[len(destination)-4:]
		switch ext {
		case ".jpg", "jpeg":
			imageType = "JPG"
		case ".gif":
			imageType = "GIF"
		}
	}

	info := pdf.RegisterImageOptionsReader(imageName, gofpdf.ImageOptions{ImageType: imageType}, imageReader)
	if info == nil {
		pdf.SetFont(r.config.FontFamily, "I", r.config.FontSize)
		pdf.MultiCell(0, r.config.FontSize*1.2, fmt.Sprintf("[Image failed to load: %s]", altText), "", "", false)
		pdf.SetFont(r.config.FontFamily, "", r.config.FontSize)
		return
	}

	// Calculate dimensions
	pageWidth, _ := pdf.GetPageSize()
	leftMargin, _, rightMargin, _ := pdf.GetMargins()
	maxWidth := pageWidth - leftMargin - rightMargin

	imgWidth, imgHeight := info.Extent()
	imgWidthMM := float64(imgWidth) * 0.264583 // Convert pixels to mm
	imgHeightMM := float64(imgHeight) * 0.264583

	// Scale if too wide
	if imgWidthMM > maxWidth {
		scale := maxWidth / imgWidthMM
		imgWidthMM = maxWidth
		imgHeightMM = imgHeightMM * scale
	}

	x, y := pdf.GetXY()
	pdf.ImageOptions(imageName, x, y, imgWidthMM, imgHeightMM, false, gofpdf.ImageOptions{ImageType: imageType}, 0, "")
	pdf.SetXY(x, y+imgHeightMM+3)
}

// extractTextFromNode recursively extracts text content from an AST node
func (r *PDFRenderer) extractTextFromNode(node ast.Node, source []byte) string {
	var result string
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindText {
			textNode := n.(*ast.Text)
			result += string(textNode.Segment.Value(source))
		}
		return ast.WalkContinue, nil
	})
	return result
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
