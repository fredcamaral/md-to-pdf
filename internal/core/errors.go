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

// Validation functions
func ValidateConfig(config *Config) error {
	var errors []string

	// Validate font size
	if config.Renderer.FontSize <= 0 || config.Renderer.FontSize > 72 {
		errors = append(errors, "font-size must be between 1 and 72 points")
	}

	// Validate margins
	if config.Renderer.Margins.Top < 0 || config.Renderer.Margins.Top > 100 {
		errors = append(errors, "margin-top must be between 0 and 100mm")
	}
	if config.Renderer.Margins.Bottom < 0 || config.Renderer.Margins.Bottom > 100 {
		errors = append(errors, "margin-bottom must be between 0 and 100mm")
	}
	if config.Renderer.Margins.Left < 0 || config.Renderer.Margins.Left > 100 {
		errors = append(errors, "margin-left must be between 0 and 100mm")
	}
	if config.Renderer.Margins.Right < 0 || config.Renderer.Margins.Right > 100 {
		errors = append(errors, "margin-right must be between 0 and 100mm")
	}

	// Validate line spacing
	if config.Renderer.LineSpacing <= 0 || config.Renderer.LineSpacing > 5 {
		errors = append(errors, "line-spacing must be between 0.1 and 5.0")
	}

	// Validate heading scale
	if config.Renderer.HeadingScale <= 0 || config.Renderer.HeadingScale > 10 {
		errors = append(errors, "heading-scale must be between 0.1 and 10.0")
	}

	// Validate code size
	if config.Renderer.CodeSize != 0 && (config.Renderer.CodeSize < 6 || config.Renderer.CodeSize > 48) {
		errors = append(errors, "code-size must be between 6 and 48 points")
	}

	// Validate mermaid scale
	if config.Renderer.Mermaid.Scale <= 0 || config.Renderer.Mermaid.Scale > 10 {
		errors = append(errors, "mermaid-scale must be between 0.1 and 10.0")
	}

	// Validate page size
	validPageSizes := []string{"A4", "A3", "A5", "LETTER", "LEGAL", "TABLOID"}
	pageSize := strings.ToUpper(config.Renderer.PageSize)
	isValidPageSize := false
	for _, valid := range validPageSizes {
		if pageSize == valid {
			isValidPageSize = true
			break
		}
	}
	if !isValidPageSize {
		errors = append(errors, fmt.Sprintf("page-size must be one of: %s", strings.Join(validPageSizes, ", ")))
	}

	if len(errors) > 0 {
		return &ConfigurationError{
			Message: strings.Join(errors, "; "),
		}
	}

	return nil
}
