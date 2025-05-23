package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fredcamaral/md-to-pdf/internal/config"
	"github.com/spf13/cobra"
)

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
		printConfigValue("  font-family", userConfig.FontFamily, "Arial")
		printConfigValue("  font-size", userConfig.FontSize, 12.0)
		printConfigValue("  heading-scale", userConfig.HeadingScale, 1.5)
		printConfigValue("  line-spacing", userConfig.LineSpacing, 1.2)

		// Code styling
		fmt.Println("\nCode Styling:")
		printConfigValue("  code-font", userConfig.CodeFont, "Courier")
		printConfigValue("  code-size", userConfig.CodeSize, 10.0)

		// Page layout
		fmt.Println("\nPage Layout:")
		printConfigValue("  page-size", userConfig.PageSize, "A4")
		printConfigValue("  margin-top", userConfig.MarginTop, 20.0)
		printConfigValue("  margin-bottom", userConfig.MarginBottom, 20.0)
		printConfigValue("  margin-left", userConfig.MarginLeft, 15.0)
		printConfigValue("  margin-right", userConfig.MarginRight, 15.0)

		// PDF metadata
		fmt.Println("\nPDF Metadata:")
		printConfigValue("  title", userConfig.Title, "")
		printConfigValue("  author", userConfig.Author, "")
		printConfigValue("  subject", userConfig.Subject, "")

		// Mermaid settings
		fmt.Println("\nMermaid Settings:")
		printConfigValue("  mermaid-scale", userConfig.MermaidScale, 2.2)
		printConfigValue("  mermaid-max-width", userConfig.MermaidMaxWidth, 0.0)
		printConfigValue("  mermaid-max-height", userConfig.MermaidMaxHeight, 150.0)

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
	switch key {
	// Typography & Fonts
	case "font-family":
		userConfig.FontFamily = value
	case "font-size":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid font-size: %s", value)
		}
		userConfig.FontSize = v
	case "heading-scale":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid heading-scale: %s", value)
		}
		userConfig.HeadingScale = v
	case "line-spacing":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid line-spacing: %s", value)
		}
		userConfig.LineSpacing = v

	// Code styling
	case "code-font":
		userConfig.CodeFont = value
	case "code-size":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid code-size: %s", value)
		}
		userConfig.CodeSize = v

	// Page layout
	case "page-size":
		if !isValidPageSize(value) {
			return fmt.Errorf("invalid page-size: %s (valid: A4, A3, Letter, Legal)", value)
		}
		userConfig.PageSize = value
	case "margin-top":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid margin-top: %s", value)
		}
		userConfig.MarginTop = v
	case "margin-bottom":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid margin-bottom: %s", value)
		}
		userConfig.MarginBottom = v
	case "margin-left":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid margin-left: %s", value)
		}
		userConfig.MarginLeft = v
	case "margin-right":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid margin-right: %s", value)
		}
		userConfig.MarginRight = v

	// PDF metadata
	case "title":
		userConfig.Title = value
	case "author":
		userConfig.Author = value
	case "subject":
		userConfig.Subject = value

	// Mermaid settings
	case "mermaid-scale":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid mermaid-scale: %s", value)
		}
		userConfig.MermaidScale = v
	case "mermaid-max-width":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid mermaid-max-width: %s", value)
		}
		userConfig.MermaidMaxWidth = v
	case "mermaid-max-height":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid mermaid-max-height: %s", value)
		}
		userConfig.MermaidMaxHeight = v

	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return nil
}

func resetConfigValue(userConfig *config.UserConfig, key string) error {
	switch key {
	case "font-family":
		userConfig.FontFamily = ""
	case "font-size":
		userConfig.FontSize = 0
	case "heading-scale":
		userConfig.HeadingScale = 0
	case "line-spacing":
		userConfig.LineSpacing = 0
	case "code-font":
		userConfig.CodeFont = ""
	case "code-size":
		userConfig.CodeSize = 0
	case "page-size":
		userConfig.PageSize = ""
	case "margin-top":
		userConfig.MarginTop = 0
	case "margin-bottom":
		userConfig.MarginBottom = 0
	case "margin-left":
		userConfig.MarginLeft = 0
	case "margin-right":
		userConfig.MarginRight = 0
	case "title":
		userConfig.Title = ""
	case "author":
		userConfig.Author = ""
	case "subject":
		userConfig.Subject = ""
	case "mermaid-scale":
		userConfig.MermaidScale = 0
	case "mermaid-max-width":
		userConfig.MermaidMaxWidth = 0
	case "mermaid-max-height":
		userConfig.MermaidMaxHeight = 0
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}
	return nil
}

func isValidPageSize(size string) bool {
	validSizes := []string{"A4", "A3", "A5", "Letter", "Legal", "Tabloid"}
	size = strings.ToUpper(size)
	for _, valid := range validSizes {
		if strings.ToUpper(valid) == size {
			return true
		}
	}
	return false
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
}
