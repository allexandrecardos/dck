package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("CLI funcionando! 🎉")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
