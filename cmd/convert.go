package cmd

import (
	"fmt"

	"github.com/fredcamaral/md-to-pdf/internal/config"
	"github.com/fredcamaral/md-to-pdf/internal/core"
	"github.com/spf13/cobra"
)

// convertCommand encapsulates all state for the convert command,
// eliminating global variables and improving testability.
type convertCommand struct {
	outputPath string
	pluginDir  string
	verbose    bool

	// Typography & Fonts
	fontFamily   string
	fontSize     float64
	headingScale float64
	lineSpacing  float64

	// Code styling
	codeFont string
	codeSize float64

	// Page layout
	pageSize     string
	marginTop    float64
	marginBottom float64
	marginLeft   float64
	marginRight  float64

	// PDF metadata
	title   string
	author  string
	subject string

	// Mermaid settings
	mermaidScale float64
}

// newConvertCommand creates and configures the convert command with all flags.
func newConvertCommand() *cobra.Command {
	c := &convertCommand{}

	cmd := &cobra.Command{
		Use:   "convert [input.md...]",
		Short: "Convert Markdown files to PDF",
		Long:  "Convert one or more Markdown files to PDF format",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.run,
	}

	// Basic options
	cmd.Flags().StringVarP(&c.outputPath, "output", "o", "", "Output PDF file path")
	cmd.Flags().StringVarP(&c.pluginDir, "plugins", "p", "./plugins", "Plugin directory path")
	cmd.Flags().BoolVarP(&c.verbose, "verbose", "v", false, "Enable verbose output")

	// Typography & Fonts
	cmd.Flags().StringVar(&c.fontFamily, "font-family", "", "Font family (Arial, Times, Helvetica, etc.)")
	cmd.Flags().Float64Var(&c.fontSize, "font-size", 0, "Base font size in points")
	cmd.Flags().Float64Var(&c.headingScale, "heading-scale", 0, "Heading size multiplier (e.g., 1.5 = 50% bigger)")
	cmd.Flags().Float64Var(&c.lineSpacing, "line-spacing", 0, "Line spacing multiplier (e.g., 1.2 = 20% spacing)")

	// Code styling
	cmd.Flags().StringVar(&c.codeFont, "code-font", "", "Font family for code blocks")
	cmd.Flags().Float64Var(&c.codeSize, "code-size", 0, "Font size for code blocks")

	// Page layout
	cmd.Flags().StringVar(&c.pageSize, "page-size", "", "Page size (A4, A3, Letter, Legal)")
	cmd.Flags().Float64Var(&c.marginTop, "margin-top", 0, "Top margin in mm")
	cmd.Flags().Float64Var(&c.marginBottom, "margin-bottom", 0, "Bottom margin in mm")
	cmd.Flags().Float64Var(&c.marginLeft, "margin-left", 0, "Left margin in mm")
	cmd.Flags().Float64Var(&c.marginRight, "margin-right", 0, "Right margin in mm")

	// PDF metadata
	cmd.Flags().StringVar(&c.title, "title", "", "PDF document title")
	cmd.Flags().StringVar(&c.author, "author", "", "PDF document author")
	cmd.Flags().StringVar(&c.subject, "subject", "", "PDF document subject")

	// Mermaid settings
	cmd.Flags().Float64Var(&c.mermaidScale, "mermaid-scale", 0, "Mermaid diagram scale factor (e.g., 1.0=original size, 2.2=default size, 3.0=even bigger)")

	return cmd
}

// run executes the convert command logic.
func (c *convertCommand) run(cmd *cobra.Command, args []string) error {
	// Validate: cannot use --output with multiple input files
	if len(args) > 1 && c.outputPath != "" {
		return fmt.Errorf("cannot use --output with multiple input files; omit --output to generate individual PDFs")
	}

	// Load base configuration
	baseConfig := core.DefaultConfig()

	// Load user configuration
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	// Apply user configuration
	config.ApplyUserConfig(baseConfig, userConfig)

	// Apply CLI flag overrides using Changed() to support zero values
	c.applyOverrides(cmd, baseConfig)

	engine, err := core.NewEngine(baseConfig)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}

	opts := core.ConversionOptions{
		InputFiles: args,
		OutputPath: c.outputPath,
		PluginDir:  c.pluginDir,
		Verbose:    c.verbose,
	}

	err = engine.Convert(opts)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	if c.verbose {
		fmt.Println("Conversion completed successfully")
	}

	return nil
}

// applyOverrides applies CLI flag overrides to the configuration.
// Uses cmd.Flags().Changed() to detect explicitly set flags,
// allowing zero values to be set intentionally (e.g., 0mm margins for full-bleed printing).
func (c *convertCommand) applyOverrides(cmd *cobra.Command, cfg *core.Config) {
	// Plugin directory is always applied
	cfg.Plugins.Directory = c.pluginDir

	// Typography & Fonts
	if cmd.Flags().Changed("font-family") {
		cfg.Renderer.FontFamily = c.fontFamily
	}
	if cmd.Flags().Changed("font-size") {
		cfg.Renderer.FontSize = c.fontSize
	}
	if cmd.Flags().Changed("heading-scale") {
		cfg.Renderer.HeadingScale = c.headingScale
	}
	if cmd.Flags().Changed("line-spacing") {
		cfg.Renderer.LineSpacing = c.lineSpacing
	}

	// Code styling
	if cmd.Flags().Changed("code-font") {
		cfg.Renderer.CodeFont = c.codeFont
	}
	if cmd.Flags().Changed("code-size") {
		cfg.Renderer.CodeSize = c.codeSize
	}

	// Page layout
	if cmd.Flags().Changed("page-size") {
		cfg.Renderer.PageSize = c.pageSize
	}
	if cmd.Flags().Changed("margin-top") {
		cfg.Renderer.Margins.Top = c.marginTop
	}
	if cmd.Flags().Changed("margin-bottom") {
		cfg.Renderer.Margins.Bottom = c.marginBottom
	}
	if cmd.Flags().Changed("margin-left") {
		cfg.Renderer.Margins.Left = c.marginLeft
	}
	if cmd.Flags().Changed("margin-right") {
		cfg.Renderer.Margins.Right = c.marginRight
	}

	// PDF metadata
	if cmd.Flags().Changed("title") {
		cfg.Document.Title = c.title
	}
	if cmd.Flags().Changed("author") {
		cfg.Document.Author = c.author
	}
	if cmd.Flags().Changed("subject") {
		cfg.Document.Subject = c.subject
	}

	// Mermaid settings
	if cmd.Flags().Changed("mermaid-scale") {
		cfg.Renderer.Mermaid.Scale = c.mermaidScale
	}
}

func init() {
	rootCmd.AddCommand(newConvertCommand())
}
