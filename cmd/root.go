package cmd

import (
	"os"

	"github.com/fredcamaral/md-to-pdf/internal/ui"
	"github.com/spf13/cobra"
)

// uiOutput is the shared UI output instance for colored terminal output.
var uiOutput = ui.NewOutput()

var rootCmd = &cobra.Command{
	Use:   "md-to-pdf",
	Short: "Convert Markdown files to PDF",
	Long: `A CLI tool to convert Markdown documents to PDF format with plugin support.

Use "md-to-pdf convert" to convert files, or "md-to-pdf --help" for more information.`,
	Run: func(cmd *cobra.Command, args []string) {
		uiOutput.Info("No command specified. Use 'md-to-pdf convert <file.md>' to convert files.")
		uiOutput.Println()
		_ = cmd.Help()
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		uiOutput.Errorf("%v", err)
		os.Exit(1)
	}
}

// GetUIOutput returns the shared UI output instance.
func GetUIOutput() *ui.Output {
	return uiOutput
}
