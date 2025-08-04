package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Comando de teste",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("CLI funcionando! 🎉")
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
