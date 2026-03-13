package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	upDetach     bool
	upForeground bool
	upDry        bool
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start services using docker compose",
	Long:  "Run docker compose up. Detached by default; use -f to stay in foreground.",
	Example: "  dck up\n" +
		"  dck up -f",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		detach := upDetach
		if upForeground {
			detach = false
		}

		command := []string{"compose"}
		if upDry {
			command = append(command, "--dry-run")
		}
		command = append(command, "up")
		if detach {
			command = append(command, "-d")
		}

		return runDockerCompose(command...)
	},
}

func init() {
	upCmd.Flags().BoolVarP(&upDetach, "detach", "d", true, "Run in background (default: true)")
	upCmd.Flags().BoolVarP(&upForeground, "foreground", "f", false, "Run in foreground")
	upCmd.Flags().BoolVar(&upDry, "dry", false, "Show what would be done without executing")
	rootCmd.AddCommand(upCmd)
}

func runDockerCompose(args ...string) error {
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose failed: %w", err)
	}
	return nil
}
