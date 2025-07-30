package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dck",
	Short: "DCK - Docker CLI Helper",
	Long:  `DCK is a Docker CLI helper that provides interactive interfaces and simplified commands for common Docker operations.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// rootCmd.AddCommand(versionCmd)
}
