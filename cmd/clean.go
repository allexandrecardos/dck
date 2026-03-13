package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:          "clean",
	Short:        "Remove unused Docker resources",
	Long:         "Prune unused containers, images, networks, and volumes with confirmation.",
	Example:      "  dck clean",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		warn := "This will remove:\n" +
			"- stopped containers\n" +
			"- unused images (including dangling)\n" +
			"- unused networks\n" +
			"- unused volumes\n"

		printWarning("Clean will remove unused Docker resources.")
		printWarning(warn)

		if !confirmAction("Proceed with clean?") {
			printCommandCanceled()
			return nil
		}

		execCmd := exec.Command("docker", "system", "prune", "-a", "--volumes", "-f")
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		if err := execCmd.Run(); err != nil {
			return fmt.Errorf("docker system prune failed: %w", err)
		}

		printInfo("Clean completed.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
