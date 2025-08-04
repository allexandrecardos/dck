package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("CLI funcionando! 🎉")
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
