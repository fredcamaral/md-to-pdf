package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fredcamaral/md-to-pdf/internal/parser"
	"github.com/fredcamaral/md-to-pdf/internal/plugins"
	"github.com/fredcamaral/md-to-pdf/internal/renderer"
)

type Engine struct {
	parser   *parser.MarkdownParser
	renderer *renderer.PDFRenderer
	plugins  *plugins.Manager
	config   *Config
}

func NewEngine(config *Config) (*Engine, error) {
	// Validate configuration
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}
	rendererConfig := &renderer.RenderConfig{
		PageSize:     config.Renderer.PageSize,
		FontFamily:   config.Renderer.FontFamily,
		FontSize:     config.Renderer.FontSize,
		HeadingScale: config.Renderer.HeadingScale,
		LineSpacing:  config.Renderer.LineSpacing,
		CodeFont:     config.Renderer.CodeFont,
		CodeSize:     config.Renderer.CodeSize,
		Margins: renderer.Margins{
			Top:    config.Renderer.Margins.Top,
			Bottom: config.Renderer.Margins.Bottom,
			Left:   config.Renderer.Margins.Left,
			Right:  config.Renderer.Margins.Right,
		},
		Mermaid: renderer.MermaidConfig{
			Scale:     config.Renderer.Mermaid.Scale,
			MaxWidth:  config.Renderer.Mermaid.MaxWidth,
			MaxHeight: config.Renderer.Mermaid.MaxHeight,
		},
	}

	pluginManager := plugins.NewManager(config.Plugins.Directory, config.Plugins.Enabled)

	return &Engine{
		parser:   parser.NewMarkdownParser(),
		renderer: renderer.NewPDFRenderer(rendererConfig, pluginManager),
		plugins:  pluginManager,
		config:   config,
	}, nil
}

func (e *Engine) Convert(opts ConversionOptions) error {
	// Load plugins
	err := e.plugins.LoadPlugins()
	if err != nil {
		return fmt.Errorf("failed to load plugins: %w", err)
	}

	defer func() {
		if cleanupErr := e.plugins.Cleanup(); cleanupErr != nil {
			fmt.Printf("Warning: plugin cleanup failed: %v\n", cleanupErr)
		}
	}()

	for _, inputFile := range opts.InputFiles {
		err := e.convertFile(inputFile, opts.OutputPath)
		if err != nil {
			return fmt.Errorf("failed to convert %s: %w", inputFile, err)
		}

		if opts.Verbose {
			fmt.Printf("Converted: %s\n", inputFile)
		}
	}

	return nil
}

func (e *Engine) convertFile(inputPath, outputPath string) error {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return &ConversionError{
			File:    inputPath,
			Phase:   "file reading",
			Message: "could not read input file",
			Cause:   err,
		}
	}

	node, err := e.parser.Parse(content)
	if err != nil {
		return &ConversionError{
			File:    inputPath,
			Phase:   "markdown parsing",
			Message: "could not parse markdown content",
			Cause:   err,
		}
	}

	pdfBuffer, err := e.renderer.Render(node, content)
	if err != nil {
		return &ConversionError{
			File:    inputPath,
			Phase:   "PDF rendering",
			Message: "could not render PDF",
			Cause:   err,
		}
	}

	finalOutputPath := e.determineOutputPath(inputPath, outputPath)

	err = os.WriteFile(finalOutputPath, pdfBuffer.Bytes(), 0644)
	if err != nil {
		return &ConversionError{
			File:    inputPath,
			Phase:   "file writing",
			Message: "could not write PDF file",
			Cause:   err,
		}
	}

	return nil
}

func (e *Engine) determineOutputPath(inputPath, outputPath string) string {
	if outputPath != "" {
		return outputPath
	}

	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	return baseName + ".pdf"
}
