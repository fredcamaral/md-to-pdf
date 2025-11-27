package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// These variables are set at build time via ldflags
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  "Print the version, commit hash, and build date of md-to-pdf",
	Run: func(cmd *cobra.Command, args []string) {
		// Try to get version from build info if not set via ldflags
		if Version == "dev" {
			if info, ok := debug.ReadBuildInfo(); ok {
				if info.Main.Version != "" && info.Main.Version != "(devel)" {
					Version = info.Main.Version
				}
			}
		}

		fmt.Printf("md-to-pdf version %s\n", Version)
		if Commit != "unknown" {
			fmt.Printf("  commit: %s\n", Commit)
		}
		if Date != "unknown" {
			fmt.Printf("  built:  %s\n", Date)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
