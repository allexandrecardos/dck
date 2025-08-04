package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("CLI funcionando! 🎉")
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
