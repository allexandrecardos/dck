package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var initNoBoilerplate bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Docker files in the current folder",
	Long:  "Create Dockerfile and docker-compose.yml in the current directory.",
	Example: "  dck init\n" +
		"  dck init -b",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		files := map[string]string{
			"Dockerfile":         dockerfileTemplate(!initNoBoilerplate),
			"docker-compose.yml": composeTemplate(!initNoBoilerplate),
		}

		for name, content := range files {
			path := filepath.Join(cwd, name)
			if _, err := os.Stat(path); err == nil {
				printInfoBox(fmt.Sprintf("%s already exists, skipping", name))
				continue
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}

			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return err
			}
			printInfoBox(fmt.Sprintf("Created %s", name))
		}

		return nil
	},
}

func init() {
	initCmd.Flags().BoolVarP(&initNoBoilerplate, "boirplate", "b", false, "Disable boilerplate content")
	rootCmd.AddCommand(initCmd)
}

func dockerfileTemplate(withBoilerplate bool) string {
	if !withBoilerplate {
		return "# Dockerfile\n"
	}
	return strings.TrimSpace(`
FROM alpine:3.20

WORKDIR /app

COPY . .

CMD ["sh"]
`) + "\n"
}

func composeTemplate(withBoilerplate bool) string {
	if !withBoilerplate {
		return "services:\n"
	}
	return strings.TrimSpace(`
services:
  app:
    build: .
    container_name: app
    working_dir: /app
    volumes:
      - .:/app
    command: ["sh"]
`) + "\n"
}
