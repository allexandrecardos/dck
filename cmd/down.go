package cmd

import (
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:          "down",
	Short:        "Stop services using docker compose",
	Long:         "Run docker compose down to stop and remove services.",
	Example:      "  dck down",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDockerCompose("compose", "down")
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
