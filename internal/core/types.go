package core

// Config holds all configuration for the conversion engine
type Config struct {
	Parser   ParserConfig
	Renderer RenderConfig
	Plugins  PluginConfig
	Output   OutputConfig
	Document DocumentConfig
}

type ParserConfig struct {
	Extensions []string
}

type RenderConfig struct {
	PageSize     string
	Margins      Margins
	FontFamily   string
	FontSize     float64
	HeadingScale float64
	LineSpacing  float64
	CodeFont     string
	CodeSize     float64
	Mermaid      MermaidConfig
}

type MermaidConfig struct {
	Scale     float64 // Scaling factor for mermaid diagrams (1.0 = normal, 1.4 = 40% bigger)
	MaxWidth  float64 // Maximum width in mm (0 = use page width)
	MaxHeight float64 // Maximum height in mm
}

type PluginConfig struct {
	Directory string
	Enabled   bool
	// Configs holds per-plugin configuration keyed by plugin name
	Configs map[string]map[string]interface{}
}

type OutputConfig struct {
	Path    string
	Quality string
}

type DocumentConfig struct {
	Title   string
	Author  string
	Subject string
}

type Margins struct {
	Top    float64
	Bottom float64
	Left   float64
	Right  float64
}

// ProgressCallback is called during conversion to report progress.
// It receives the current file index (1-based), total file count, input filename, and output filename.
type ProgressCallback func(current, total int, inputFile, outputFile string)

type ConversionOptions struct {
	InputFiles []string
	OutputPath string
	PluginDir  string
	Verbose    bool
	// OnProgress is called before converting each file (optional).
	OnProgress ProgressCallback
	// OnComplete is called after successfully converting each file (optional).
	OnComplete ProgressCallback
}
