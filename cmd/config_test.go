package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fredcamaral/md-to-pdf/internal/config"
	"github.com/fredcamaral/md-to-pdf/internal/core"
)

func TestSetConfigValue_ValidValues(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		validate func(*config.UserConfig) bool
	}{
		// Typography & Fonts
		{
			name:  "font-family",
			key:   "font-family",
			value: "Helvetica",
			validate: func(c *config.UserConfig) bool {
				return c.FontFamily == "Helvetica"
			},
		},
		{
			name:  "font-size",
			key:   "font-size",
			value: "14",
			validate: func(c *config.UserConfig) bool {
				return c.FontSize == 14.0
			},
		},
		{
			name:  "heading-scale",
			key:   "heading-scale",
			value: "1.8",
			validate: func(c *config.UserConfig) bool {
				return c.HeadingScale == 1.8
			},
		},
		{
			name:  "line-spacing",
			key:   "line-spacing",
			value: "1.5",
			validate: func(c *config.UserConfig) bool {
				return c.LineSpacing == 1.5
			},
		},
		// Code styling
		{
			name:  "code-font",
			key:   "code-font",
			value: "Monaco",
			validate: func(c *config.UserConfig) bool {
				return c.CodeFont == "Monaco"
			},
		},
		{
			name:  "code-size",
			key:   "code-size",
			value: "11",
			validate: func(c *config.UserConfig) bool {
				return c.CodeSize == 11.0
			},
		},
		// Page layout
		{
			name:  "page-size-A4",
			key:   "page-size",
			value: "A4",
			validate: func(c *config.UserConfig) bool {
				return c.PageSize == "A4"
			},
		},
		{
			name:  "page-size-Letter",
			key:   "page-size",
			value: "Letter",
			validate: func(c *config.UserConfig) bool {
				return c.PageSize == "Letter"
			},
		},
		{
			name:  "margin-top",
			key:   "margin-top",
			value: "25",
			validate: func(c *config.UserConfig) bool {
				return c.MarginTop == 25.0
			},
		},
		{
			name:  "margin-bottom",
			key:   "margin-bottom",
			value: "25",
			validate: func(c *config.UserConfig) bool {
				return c.MarginBottom == 25.0
			},
		},
		{
			name:  "margin-left",
			key:   "margin-left",
			value: "20",
			validate: func(c *config.UserConfig) bool {
				return c.MarginLeft == 20.0
			},
		},
		{
			name:  "margin-right",
			key:   "margin-right",
			value: "20",
			validate: func(c *config.UserConfig) bool {
				return c.MarginRight == 20.0
			},
		},
		// PDF metadata
		{
			name:  "title",
			key:   "title",
			value: "My Document",
			validate: func(c *config.UserConfig) bool {
				return c.Title == "My Document"
			},
		},
		{
			name:  "author",
			key:   "author",
			value: "John Doe",
			validate: func(c *config.UserConfig) bool {
				return c.Author == "John Doe"
			},
		},
		{
			name:  "subject",
			key:   "subject",
			value: "Test Subject",
			validate: func(c *config.UserConfig) bool {
				return c.Subject == "Test Subject"
			},
		},
		// Mermaid settings
		{
			name:  "mermaid-scale",
			key:   "mermaid-scale",
			value: "2.5",
			validate: func(c *config.UserConfig) bool {
				return c.MermaidScale == 2.5
			},
		},
		{
			name:  "mermaid-max-width",
			key:   "mermaid-max-width",
			value: "180",
			validate: func(c *config.UserConfig) bool {
				return c.MermaidMaxWidth == 180.0
			},
		},
		{
			name:  "mermaid-max-height",
			key:   "mermaid-max-height",
			value: "200",
			validate: func(c *config.UserConfig) bool {
				return c.MermaidMaxHeight == 200.0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userConfig := &config.UserConfig{}
			err := setConfigValue(userConfig, tt.key, tt.value)
			if err != nil {
				t.Errorf("setConfigValue(%s, %s) returned error: %v", tt.key, tt.value, err)
			}
			if !tt.validate(userConfig) {
				t.Errorf("setConfigValue(%s, %s) did not set value correctly", tt.key, tt.value)
			}
		})
	}
}

func TestSetConfigValue_InvalidRange(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		wantError bool
	}{
		{
			name:      "invalid_font_size_non_numeric",
			key:       "font-size",
			value:     "abc",
			wantError: true,
		},
		{
			name:      "invalid_heading_scale_non_numeric",
			key:       "heading-scale",
			value:     "not-a-number",
			wantError: true,
		},
		{
			name:      "invalid_line_spacing_non_numeric",
			key:       "line-spacing",
			value:     "invalid",
			wantError: true,
		},
		{
			name:      "invalid_margin_non_numeric",
			key:       "margin-top",
			value:     "twenty",
			wantError: true,
		},
		{
			name:      "invalid_code_size_non_numeric",
			key:       "code-size",
			value:     "large",
			wantError: true,
		},
		{
			name:      "invalid_mermaid_scale_non_numeric",
			key:       "mermaid-scale",
			value:     "big",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userConfig := &config.UserConfig{}
			err := setConfigValue(userConfig, tt.key, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("setConfigValue(%s, %s) error = %v, wantError %v",
					tt.key, tt.value, err, tt.wantError)
			}
		})
	}
}

func TestSetConfigValue_RangeValidation(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		wantError bool
	}{
		// font-size: 1-72
		{"font-size_below_min", "font-size", "0.5", true},
		{"font-size_at_min", "font-size", "1", false},
		{"font-size_at_max", "font-size", "72", false},
		{"font-size_above_max", "font-size", "73", true},
		{"font-size_way_above_max", "font-size", "1000", true},

		// margin-*: 0-100
		{"margin-top_below_min", "margin-top", "-1", true},
		{"margin-top_at_min", "margin-top", "0", false},
		{"margin-top_at_max", "margin-top", "100", false},
		{"margin-top_above_max", "margin-top", "101", true},
		{"margin-bottom_valid", "margin-bottom", "50", false},
		{"margin-left_valid", "margin-left", "25", false},
		{"margin-right_valid", "margin-right", "25", false},

		// line-spacing: 0.1-5.0
		{"line-spacing_below_min", "line-spacing", "0.05", true},
		{"line-spacing_at_min", "line-spacing", "0.1", false},
		{"line-spacing_at_max", "line-spacing", "5.0", false},
		{"line-spacing_above_max", "line-spacing", "5.1", true},

		// heading-scale: 0.1-10.0
		{"heading-scale_below_min", "heading-scale", "0.05", true},
		{"heading-scale_at_min", "heading-scale", "0.1", false},
		{"heading-scale_at_max", "heading-scale", "10.0", false},
		{"heading-scale_above_max", "heading-scale", "10.1", true},

		// code-size: 6-48
		{"code-size_below_min", "code-size", "5", true},
		{"code-size_at_min", "code-size", "6", false},
		{"code-size_at_max", "code-size", "48", false},
		{"code-size_above_max", "code-size", "49", true},

		// mermaid-scale: 0.1-10.0
		{"mermaid-scale_below_min", "mermaid-scale", "0.05", true},
		{"mermaid-scale_at_min", "mermaid-scale", "0.1", false},
		{"mermaid-scale_at_max", "mermaid-scale", "10.0", false},
		{"mermaid-scale_above_max", "mermaid-scale", "10.1", true},

		// mermaid-max-width: 0-1000
		{"mermaid-max-width_below_min", "mermaid-max-width", "-1", true},
		{"mermaid-max-width_at_min", "mermaid-max-width", "0", false},
		{"mermaid-max-width_at_max", "mermaid-max-width", "1000", false},
		{"mermaid-max-width_above_max", "mermaid-max-width", "1001", true},

		// mermaid-max-height: 0-1000
		{"mermaid-max-height_below_min", "mermaid-max-height", "-1", true},
		{"mermaid-max-height_at_min", "mermaid-max-height", "0", false},
		{"mermaid-max-height_at_max", "mermaid-max-height", "1000", false},
		{"mermaid-max-height_above_max", "mermaid-max-height", "1001", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userConfig := &config.UserConfig{}
			err := setConfigValue(userConfig, tt.key, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("setConfigValue(%s, %s) error = %v, wantError %v",
					tt.key, tt.value, err, tt.wantError)
			}
		})
	}
}

func TestSetConfigValue_InvalidType(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		wantError bool
	}{
		{
			name:      "float_value_for_float_field",
			key:       "font-size",
			value:     "12.5",
			wantError: false,
		},
		{
			name:      "string_for_numeric_field",
			key:       "font-size",
			value:     "twelve",
			wantError: true,
		},
		{
			name:      "empty_string_for_numeric_field",
			key:       "font-size",
			value:     "",
			wantError: true,
		},
		{
			name:      "negative_for_margin",
			key:       "margin-top",
			value:     "-10",
			wantError: true, // negative margins are now rejected by range validation
		},
		{
			name:      "scientific_notation_out_of_range",
			key:       "font-size",
			value:     "1e2",
			wantError: true, // 100 is out of range for font-size (1-72)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userConfig := &config.UserConfig{}
			err := setConfigValue(userConfig, tt.key, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("setConfigValue(%s, %s) error = %v, wantError %v",
					tt.key, tt.value, err, tt.wantError)
			}
		})
	}
}

func TestResetConfigValue(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		setup    func(*config.UserConfig)
		validate func(*config.UserConfig) bool
	}{
		{
			name: "reset_font_family",
			key:  "font-family",
			setup: func(c *config.UserConfig) {
				c.FontFamily = "Helvetica"
			},
			validate: func(c *config.UserConfig) bool {
				return c.FontFamily == ""
			},
		},
		{
			name: "reset_font_size",
			key:  "font-size",
			setup: func(c *config.UserConfig) {
				c.FontSize = 14.0
			},
			validate: func(c *config.UserConfig) bool {
				return c.FontSize == 0
			},
		},
		{
			name: "reset_heading_scale",
			key:  "heading-scale",
			setup: func(c *config.UserConfig) {
				c.HeadingScale = 2.0
			},
			validate: func(c *config.UserConfig) bool {
				return c.HeadingScale == 0
			},
		},
		{
			name: "reset_line_spacing",
			key:  "line-spacing",
			setup: func(c *config.UserConfig) {
				c.LineSpacing = 1.5
			},
			validate: func(c *config.UserConfig) bool {
				return c.LineSpacing == 0
			},
		},
		{
			name: "reset_code_font",
			key:  "code-font",
			setup: func(c *config.UserConfig) {
				c.CodeFont = "Monaco"
			},
			validate: func(c *config.UserConfig) bool {
				return c.CodeFont == ""
			},
		},
		{
			name: "reset_code_size",
			key:  "code-size",
			setup: func(c *config.UserConfig) {
				c.CodeSize = 11.0
			},
			validate: func(c *config.UserConfig) bool {
				return c.CodeSize == 0
			},
		},
		{
			name: "reset_page_size",
			key:  "page-size",
			setup: func(c *config.UserConfig) {
				c.PageSize = "Letter"
			},
			validate: func(c *config.UserConfig) bool {
				return c.PageSize == ""
			},
		},
		{
			name: "reset_margin_top",
			key:  "margin-top",
			setup: func(c *config.UserConfig) {
				c.MarginTop = 25.0
			},
			validate: func(c *config.UserConfig) bool {
				return c.MarginTop == 0
			},
		},
		{
			name: "reset_title",
			key:  "title",
			setup: func(c *config.UserConfig) {
				c.Title = "My Document"
			},
			validate: func(c *config.UserConfig) bool {
				return c.Title == ""
			},
		},
		{
			name: "reset_mermaid_scale",
			key:  "mermaid-scale",
			setup: func(c *config.UserConfig) {
				c.MermaidScale = 3.0
			},
			validate: func(c *config.UserConfig) bool {
				return c.MermaidScale == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userConfig := &config.UserConfig{}
			tt.setup(userConfig)

			err := resetConfigValue(userConfig, tt.key)
			if err != nil {
				t.Errorf("resetConfigValue(%s) returned error: %v", tt.key, err)
			}
			if !tt.validate(userConfig) {
				t.Errorf("resetConfigValue(%s) did not reset value correctly", tt.key)
			}
		})
	}
}

func TestResetConfigValue_UnknownKey(t *testing.T) {
	userConfig := &config.UserConfig{}
	err := resetConfigValue(userConfig, "unknown-key")
	if err == nil {
		t.Error("resetConfigValue with unknown key should return error")
	}
}

func TestSetConfigValue_UnknownKey(t *testing.T) {
	userConfig := &config.UserConfig{}
	err := setConfigValue(userConfig, "unknown-key", "value")
	if err == nil {
		t.Error("setConfigValue with unknown key should return error")
	}
}

func TestSetConfigValue_UnknownKeyListsValidKeys(t *testing.T) {
	userConfig := &config.UserConfig{}
	err := setConfigValue(userConfig, "fontsize", "12")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "unknown configuration key: fontsize") {
		t.Errorf("error should mention the unknown key, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Valid keys:") {
		t.Errorf("error should list valid keys, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "font-size") {
		t.Errorf("error should include 'font-size' as a valid key, got: %s", errMsg)
	}
}

func TestResetConfigValue_UnknownKeyListsValidKeys(t *testing.T) {
	userConfig := &config.UserConfig{}
	err := resetConfigValue(userConfig, "fontsize")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Valid keys:") {
		t.Errorf("error should list valid keys, got: %s", errMsg)
	}
}

func TestFindConfigKey(t *testing.T) {
	tests := []struct {
		name     string
		keyName  string
		wantNil  bool
	}{
		{"valid_key_font_size", "font-size", false},
		{"valid_key_page_size", "page-size", false},
		{"valid_key_mermaid_scale", "mermaid-scale", false},
		{"invalid_key", "not-a-key", true},
		{"typo", "fontsize", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findConfigKey(tt.keyName)
			gotNil := result == nil
			if gotNil != tt.wantNil {
				t.Errorf("findConfigKey(%q) nil = %v, want nil = %v", tt.keyName, gotNil, tt.wantNil)
			}
		})
	}
}

func TestValidKeysString(t *testing.T) {
	result := validKeysString()

	expectedKeys := []string{
		"font-family", "font-size", "heading-scale", "line-spacing",
		"code-font", "code-size", "page-size",
		"margin-top", "margin-bottom", "margin-left", "margin-right",
		"title", "author", "subject",
		"mermaid-scale", "mermaid-max-width", "mermaid-max-height",
	}

	for _, key := range expectedKeys {
		if !strings.Contains(result, key) {
			t.Errorf("validKeysString() should contain %q, got: %s", key, result)
		}
	}
}

func TestIsValidPageSize(t *testing.T) {
	tests := []struct {
		name     string
		size     string
		expected bool
	}{
		// Valid sizes
		{"A4_uppercase", "A4", true},
		{"A4_lowercase", "a4", true},
		{"A3_uppercase", "A3", true},
		{"A5_uppercase", "A5", true},
		{"Letter_titlecase", "Letter", true},
		{"Letter_uppercase", "LETTER", true},
		{"Letter_lowercase", "letter", true},
		{"Legal_titlecase", "Legal", true},
		{"Tabloid_titlecase", "Tabloid", true},

		// Invalid sizes
		{"A6", "A6", false},
		{"B4", "B4", false},
		{"empty_string", "", false},
		{"invalid_string", "InvalidSize", false},
		{"numbers_only", "123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := core.IsValidPageSize(tt.size)
			if result != tt.expected {
				t.Errorf("core.IsValidPageSize(%q) = %v, want %v", tt.size, result, tt.expected)
			}
		})
	}
}

func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"empty_string", "", true},
		{"non_empty_string", "hello", false},
		{"zero_float", float64(0), true},
		{"non_zero_float", float64(10.5), false},
		{"negative_float", float64(-5.0), false},
		{"zero_int", int(0), true},
		{"non_zero_int", int(10), false},
		{"negative_int", int(-5), false},
		{"nil_interface", nil, false}, // nil returns false (default case)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isZeroValue(tt.value)
			if result != tt.expected {
				t.Errorf("isZeroValue(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestRemoveConfigFile(t *testing.T) {
	t.Run("remove_existing_file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "config.yaml")

		// Create a test file
		if err := os.WriteFile(testFile, []byte("test: value"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Remove the file
		err := removeConfigFile(testFile)
		if err != nil {
			t.Errorf("removeConfigFile returned error: %v", err)
		}

		// Verify file is removed
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Error("file should have been removed")
		}
	})

	t.Run("remove_non_existing_file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "nonexistent.yaml")

		// Should not return error for non-existing file
		err := removeConfigFile(testFile)
		if err != nil {
			t.Errorf("removeConfigFile should not error for non-existing file: %v", err)
		}
	})
}

func TestPrintConfigValue(t *testing.T) {
	// This test mainly ensures the function doesn't panic
	// Actual output verification would require capturing stdout

	tests := []struct {
		name         string
		key          string
		userValue    interface{}
		defaultValue interface{}
	}{
		{"string_default", "font-family", "", "Arial"},
		{"string_user", "font-family", "Helvetica", "Arial"},
		{"float_default", "font-size", float64(0), 12.0},
		{"float_user", "font-size", 14.0, 12.0},
		{"int_default", "some-int", int(0), 10},
		{"int_user", "some-int", int(5), 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			printConfigValue(tt.key, tt.userValue, tt.defaultValue)
		})
	}
}

func TestSetConfigValue_InvalidPageSize(t *testing.T) {
	userConfig := &config.UserConfig{}
	err := setConfigValue(userConfig, "page-size", "InvalidSize")
	if err == nil {
		t.Error("setConfigValue with invalid page-size should return error")
	}
}
