package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fredcamaral/md-to-pdf/internal/config"
	"github.com/fredcamaral/md-to-pdf/internal/core"
	"github.com/spf13/cobra"
)

// configKeyType defines the type of a configuration key.
type configKeyType int

const (
	configKeyString configKeyType = iota
	configKeyFloat64
	configKeyPageSize
)

// configCategory groups related configuration keys.
type configCategory string

const (
	categoryTypography configCategory = "Typography"
	categoryCode       configCategory = "Code Styling"
	categoryPage       configCategory = "Page Layout"
	categoryMetadata   configCategory = "PDF Metadata"
	categoryMermaid    configCategory = "Mermaid Settings"
)

// configKeyDef defines metadata for a configuration key including validation rules.
type configKeyDef struct {
	name         string
	category     configCategory
	description  string
	keyType      configKeyType
	defaultValue interface{}
	minValue     float64
	maxValue     float64
	getter       func(*config.UserConfig) interface{}
	setter       func(*config.UserConfig, interface{})
	resetter     func(*config.UserConfig)
}

// configKeys is the single source of truth for all configuration keys.
var configKeys = []configKeyDef{
	// Typography & Fonts
	{
		name:         "font-family",
		category:     categoryTypography,
		description:  "Font family for body text (Arial, Times, Helvetica, Courier)",
		keyType:      configKeyString,
		defaultValue: "Arial",
		getter:       func(c *config.UserConfig) interface{} { return c.FontFamily },
		setter:       func(c *config.UserConfig, v interface{}) { c.FontFamily = v.(string) },
		resetter:     func(c *config.UserConfig) { c.FontFamily = "" },
	},
	{
		name:         "font-size",
		category:     categoryTypography,
		description:  "Font size in points (range: 1-72)",
		keyType:      configKeyFloat64,
		defaultValue: 12.0,
		minValue:     core.FontSizeMin,
		maxValue:     core.FontSizeMax,
		getter:       func(c *config.UserConfig) interface{} { return c.FontSize },
		setter:       func(c *config.UserConfig, v interface{}) { c.FontSize = v.(float64) },
		resetter:     func(c *config.UserConfig) { c.FontSize = 0 },
	},
	{
		name:         "heading-scale",
		category:     categoryTypography,
		description:  "Heading size multiplier (range: 0.1-10.0)",
		keyType:      configKeyFloat64,
		defaultValue: 1.5,
		minValue:     core.HeadingScaleMin,
		maxValue:     core.HeadingScaleMax,
		getter:       func(c *config.UserConfig) interface{} { return c.HeadingScale },
		setter:       func(c *config.UserConfig, v interface{}) { c.HeadingScale = v.(float64) },
		resetter:     func(c *config.UserConfig) { c.HeadingScale = 0 },
	},
	{
		name:         "line-spacing",
		category:     categoryTypography,
		description:  "Line spacing multiplier (range: 0.1-5.0)",
		keyType:      configKeyFloat64,
		defaultValue: 1.2,
		minValue:     core.LineSpacingMin,
		maxValue:     core.LineSpacingMax,
		getter:       func(c *config.UserConfig) interface{} { return c.LineSpacing },
		setter:       func(c *config.UserConfig, v interface{}) { c.LineSpacing = v.(float64) },
		resetter:     func(c *config.UserConfig) { c.LineSpacing = 0 },
	},
	// Code styling
	{
		name:         "code-font",
		category:     categoryCode,
		description:  "Font family for code blocks (typically monospace)",
		keyType:      configKeyString,
		defaultValue: "Courier",
		getter:       func(c *config.UserConfig) interface{} { return c.CodeFont },
		setter:       func(c *config.UserConfig, v interface{}) { c.CodeFont = v.(string) },
		resetter:     func(c *config.UserConfig) { c.CodeFont = "" },
	},
	{
		name:         "code-size",
		category:     categoryCode,
		description:  "Font size for code blocks in points (range: 6-48)",
		keyType:      configKeyFloat64,
		defaultValue: 10.0,
		minValue:     core.CodeSizeMin,
		maxValue:     core.CodeSizeMax,
		getter:       func(c *config.UserConfig) interface{} { return c.CodeSize },
		setter:       func(c *config.UserConfig, v interface{}) { c.CodeSize = v.(float64) },
		resetter:     func(c *config.UserConfig) { c.CodeSize = 0 },
	},
	// Page layout
	{
		name:         "page-size",
		category:     categoryPage,
		description:  "Page size (A3, A4, A5, Letter, Legal, Tabloid)",
		keyType:      configKeyPageSize,
		defaultValue: "A4",
		getter:       func(c *config.UserConfig) interface{} { return c.PageSize },
		setter:       func(c *config.UserConfig, v interface{}) { c.PageSize = v.(string) },
		resetter:     func(c *config.UserConfig) { c.PageSize = "" },
	},
	{
		name:         "margin-top",
		category:     categoryPage,
		description:  "Top margin in mm (range: 0-100)",
		keyType:      configKeyFloat64,
		defaultValue: 20.0,
		minValue:     core.MarginMin,
		maxValue:     core.MarginMax,
		getter:       func(c *config.UserConfig) interface{} { return c.MarginTop },
		setter:       func(c *config.UserConfig, v interface{}) { c.MarginTop = v.(float64) },
		resetter:     func(c *config.UserConfig) { c.MarginTop = 0 },
	},
	{
		name:         "margin-bottom",
		category:     categoryPage,
		description:  "Bottom margin in mm (range: 0-100)",
		keyType:      configKeyFloat64,
		defaultValue: 20.0,
		minValue:     core.MarginMin,
		maxValue:     core.MarginMax,
		getter:       func(c *config.UserConfig) interface{} { return c.MarginBottom },
		setter:       func(c *config.UserConfig, v interface{}) { c.MarginBottom = v.(float64) },
		resetter:     func(c *config.UserConfig) { c.MarginBottom = 0 },
	},
	{
		name:         "margin-left",
		category:     categoryPage,
		description:  "Left margin in mm (range: 0-100)",
		keyType:      configKeyFloat64,
		defaultValue: 15.0,
		minValue:     core.MarginMin,
		maxValue:     core.MarginMax,
		getter:       func(c *config.UserConfig) interface{} { return c.MarginLeft },
		setter:       func(c *config.UserConfig, v interface{}) { c.MarginLeft = v.(float64) },
		resetter:     func(c *config.UserConfig) { c.MarginLeft = 0 },
	},
	{
		name:         "margin-right",
		category:     categoryPage,
		description:  "Right margin in mm (range: 0-100)",
		keyType:      configKeyFloat64,
		defaultValue: 15.0,
		minValue:     core.MarginMin,
		maxValue:     core.MarginMax,
		getter:       func(c *config.UserConfig) interface{} { return c.MarginRight },
		setter:       func(c *config.UserConfig, v interface{}) { c.MarginRight = v.(float64) },
		resetter:     func(c *config.UserConfig) { c.MarginRight = 0 },
	},
	// PDF metadata
	{
		name:         "title",
		category:     categoryMetadata,
		description:  "PDF document title (embedded in PDF metadata)",
		keyType:      configKeyString,
		defaultValue: "",
		getter:       func(c *config.UserConfig) interface{} { return c.Title },
		setter:       func(c *config.UserConfig, v interface{}) { c.Title = v.(string) },
		resetter:     func(c *config.UserConfig) { c.Title = "" },
	},
	{
		name:         "author",
		category:     categoryMetadata,
		description:  "PDF document author (embedded in PDF metadata)",
		keyType:      configKeyString,
		defaultValue: "",
		getter:       func(c *config.UserConfig) interface{} { return c.Author },
		setter:       func(c *config.UserConfig, v interface{}) { c.Author = v.(string) },
		resetter:     func(c *config.UserConfig) { c.Author = "" },
	},
	{
		name:         "subject",
		category:     categoryMetadata,
		description:  "PDF document subject (embedded in PDF metadata)",
		keyType:      configKeyString,
		defaultValue: "",
		getter:       func(c *config.UserConfig) interface{} { return c.Subject },
		setter:       func(c *config.UserConfig, v interface{}) { c.Subject = v.(string) },
		resetter:     func(c *config.UserConfig) { c.Subject = "" },
	},
	// Mermaid settings
	{
		name:         "mermaid-scale",
		category:     categoryMermaid,
		description:  "Mermaid diagram scale factor (range: 0.1-10.0)",
		keyType:      configKeyFloat64,
		defaultValue: 2.2,
		minValue:     core.MermaidScaleMin,
		maxValue:     core.MermaidScaleMax,
		getter:       func(c *config.UserConfig) interface{} { return c.MermaidScale },
		setter:       func(c *config.UserConfig, v interface{}) { c.MermaidScale = v.(float64) },
		resetter:     func(c *config.UserConfig) { c.MermaidScale = 0 },
	},
	{
		name:         "mermaid-max-width",
		category:     categoryMermaid,
		description:  "Max diagram width in mm, 0=page width (range: 0-1000)",
		keyType:      configKeyFloat64,
		defaultValue: 0.0,
		minValue:     core.MermaidDimensionMin,
		maxValue:     core.MermaidDimensionMax,
		getter:       func(c *config.UserConfig) interface{} { return c.MermaidMaxWidth },
		setter:       func(c *config.UserConfig, v interface{}) { c.MermaidMaxWidth = v.(float64) },
		resetter:     func(c *config.UserConfig) { c.MermaidMaxWidth = 0 },
	},
	{
		name:         "mermaid-max-height",
		category:     categoryMermaid,
		description:  "Max diagram height in mm (range: 0-1000)",
		keyType:      configKeyFloat64,
		defaultValue: 150.0,
		minValue:     core.MermaidDimensionMin,
		maxValue:     core.MermaidDimensionMax,
		getter:       func(c *config.UserConfig) interface{} { return c.MermaidMaxHeight },
		setter:       func(c *config.UserConfig, v interface{}) { c.MermaidMaxHeight = v.(float64) },
		resetter:     func(c *config.UserConfig) { c.MermaidMaxHeight = 0 },
	},
}

// findConfigKey looks up a config key definition by name.
func findConfigKey(name string) *configKeyDef {
	for i := range configKeys {
		if configKeys[i].name == name {
			return &configKeys[i]
		}
	}
	return nil
}

// validKeysString returns a comma-separated list of all valid configuration keys.
func validKeysString() string {
	keys := make([]string, len(configKeys))
	for i, k := range configKeys {
		keys[i] = k.name
	}
	return strings.Join(keys, ", ")
}

// getKeysByCategory returns all keys grouped by category in display order.
func getKeysByCategory() map[configCategory][]configKeyDef {
	result := make(map[configCategory][]configKeyDef)
	for _, k := range configKeys {
		result[k.category] = append(result[k.category], k)
	}
	return result
}

// categoryOrder defines the display order for categories.
var categoryOrder = []configCategory{
	categoryTypography,
	categoryCode,
	categoryPage,
	categoryMetadata,
	categoryMermaid,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long:  "View and modify default configuration settings stored in ~/.config/md-to-pdf/config.yaml",
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		userConfig, err := config.LoadUserConfig()
		if err != nil {
			return err
		}

		fmt.Println("Current configuration:")
		fmt.Printf("Config file: %s\n\n", config.GetConfigPath())

		// Typography & Fonts
		fmt.Println("Typography & Fonts:")
		printConfigValueFromKey(userConfig, "font-family")
		printConfigValueFromKey(userConfig, "font-size")
		printConfigValueFromKey(userConfig, "heading-scale")
		printConfigValueFromKey(userConfig, "line-spacing")

		// Code styling
		fmt.Println("\nCode Styling:")
		printConfigValueFromKey(userConfig, "code-font")
		printConfigValueFromKey(userConfig, "code-size")

		// Page layout
		fmt.Println("\nPage Layout:")
		printConfigValueFromKey(userConfig, "page-size")
		printConfigValueFromKey(userConfig, "margin-top")
		printConfigValueFromKey(userConfig, "margin-bottom")
		printConfigValueFromKey(userConfig, "margin-left")
		printConfigValueFromKey(userConfig, "margin-right")

		// PDF metadata
		fmt.Println("\nPDF Metadata:")
		printConfigValueFromKey(userConfig, "title")
		printConfigValueFromKey(userConfig, "author")
		printConfigValueFromKey(userConfig, "subject")

		// Mermaid settings
		fmt.Println("\nMermaid Settings:")
		printConfigValueFromKey(userConfig, "mermaid-scale")
		printConfigValueFromKey(userConfig, "mermaid-max-width")
		printConfigValueFromKey(userConfig, "mermaid-max-height")

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		userConfig, err := config.LoadUserConfig()
		if err != nil {
			return err
		}

		err = setConfigValue(userConfig, key, value)
		if err != nil {
			return err
		}

		err = config.SaveUserConfig(userConfig)
		if err != nil {
			return err
		}

		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset [key]",
	Short: "Reset configuration to defaults",
	Long:  "Reset a specific key or all configuration to defaults",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// Reset all - remove config file
			configPath := config.GetConfigPath()
			err := removeConfigFile(configPath)
			if err != nil {
				return err
			}
			fmt.Println("All configuration reset to defaults")
		} else {
			// Reset specific key
			key := args[0]
			userConfig, err := config.LoadUserConfig()
			if err != nil {
				return err
			}

			err = resetConfigValue(userConfig, key)
			if err != nil {
				return err
			}

			err = config.SaveUserConfig(userConfig)
			if err != nil {
				return err
			}

			fmt.Printf("Reset %s to default\n", key)
		}
		return nil
	},
}

// configKeysJSONMode tracks whether to output JSON format
var configKeysJSONMode bool

var configKeysCmd = &cobra.Command{
	Use:   "keys",
	Short: "List all available configuration keys",
	Long:  "Display all available configuration keys with descriptions and default values",
	RunE: func(cmd *cobra.Command, args []string) error {
		if configKeysJSONMode {
			return printConfigKeysJSON()
		}
		return printConfigKeysText()
	},
}

// printConfigKeysText outputs configuration keys in human-readable format.
func printConfigKeysText() error {
	output := uiOutput

	output.Info("Available configuration keys:")
	output.Println()

	keysByCategory := getKeysByCategory()

	for _, cat := range categoryOrder {
		keys, ok := keysByCategory[cat]
		if !ok {
			continue
		}

		// Print category header
		output.Println(output.Bold(string(cat) + ":"))

		for _, k := range keys {
			// Format: key-name    description (default: value)
			defaultStr := formatDefaultValue(k.defaultValue)
			fmt.Printf("  %s\n", output.Highlight(k.name))
			fmt.Printf("      %s (default: %s)\n", k.description, defaultStr)
		}
		output.Println()
	}

	return nil
}

// configKeyJSON represents a config key in JSON format.
type configKeyJSON struct {
	Name         string      `json:"name"`
	Category     string      `json:"category"`
	Description  string      `json:"description"`
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"default"`
	MinValue     *float64    `json:"min,omitempty"`
	MaxValue     *float64    `json:"max,omitempty"`
}

// printConfigKeysJSON outputs configuration keys in JSON format.
func printConfigKeysJSON() error {
	var keys []configKeyJSON

	for i := range configKeys {
		k := &configKeys[i]
		keyJSON := configKeyJSON{
			Name:         k.name,
			Category:     string(k.category),
			Description:  k.description,
			DefaultValue: k.defaultValue,
		}

		switch k.keyType {
		case configKeyString:
			keyJSON.Type = "string"
		case configKeyFloat64:
			keyJSON.Type = "number"
			// Copy values to avoid pointer issues
			minVal := k.minValue
			maxVal := k.maxValue
			keyJSON.MinValue = &minVal
			keyJSON.MaxValue = &maxVal
		case configKeyPageSize:
			keyJSON.Type = "enum"
		}

		keys = append(keys, keyJSON)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(keys)
}

// formatDefaultValue returns a string representation of a default value.
func formatDefaultValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		if val == "" {
			return "(none)"
		}
		return fmt.Sprintf("%q", val)
	case float64:
		if val == float64(int(val)) {
			return fmt.Sprintf("%.0f", val)
		}
		return fmt.Sprintf("%.1f", val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// printConfigValueFromKey prints a config value using the registry.
func printConfigValueFromKey(userConfig *config.UserConfig, keyName string) {
	keyDef := findConfigKey(keyName)
	if keyDef == nil {
		return
	}
	userValue := keyDef.getter(userConfig)
	printConfigValue("  "+keyName, userValue, keyDef.defaultValue)
}

func printConfigValue(key string, userValue interface{}, defaultValue interface{}) {
	if isZeroValue(userValue) {
		fmt.Printf("%s: %v (default)\n", key, defaultValue)
	} else {
		fmt.Printf("%s: %v\n", key, userValue)
	}
}

func isZeroValue(value interface{}) bool {
	switch v := value.(type) {
	case string:
		return v == ""
	case float64:
		return v == 0
	case int:
		return v == 0
	default:
		return false
	}
}

func setConfigValue(userConfig *config.UserConfig, key, value string) error {
	keyDef := findConfigKey(key)
	if keyDef == nil {
		return fmt.Errorf("unknown configuration key: %s\nValid keys: %s", key, validKeysString())
	}

	switch keyDef.keyType {
	case configKeyString:
		keyDef.setter(userConfig, value)

	case configKeyFloat64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid %s: %s (must be a number)", key, value)
		}
		if v < keyDef.minValue || v > keyDef.maxValue {
			return fmt.Errorf("%s must be between %.1f and %.1f, got %.1f", key, keyDef.minValue, keyDef.maxValue, v)
		}
		keyDef.setter(userConfig, v)

	case configKeyPageSize:
		if !core.IsValidPageSize(value) {
			return fmt.Errorf("invalid page-size: %s (valid: %s)", value, core.ValidPageSizesString())
		}
		keyDef.setter(userConfig, value)
	}

	return nil
}

func resetConfigValue(userConfig *config.UserConfig, key string) error {
	keyDef := findConfigKey(key)
	if keyDef == nil {
		return fmt.Errorf("unknown configuration key: %s\nValid keys: %s", key, validKeysString())
	}

	keyDef.resetter(userConfig)
	return nil
}

func removeConfigFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to remove
	}
	return os.Remove(path)
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configKeysCmd)

	// Add --json flag to keys command
	configKeysCmd.Flags().BoolVar(&configKeysJSONMode, "json", false, "Output in JSON format")
}
