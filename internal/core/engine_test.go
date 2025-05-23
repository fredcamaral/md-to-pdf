package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	// Test basic config values
	if config.Renderer.FontSize != 12 {
		t.Errorf("Expected font size 12, got %f", config.Renderer.FontSize)
	}
	
	if config.Renderer.PageSize != "A4" {
		t.Errorf("Expected page size A4, got %s", config.Renderer.PageSize)
	}
	
	if config.Renderer.Mermaid.Scale != 2.2 {
		t.Errorf("Expected mermaid scale 2.2, got %f", config.Renderer.Mermaid.Scale)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name:      "Valid default config",
			config:    DefaultConfig(),
			expectErr: false,
		},
		{
			name: "Invalid font size - too small",
			config: &Config{
				Renderer: RenderConfig{
					FontSize:     0,
					PageSize:     "A4",
					LineSpacing:  1.2,
					HeadingScale: 1.5,
					Margins:      Margins{Top: 20, Bottom: 20, Left: 15, Right: 15},
					Mermaid:      MermaidConfig{Scale: 2.2},
				},
			},
			expectErr: true,
		},
		{
			name: "Invalid font size - too large",
			config: &Config{
				Renderer: RenderConfig{
					FontSize:     100,
					PageSize:     "A4",
					LineSpacing:  1.2,
					HeadingScale: 1.5,
					Margins:      Margins{Top: 20, Bottom: 20, Left: 15, Right: 15},
					Mermaid:      MermaidConfig{Scale: 2.2},
				},
			},
			expectErr: true,
		},
		{
			name: "Invalid page size",
			config: &Config{
				Renderer: RenderConfig{
					FontSize:     12,
					PageSize:     "INVALID",
					LineSpacing:  1.2,
					HeadingScale: 1.5,
					Margins:      Margins{Top: 20, Bottom: 20, Left: 15, Right: 15},
					Mermaid:      MermaidConfig{Scale: 2.2},
				},
			},
			expectErr: true,
		},
		{
			name: "Invalid margins",
			config: &Config{
				Renderer: RenderConfig{
					FontSize:     12,
					PageSize:     "A4",
					LineSpacing:  1.2,
					HeadingScale: 1.5,
					Margins:      Margins{Top: -5, Bottom: 20, Left: 15, Right: 15},
					Mermaid:      MermaidConfig{Scale: 2.2},
				},
			},
			expectErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestNewEngine(t *testing.T) {
	// Test with valid config
	config := DefaultConfig()
	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if engine == nil {
		t.Fatal("Expected engine, got nil")
	}
	
	// Test with invalid config
	invalidConfig := DefaultConfig()
	invalidConfig.Renderer.FontSize = 0
	_, err = NewEngine(invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid config, got none")
	}
}

func TestConversionError(t *testing.T) {
	err := &ConversionError{
		File:    "test.md",
		Phase:   "parsing",
		Message: "syntax error",
	}
	
	expected := "conversion failed for test.md during parsing: syntax error"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestEngine_Convert(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")
	testContent := "# Test Document\n\nThis is a test."
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Create engine
	config := DefaultConfig()
	config.Plugins.Enabled = false // Disable plugins for test
	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test conversion
	outputFile := filepath.Join(tempDir, "test.pdf")
	opts := ConversionOptions{
		InputFiles: []string{testFile},
		OutputPath: outputFile,
		Verbose:    false,
	}
	
	err = engine.Convert(opts)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}
	
	// Check if output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output PDF file was not created")
	}
}

func TestEngine_Convert_InvalidFile(t *testing.T) {
	config := DefaultConfig()
	config.Plugins.Enabled = false
	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	opts := ConversionOptions{
		InputFiles: []string{"nonexistent.md"},
		OutputPath: "",
		Verbose:    false,
	}
	
	err = engine.Convert(opts)
	if err == nil {
		t.Error("Expected error for nonexistent file, got none")
	}
	
	// Check if it's the right type of error (unwrap to find ConversionError)
	var convErr *ConversionError
	if !errors.As(err, &convErr) {
		t.Errorf("Expected ConversionError, got %T", err)
	}
}

// Helper function to check error types
func isErrorType(err error, target interface{}) bool {
	switch target.(type) {
	case **ConversionError:
		_, ok := err.(*ConversionError)
		return ok
	case **ConfigurationError:
		_, ok := err.(*ConfigurationError)
		return ok
	default:
		return false
	}
}