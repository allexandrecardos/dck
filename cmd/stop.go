package cmd

import (
	"context"
	"fmt"

	"github.com/allexandrecardos/dck/pkg/docker"
	"github.com/allexandrecardos/dck/pkg/ui"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Seleciona múltiplos containers para parar (teste)",
	Run: func(cmd *cobra.Command, args []string) {
		dockerClient, _ := docker.Client()
		options, err := dockerClient.GetContainerOptions(context.Background(), false)

		if err != nil {
			fmt.Println("Erro ao obter containers:", err)
			return
		}

		labels := make([]string, len(options))
		labelToID := make(map[string]string)

		for i, opt := range options {
			labels[i] = opt.Label
			labelToID[opt.Label] = opt.ID
		}

		selectedLabels, err := ui.MultiSelectPrompt("Selecione os containers para parar:", labels, nil)
		if err != nil {
			fmt.Println("Erro na seleção:", err)
			return
		}

		for _, label := range selectedLabels {
			id := labelToID[label]
			if err := dockerClient.StopContainer(context.Background(), id); err != nil {
				fmt.Printf("Erro ao parar container %s: %v\n", label, err)
			} else {
				fmt.Printf("Container %s parado com sucesso.\n", label)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
