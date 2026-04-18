package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonOutput  bool
	outputFile  string
	quiet       bool
	dryRun      bool
)

var rootCmd = &cobra.Command{
	Use:   "gsecret",
	Short: "Retrieve GitHub secrets values easily",
	Long: `gsecret is a CLI tool that retrieves GitHub secret values using GitHub Actions as a secure bridge.
	
Since GitHub's API doesn't expose secret values, gsecret creates a temporary workflow
that safely exports secrets via artifacts, then automatically cleans up all traces.`,
	Version: "0.1.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Write output to file instead of stdout")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress progress messages")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without actually doing it")
}
