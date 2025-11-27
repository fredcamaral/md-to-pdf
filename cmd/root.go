package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "md-to-pdf",
	Short: "Convert Markdown files to PDF",
	Long: `A CLI tool to convert Markdown documents to PDF format with plugin support.

Use "md-to-pdf convert" to convert files, or "md-to-pdf --help" for more information.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("No command specified. Use 'md-to-pdf convert <file.md>' to convert files.")
		fmt.Println()
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
