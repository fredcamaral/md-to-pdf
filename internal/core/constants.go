package core

import "strings"

// ValidPageSizes defines the canonical list of supported page sizes.
// This is the single source of truth for page size validation across the application.
var ValidPageSizes = []string{"A3", "A4", "A5", "Letter", "Legal", "Tabloid"}

// Validation range constants for configuration values.
const (
	// Font size range in points
	FontSizeMin = 1.0
	FontSizeMax = 72.0

	// Code font size range in points
	CodeSizeMin = 6.0
	CodeSizeMax = 48.0

	// Margin range in millimeters
	MarginMin = 0.0
	MarginMax = 100.0

	// Line spacing multiplier range
	LineSpacingMin = 0.1
	LineSpacingMax = 5.0

	// Heading scale multiplier range
	HeadingScaleMin = 0.1
	HeadingScaleMax = 10.0

	// Mermaid scale multiplier range
	MermaidScaleMin = 0.1
	MermaidScaleMax = 10.0

	// Mermaid dimension range in mm
	MermaidDimensionMin = 0.0
	MermaidDimensionMax = 1000.0
)

// IsValidPageSize checks if the given page size is valid (case-insensitive).
func IsValidPageSize(size string) bool {
	size = strings.ToUpper(size)
	for _, valid := range ValidPageSizes {
		if strings.ToUpper(valid) == size {
			return true
		}
	}
	return false
}

// ValidPageSizesString returns a comma-separated list of valid page sizes for error messages.
func ValidPageSizesString() string {
	return strings.Join(ValidPageSizes, ", ")
}
