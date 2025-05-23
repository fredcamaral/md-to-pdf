package core

func DefaultConfig() *Config {
	return &Config{
		Parser: ParserConfig{
			Extensions: []string{},
		},
		Renderer: RenderConfig{
			PageSize:     "A4",
			FontFamily:   "Arial",
			FontSize:     12,
			HeadingScale: 1.5, // Headings 50% bigger than base font
			LineSpacing:  1.2, // 20% line spacing
			CodeFont:     "Courier",
			CodeSize:     10, // Code slightly smaller than base font
			Margins: Margins{
				Top:    20,
				Bottom: 20,
				Left:   15,
				Right:  15,
			},
			Mermaid: MermaidConfig{
				Scale:     2.2,   // Double size + 20% by default
				MaxWidth:  0,     // Use page width
				MaxHeight: 150.0, // 150mm max height
			},
		},
		Plugins: PluginConfig{
			Directory: "./plugins",
			Enabled:   true,
		},
		Output: OutputConfig{
			Quality: "standard",
		},
		Document: DocumentConfig{
			Title:   "",
			Author:  "",
			Subject: "",
		},
	}
}
