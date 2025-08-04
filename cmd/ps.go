package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/allexandrecardos/dck/pkg/docker"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "Lista os containers em execução",
	Run: func(cmd *cobra.Command, args []string) {

		dockerClient, err := docker.Client()
		if err != nil {
			fmt.Printf("Erro ao criar cliente Docker: %v\n", err)
			return
		}

		containers, err := dockerClient.ListContainers(context.Background(), true)
		if err != nil {
			fmt.Printf("Erro ao listar containers: %v\n", err)
			return
		}

		if len(containers) == 0 {
			fmt.Println("⚠️ Nenhum container encontrado.")
			return
		}

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleLight)
		t.Style().Options.SeparateRows = true
		t.AppendHeader(table.Row{"id", "image", "status", "ports"})

		for _, container := range containers {
			idShort := container.ID
			if len(idShort) > 12 {
				idShort = idShort[:12]
			}

			t.AppendRow(table.Row{
				idShort,
				container.Image,
				text.FgGreen.Sprint(container.Status),
				container.Ports,
			})
		}

		t.Render()
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
}
