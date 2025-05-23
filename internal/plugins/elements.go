package plugins

import (
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

// Basic PDF elements that plugins can generate

type TextElement struct {
	Content  string
	FontSize float64
	Style    string // "", "B", "I", "U"
	X, Y     float64
}

func (t *TextElement) Render(pdf *gofpdf.Fpdf, ctx *RenderContext) error {
	if t.X != 0 || t.Y != 0 {
		pdf.SetXY(t.X, t.Y)
	}

	fontSize := t.FontSize
	if fontSize == 0 {
		fontSize = 12 // Default font size
	}

	style := t.Style
	if style == "" {
		style = ""
	}

	pdf.SetFont("Arial", style, fontSize)
	pdf.Cell(0, 6, t.Content)

	return nil
}

func (t *TextElement) Height() float64 {
	return 6
}

func (t *TextElement) Width() float64 {
	return 0 // Auto width
}

type ImageElement struct {
	Data        []byte
	Format      string // "PNG", "JPG", "GIF"
	ImageWidth  float64
	ImageHeight float64
	X, Y        float64
}

func (i *ImageElement) Render(pdf *gofpdf.Fpdf, ctx *RenderContext) error {
	if i.X != 0 || i.Y != 0 {
		pdf.SetXY(i.X, i.Y)
	}

	// Register and place the image
	if len(i.Data) > 0 {
		// Register image with PDF
		imageInfo := pdf.RegisterImageOptionsReader(
			fmt.Sprintf("img_%p", i), // Unique name for this image
			gofpdf.ImageOptions{ImageType: i.Format},
			bytes.NewReader(i.Data),
		)

		if imageInfo != nil {
			// Calculate dimensions to fit within page margins
			pageWidth, _ := pdf.GetPageSize()
			leftMargin, _, rightMargin, _ := pdf.GetMargins()
			maxWidth := pageWidth - leftMargin - rightMargin

			width := i.ImageWidth
			height := i.ImageHeight

			// Scale down if too wide
			if width > maxWidth {
				scale := maxWidth / width
				width = maxWidth
				height = height * scale
			}

			// Place the image
			pdf.ImageOptions(
				fmt.Sprintf("img_%p", i),
				-1, -1, // Use current position
				width, height,
				false,
				gofpdf.ImageOptions{ImageType: i.Format},
				0,
				"",
			)

			// Move cursor below image
			pdf.Ln(height + 3)
		}
	} else {
		// Fallback to text placeholder
		pdf.Cell(0, i.ImageHeight, "[Image placeholder]")
		pdf.Ln(i.ImageHeight)
	}

	return nil
}

func (i *ImageElement) Height() float64 {
	return i.ImageHeight
}

func (i *ImageElement) Width() float64 {
	return i.ImageWidth
}

type LineElement struct {
	X1, Y1    float64
	X2, Y2    float64
	LineWidth float64
	DrawMode  string // "D" for draw, "F" for fill
}

func (l *LineElement) Render(pdf *gofpdf.Fpdf, ctx *RenderContext) error {
	pdf.SetLineWidth(l.LineWidth)
	pdf.Line(l.X1, l.Y1, l.X2, l.Y2)
	return nil
}

func (l *LineElement) Height() float64 {
	if l.Y2 > l.Y1 {
		return l.Y2 - l.Y1
	}
	return l.Y1 - l.Y2
}

func (l *LineElement) Width() float64 {
	if l.X2 > l.X1 {
		return l.X2 - l.X1
	}
	return l.X1 - l.X2
}
