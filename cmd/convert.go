package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fredcamaral/md-to-pdf/internal/config"
	"github.com/fredcamaral/md-to-pdf/internal/core"
	"github.com/fredcamaral/md-to-pdf/internal/output"
	"github.com/fredcamaral/md-to-pdf/internal/ui"
	"github.com/fredcamaral/md-to-pdf/internal/watcher"
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

	// New features
	watch    bool
	jsonMode bool
}

// newConvertCommand creates and configures the convert command with all flags.
func newConvertCommand() *cobra.Command {
	c := &convertCommand{}

	cmd := &cobra.Command{
		Use:   "convert [input.md...]",
		Short: "Convert Markdown files to PDF",
		Long: `Convert one or more Markdown files to PDF format.

Use "-" as input to read from stdin (requires --output flag).

Examples:
  md-to-pdf convert document.md
  md-to-pdf convert doc1.md doc2.md
  md-to-pdf convert document.md -o output.pdf
  md-to-pdf convert document.md --watch
  echo "# Hello" | md-to-pdf convert - -o hello.pdf
  md-to-pdf convert document.md --json`,
		Args: cobra.MinimumNArgs(1),
		RunE: c.run,
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

	// New features
	cmd.Flags().BoolVarP(&c.watch, "watch", "w", false, "Watch input files for changes and re-convert automatically")
	cmd.Flags().BoolVar(&c.jsonMode, "json", false, "Output results in JSON format")

	return cmd
}

// run executes the convert command logic.
func (c *convertCommand) run(cmd *cobra.Command, args []string) error {
	// Check for stdin input
	isStdin := len(args) == 1 && args[0] == "-"

	// Validate stdin requirements
	if isStdin {
		if c.outputPath == "" {
			return fmt.Errorf("--output flag is required when reading from stdin")
		}
		if c.watch {
			return fmt.Errorf("--watch flag cannot be used with stdin input")
		}
	}

	// Validate: cannot use --output with multiple input files
	if len(args) > 1 && c.outputPath != "" {
		return fmt.Errorf("cannot use --output with multiple input files; omit --output to generate individual PDFs")
	}

	// Validate: watch mode with multiple files generates individual PDFs
	if c.watch && c.outputPath != "" && len(args) > 1 {
		return fmt.Errorf("cannot use --output with --watch and multiple input files")
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

	// Handle stdin input
	if isStdin {
		return c.runStdin(engine)
	}

	// Handle watch mode
	if c.watch {
		return c.runWatch(engine, args)
	}

	// Normal conversion
	return c.runConvert(engine, args)
}

// runStdin handles conversion from stdin.
func (c *convertCommand) runStdin(engine *core.Engine) error {
	formatter := output.NewFormatter(c.jsonMode)
	startTime := time.Now()

	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		duration := time.Since(startTime)
		convErr := fmt.Errorf("failed to read from stdin: %w", err)
		formatter.RecordError("stdin", duration, convErr)
		if c.jsonMode {
			return formatter.Print()
		}
		return convErr
	}

	if len(content) == 0 {
		duration := time.Since(startTime)
		convErr := fmt.Errorf("stdin is empty")
		formatter.RecordError("stdin", duration, convErr)
		if c.jsonMode {
			return formatter.Print()
		}
		return convErr
	}

	err = engine.ConvertFromContent(content, c.outputPath)
	duration := time.Since(startTime)

	if err != nil {
		formatter.RecordError("stdin", duration, err)
		if c.jsonMode {
			return formatter.Print()
		}
		return fmt.Errorf("conversion failed: %w", err)
	}

	formatter.RecordSuccess("stdin", c.outputPath, duration)

	if c.jsonMode {
		return formatter.Print()
	}

	if c.verbose {
		fmt.Printf("Converted stdin to %s\n", c.outputPath)
	}

	return nil
}

// runWatch handles watch mode.
func (c *convertCommand) runWatch(engine *core.Engine, args []string) error {
	// Validate files exist before starting watch
	for _, inputFile := range args {
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", inputFile)
		}
	}

	// Create convert function for watcher
	convertFunc := func(inputFile string) error {
		opts := core.ConversionOptions{
			InputFiles: []string{inputFile},
			OutputPath: c.outputPath,
			PluginDir:  c.pluginDir,
			Verbose:    false, // Watcher handles its own output
		}
		return engine.Convert(opts)
	}

	// Create watcher
	w, err := watcher.New(convertFunc)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	// Add files to watch
	for _, inputFile := range args {
		if err := w.AddFile(inputFile); err != nil {
			return fmt.Errorf("failed to watch file %s: %w", inputFile, err)
		}
	}

	// Do initial conversion
	fmt.Println("Performing initial conversion...")
	for _, inputFile := range args {
		if err := convertFunc(inputFile); err != nil {
			fmt.Printf("Initial conversion failed for %s: %v\n", inputFile, err)
		} else {
			fmt.Printf("Converted: %s\n", inputFile)
		}
	}

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nStopping file watcher...")
		cancel()
	}()

	fmt.Printf("\nWatching %d file(s) for changes. Press Ctrl+C to stop.\n", len(args))

	return w.Watch(ctx)
}

// runConvert handles normal conversion.
func (c *convertCommand) runConvert(engine *core.Engine, args []string) error {
	formatter := output.NewFormatter(c.jsonMode)

	// Setup UI output and progress
	uiOutput := ui.NewOutput()

	// Disable colors and progress for JSON mode
	if c.jsonMode {
		uiOutput.SetColorsEnabled(false)
	}

	// Create batch progress tracker
	batchProgress := ui.NewBatchProgress(uiOutput, len(args))
	if c.jsonMode {
		batchProgress.SetEnabled(false)
	}

	for i, inputFile := range args {
		startTime := time.Now()

		// Determine output path before conversion
		outputPath := c.outputPath
		if outputPath == "" {
			outputPath = deriveOutputPath(inputFile)
		}

		// Start progress for this file
		batchProgress.StartFile(filepath.Base(inputFile))

		opts := core.ConversionOptions{
			InputFiles: []string{inputFile},
			OutputPath: c.outputPath,
			PluginDir:  c.pluginDir,
			Verbose:    false, // We handle verbose output ourselves for JSON support
		}

		err := engine.Convert(opts)
		duration := time.Since(startTime)

		if err != nil {
			batchProgress.Error(err)
			formatter.RecordError(inputFile, duration, err)
			if !c.jsonMode {
				return fmt.Errorf("conversion failed: %w", err)
			}
			continue
		}

		formatter.RecordSuccess(inputFile, outputPath, duration)

		// Show completion for non-TTY (TTY shows spinner instead)
		if !batchProgress.IsEnabled() && !c.jsonMode {
			uiOutput.Successf("Converted: %s -> %s", filepath.Base(inputFile), outputPath)
		}

		// For single file in TTY mode, show success
		if batchProgress.IsEnabled() && len(args) == 1 {
			batchProgress.CompleteWithMessage(fmt.Sprintf("Converted: %s -> %s", filepath.Base(inputFile), outputPath))
		}

		// For multi-file, just update progress (completion will be shown at end)
		if i == len(args)-1 && len(args) > 1 {
			batchProgress.Complete()
		}
	}

	if c.jsonMode {
		if err := formatter.Print(); err != nil {
			return err
		}
		if formatter.HasErrors() {
			return fmt.Errorf("one or more conversions failed")
		}
		return nil
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

// deriveOutputPath generates the output PDF path from an input markdown path.
func deriveOutputPath(inputPath string) string {
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	return baseName + ".pdf"
}

func init() {
	rootCmd.AddCommand(newConvertCommand())
}
