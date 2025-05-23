package cmd

import (
	"fmt"

	"github.com/fredcamaral/md-to-pdf/internal/config"
	"github.com/fredcamaral/md-to-pdf/internal/core"
	"github.com/spf13/cobra"
)

var (
	outputPath    string
	pluginDir     string
	verbose       bool
	
	// Typography & Fonts
	fontFamily    string
	fontSize      float64
	headingScale  float64
	lineSpacing   float64
	
	// Code styling
	codeFont      string
	codeSize      float64
	
	// Page layout
	pageSize      string
	marginTop     float64
	marginBottom  float64
	marginLeft    float64
	marginRight   float64
	
	// PDF metadata
	title         string
	author        string
	subject       string
	
	// Mermaid settings
	mermaidScale  float64
)

var convertCmd = &cobra.Command{
	Use:   "convert [input.md...]",
	Short: "Convert Markdown files to PDF",
	Long:  "Convert one or more Markdown files to PDF format",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load base configuration
		baseConfig := core.DefaultConfig()
		
		// Load user configuration
		userConfig, err := config.LoadUserConfig()
		if err != nil {
			return fmt.Errorf("failed to load user config: %w", err)
		}
		
		// Apply user configuration
		config.ApplyUserConfig(baseConfig, userConfig)
		
		// Apply CLI flag overrides
		baseConfig.Plugins.Directory = pluginDir
		
		// Typography & Fonts
		if fontFamily != "" {
			baseConfig.Renderer.FontFamily = fontFamily
		}
		if fontSize > 0 {
			baseConfig.Renderer.FontSize = fontSize
		}
		if headingScale > 0 {
			baseConfig.Renderer.HeadingScale = headingScale
		}
		if lineSpacing > 0 {
			baseConfig.Renderer.LineSpacing = lineSpacing
		}
		
		// Code styling
		if codeFont != "" {
			baseConfig.Renderer.CodeFont = codeFont
		}
		if codeSize > 0 {
			baseConfig.Renderer.CodeSize = codeSize
		}
		
		// Page layout
		if pageSize != "" {
			baseConfig.Renderer.PageSize = pageSize
		}
		if marginTop > 0 {
			baseConfig.Renderer.Margins.Top = marginTop
		}
		if marginBottom > 0 {
			baseConfig.Renderer.Margins.Bottom = marginBottom
		}
		if marginLeft > 0 {
			baseConfig.Renderer.Margins.Left = marginLeft
		}
		if marginRight > 0 {
			baseConfig.Renderer.Margins.Right = marginRight
		}
		
		// PDF metadata
		if title != "" {
			baseConfig.Document.Title = title
		}
		if author != "" {
			baseConfig.Document.Author = author
		}
		if subject != "" {
			baseConfig.Document.Subject = subject
		}
		
		// Mermaid settings
		if mermaidScale > 0 {
			baseConfig.Renderer.Mermaid.Scale = mermaidScale
		}
		
		engine, err := core.NewEngine(baseConfig)
		if err != nil {
			return fmt.Errorf("failed to create engine: %w", err)
		}
		
		opts := core.ConversionOptions{
			InputFiles: args,
			OutputPath: outputPath,
			PluginDir:  pluginDir,
			Verbose:    verbose,
		}
		
		err = engine.Convert(opts)
		if err != nil {
			return fmt.Errorf("conversion failed: %w", err)
		}
		
		if verbose {
			fmt.Println("Conversion completed successfully")
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
	
	// Basic options
	convertCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output PDF file path")
	convertCmd.Flags().StringVarP(&pluginDir, "plugins", "p", "./plugins", "Plugin directory path")
	convertCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	
	// Typography & Fonts
	convertCmd.Flags().StringVar(&fontFamily, "font-family", "", "Font family (Arial, Times, Helvetica, etc.)")
	convertCmd.Flags().Float64Var(&fontSize, "font-size", 0, "Base font size in points")
	convertCmd.Flags().Float64Var(&headingScale, "heading-scale", 0, "Heading size multiplier (e.g., 1.5 = 50% bigger)")
	convertCmd.Flags().Float64Var(&lineSpacing, "line-spacing", 0, "Line spacing multiplier (e.g., 1.2 = 20% spacing)")
	
	// Code styling
	convertCmd.Flags().StringVar(&codeFont, "code-font", "", "Font family for code blocks")
	convertCmd.Flags().Float64Var(&codeSize, "code-size", 0, "Font size for code blocks")
	
	// Page layout
	convertCmd.Flags().StringVar(&pageSize, "page-size", "", "Page size (A4, A3, Letter, Legal)")
	convertCmd.Flags().Float64Var(&marginTop, "margin-top", 0, "Top margin in mm")
	convertCmd.Flags().Float64Var(&marginBottom, "margin-bottom", 0, "Bottom margin in mm")
	convertCmd.Flags().Float64Var(&marginLeft, "margin-left", 0, "Left margin in mm")
	convertCmd.Flags().Float64Var(&marginRight, "margin-right", 0, "Right margin in mm")
	
	// PDF metadata
	convertCmd.Flags().StringVar(&title, "title", "", "PDF document title")
	convertCmd.Flags().StringVar(&author, "author", "", "PDF document author")
	convertCmd.Flags().StringVar(&subject, "subject", "", "PDF document subject")
	
	// Mermaid settings
	convertCmd.Flags().Float64Var(&mermaidScale, "mermaid-scale", 0, "Mermaid diagram scale factor (e.g., 1.0=original size, 2.2=default size, 3.0=even bigger)")
}