package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dck",
	Short: "DCK Helper CLI",
	Long:  "DCK Helper CLI simplifies day-to-day Docker tasks with safe defaults and friendly output.",
	Example: "  dck ps\n" +
		"  dck run nginx\n" +
		"  dck exec api\n" +
		"  dck stop -a\n" +
		"  dck update --check",
	Version:       buildVersion(),
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Non-blocking update check before any command
		checkForUpdatesSilent()
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printUnexpectedError(err)
		os.Exit(1)
	}
}

func printUnexpectedError(err error) {
	if err == nil {
		return
	}
	printError(fmt.Sprintf("Error: %v", err))
}
