package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "md-to-pdf",
	Short: "Convert Markdown files to PDF",
	Long:  "A CLI tool to convert Markdown documents to PDF format with plugin support",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}