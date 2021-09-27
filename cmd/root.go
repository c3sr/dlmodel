package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "dlmodel",
	Short: "DLModel manages model manifests for deep learning models.",
	Long: `DLModel manages model manifests for deep learning models,
including validating and downloading model manifests.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	// rootCmd.AddCommand(validateCmd)
}
