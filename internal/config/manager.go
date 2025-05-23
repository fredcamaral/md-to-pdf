package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fredcamaral/md-to-pdf/internal/core"
	"gopkg.in/yaml.v3"
)

const (
	ConfigDir  = ".config/md-to-pdf"
	ConfigFile = "config.yaml"
)

type UserConfig struct {
	// Typography & Fonts
	FontFamily   string  `yaml:"font_family,omitempty"`
	FontSize     float64 `yaml:"font_size,omitempty"`
	HeadingScale float64 `yaml:"heading_scale,omitempty"`
	LineSpacing  float64 `yaml:"line_spacing,omitempty"`

	// Code styling
	CodeFont string  `yaml:"code_font,omitempty"`
	CodeSize float64 `yaml:"code_size,omitempty"`

	// Page layout
	PageSize     string  `yaml:"page_size,omitempty"`
	MarginTop    float64 `yaml:"margin_top,omitempty"`
	MarginBottom float64 `yaml:"margin_bottom,omitempty"`
	MarginLeft   float64 `yaml:"margin_left,omitempty"`
	MarginRight  float64 `yaml:"margin_right,omitempty"`

	// PDF metadata
	Title   string `yaml:"title,omitempty"`
	Author  string `yaml:"author,omitempty"`
	Subject string `yaml:"subject,omitempty"`

	// Mermaid settings
	MermaidScale     float64 `yaml:"mermaid_scale,omitempty"`
	MermaidMaxWidth  float64 `yaml:"mermaid_max_width,omitempty"`
	MermaidMaxHeight float64 `yaml:"mermaid_max_height,omitempty"`
}

func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ConfigDir, ConfigFile)
}

func LoadUserConfig() (*UserConfig, error) {
	configPath := GetConfigPath()

	// Return empty config if file doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &UserConfig{}, nil
	}

	data, err := os.ReadFile(configPath) // #nosec G304 - config path is generated from user's home directory
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config UserConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func SaveUserConfig(config *UserConfig) error {
	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)

	// Create config directory if it doesn't exist
	err := os.MkdirAll(configDir, 0750)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func ApplyUserConfig(baseConfig *core.Config, userConfig *UserConfig) {
	// Typography & Fonts
	if userConfig.FontFamily != "" {
		baseConfig.Renderer.FontFamily = userConfig.FontFamily
	}
	if userConfig.FontSize > 0 {
		baseConfig.Renderer.FontSize = userConfig.FontSize
	}
	if userConfig.HeadingScale > 0 {
		baseConfig.Renderer.HeadingScale = userConfig.HeadingScale
	}
	if userConfig.LineSpacing > 0 {
		baseConfig.Renderer.LineSpacing = userConfig.LineSpacing
	}

	// Code styling
	if userConfig.CodeFont != "" {
		baseConfig.Renderer.CodeFont = userConfig.CodeFont
	}
	if userConfig.CodeSize > 0 {
		baseConfig.Renderer.CodeSize = userConfig.CodeSize
	}

	// Page layout
	if userConfig.PageSize != "" {
		baseConfig.Renderer.PageSize = userConfig.PageSize
	}
	if userConfig.MarginTop > 0 {
		baseConfig.Renderer.Margins.Top = userConfig.MarginTop
	}
	if userConfig.MarginBottom > 0 {
		baseConfig.Renderer.Margins.Bottom = userConfig.MarginBottom
	}
	if userConfig.MarginLeft > 0 {
		baseConfig.Renderer.Margins.Left = userConfig.MarginLeft
	}
	if userConfig.MarginRight > 0 {
		baseConfig.Renderer.Margins.Right = userConfig.MarginRight
	}

	// PDF metadata
	if userConfig.Title != "" {
		baseConfig.Document.Title = userConfig.Title
	}
	if userConfig.Author != "" {
		baseConfig.Document.Author = userConfig.Author
	}
	if userConfig.Subject != "" {
		baseConfig.Document.Subject = userConfig.Subject
	}

	// Mermaid settings
	if userConfig.MermaidScale > 0 {
		baseConfig.Renderer.Mermaid.Scale = userConfig.MermaidScale
	}
	if userConfig.MermaidMaxWidth > 0 {
		baseConfig.Renderer.Mermaid.MaxWidth = userConfig.MermaidMaxWidth
	}
	if userConfig.MermaidMaxHeight > 0 {
		baseConfig.Renderer.Mermaid.MaxHeight = userConfig.MermaidMaxHeight
	}
}
