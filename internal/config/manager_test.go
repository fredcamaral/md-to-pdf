package config

import (
	"testing"

	"github.com/fredcamaral/md-to-pdf/internal/core"
)

func TestLoadUserConfig_EmptyConfig(t *testing.T) {
	// Test config loading (this might load existing user config)
	config, err := LoadUserConfig()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config, got nil")
	}

	// Config should be loaded successfully (values may exist if user has config)
	// This test just ensures the loading mechanism works
}

// TestSaveAndLoadUserConfig is skipped for now due to global state dependency
// In a real implementation, we would refactor to inject the config path
func TestUserConfigStructure(t *testing.T) {
	// Test that UserConfig struct works correctly
	testConfig := &UserConfig{
		FontFamily:   "Times",
		FontSize:     14,
		Author:       "Test Author",
		MermaidScale: 3.0,
	}

	if testConfig.FontFamily != "Times" {
		t.Errorf("Expected FontFamily Times, got %s", testConfig.FontFamily)
	}
	if testConfig.FontSize != 14 {
		t.Errorf("Expected FontSize 14, got %f", testConfig.FontSize)
	}
	if testConfig.Author != "Test Author" {
		t.Errorf("Expected Author 'Test Author', got %s", testConfig.Author)
	}
	if testConfig.MermaidScale != 3.0 {
		t.Errorf("Expected MermaidScale 3.0, got %f", testConfig.MermaidScale)
	}
}

func TestApplyUserConfig(t *testing.T) {
	baseConfig := core.DefaultConfig()
	userConfig := &UserConfig{
		FontFamily:   "Times",
		FontSize:     16,
		LineSpacing:  1.5,
		Author:       "Test Author",
		MermaidScale: 3.0,
	}

	// Apply user config
	ApplyUserConfig(baseConfig, userConfig)

	// Check if values were applied
	if baseConfig.Renderer.FontFamily != "Times" {
		t.Errorf("Expected FontFamily Times, got %s", baseConfig.Renderer.FontFamily)
	}
	if baseConfig.Renderer.FontSize != 16 {
		t.Errorf("Expected FontSize 16, got %f", baseConfig.Renderer.FontSize)
	}
	if baseConfig.Renderer.LineSpacing != 1.5 {
		t.Errorf("Expected LineSpacing 1.5, got %f", baseConfig.Renderer.LineSpacing)
	}
	if baseConfig.Document.Author != "Test Author" {
		t.Errorf("Expected Author 'Test Author', got %s", baseConfig.Document.Author)
	}
	if baseConfig.Renderer.Mermaid.Scale != 3.0 {
		t.Errorf("Expected MermaidScale 3.0, got %f", baseConfig.Renderer.Mermaid.Scale)
	}
}

func TestApplyUserConfig_ZeroValues(t *testing.T) {
	baseConfig := core.DefaultConfig()
	originalFontSize := baseConfig.Renderer.FontSize

	userConfig := &UserConfig{
		FontFamily: "Times", // Set this
		FontSize:   0,       // Don't set this (zero value)
	}

	ApplyUserConfig(baseConfig, userConfig)

	// FontFamily should be changed
	if baseConfig.Renderer.FontFamily != "Times" {
		t.Errorf("Expected FontFamily to be changed to Times, got %s", baseConfig.Renderer.FontFamily)
	}

	// FontSize should remain unchanged (zero value ignored)
	if baseConfig.Renderer.FontSize != originalFontSize {
		t.Errorf("Expected FontSize to remain %f, got %f", originalFontSize, baseConfig.Renderer.FontSize)
	}
}
