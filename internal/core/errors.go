package core

import (
	"fmt"
	"strings"
)

// Error types for better error handling
type ConversionError struct {
	File    string
	Phase   string
	Message string
	Cause   error
}

func (e *ConversionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("conversion failed for %s during %s: %s (%v)", e.File, e.Phase, e.Message, e.Cause)
	}
	return fmt.Sprintf("conversion failed for %s during %s: %s", e.File, e.Phase, e.Message)
}

func (e *ConversionError) Unwrap() error {
	return e.Cause
}

type PluginError struct {
	Plugin    string
	Operation string
	Message   string
	Cause     error
}

func (e *PluginError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("plugin %s failed during %s: %s (%v)", e.Plugin, e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("plugin %s failed during %s: %s", e.Plugin, e.Operation, e.Message)
}

func (e *PluginError) Unwrap() error {
	return e.Cause
}

type ConfigurationError struct {
	Key     string
	Value   string
	Message string
	Cause   error
}

func (e *ConfigurationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("configuration error for %s=%s: %s (%v)", e.Key, e.Value, e.Message, e.Cause)
	}
	return fmt.Sprintf("configuration error for %s=%s: %s", e.Key, e.Value, e.Message)
}

func (e *ConfigurationError) Unwrap() error {
	return e.Cause
}

// ValidateConfig validates a configuration against defined constraints.
// Uses constants from constants.go as the single source of truth for validation ranges.
func ValidateConfig(config *Config) error {
	var errors []string

	// Validate font size
	if config.Renderer.FontSize < FontSizeMin || config.Renderer.FontSize > FontSizeMax {
		errors = append(errors, fmt.Sprintf("font-size must be between %.0f and %.0f points", FontSizeMin, FontSizeMax))
	}

	// Validate margins
	if config.Renderer.Margins.Top < MarginMin || config.Renderer.Margins.Top > MarginMax {
		errors = append(errors, fmt.Sprintf("margin-top must be between %.0f and %.0fmm", MarginMin, MarginMax))
	}
	if config.Renderer.Margins.Bottom < MarginMin || config.Renderer.Margins.Bottom > MarginMax {
		errors = append(errors, fmt.Sprintf("margin-bottom must be between %.0f and %.0fmm", MarginMin, MarginMax))
	}
	if config.Renderer.Margins.Left < MarginMin || config.Renderer.Margins.Left > MarginMax {
		errors = append(errors, fmt.Sprintf("margin-left must be between %.0f and %.0fmm", MarginMin, MarginMax))
	}
	if config.Renderer.Margins.Right < MarginMin || config.Renderer.Margins.Right > MarginMax {
		errors = append(errors, fmt.Sprintf("margin-right must be between %.0f and %.0fmm", MarginMin, MarginMax))
	}

	// Validate line spacing
	if config.Renderer.LineSpacing < LineSpacingMin || config.Renderer.LineSpacing > LineSpacingMax {
		errors = append(errors, fmt.Sprintf("line-spacing must be between %.1f and %.1f", LineSpacingMin, LineSpacingMax))
	}

	// Validate heading scale
	if config.Renderer.HeadingScale < HeadingScaleMin || config.Renderer.HeadingScale > HeadingScaleMax {
		errors = append(errors, fmt.Sprintf("heading-scale must be between %.1f and %.1f", HeadingScaleMin, HeadingScaleMax))
	}

	// Validate code size (0 means use default, so only validate non-zero values)
	if config.Renderer.CodeSize != 0 && (config.Renderer.CodeSize < CodeSizeMin || config.Renderer.CodeSize > CodeSizeMax) {
		errors = append(errors, fmt.Sprintf("code-size must be between %.0f and %.0f points", CodeSizeMin, CodeSizeMax))
	}

	// Validate mermaid scale
	if config.Renderer.Mermaid.Scale < MermaidScaleMin || config.Renderer.Mermaid.Scale > MermaidScaleMax {
		errors = append(errors, fmt.Sprintf("mermaid-scale must be between %.1f and %.1f", MermaidScaleMin, MermaidScaleMax))
	}

	// Validate page size using shared function
	if !IsValidPageSize(config.Renderer.PageSize) {
		errors = append(errors, fmt.Sprintf("page-size must be one of: %s", ValidPageSizesString()))
	}

	if len(errors) > 0 {
		return &ConfigurationError{
			Message: strings.Join(errors, "; "),
		}
	}

	return nil
}
