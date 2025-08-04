package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("CLI funcionando! 🎉")
	},
}

func init() {
	rootCmd.AddCommand(rebuildCmd)
}
