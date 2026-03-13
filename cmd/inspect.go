package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/allexandrecardos/dck/internal/docker"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [container]",
	Short: "Inspect a container",
	Long:  "Show the full JSON output of docker inspect for a container.",
	Example: "  dck inspect api\n" +
		"  dck inspect",
	SilenceUsage: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("provide at most one container")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cli, err := docker.New()
		if err != nil {
			return err
		}

		pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer pingCancel()
		if err := docker.Ping(pingCtx, cli); err != nil {
			return fmt.Errorf("docker unavailable: %w", err)
		}

		var target string
		if len(args) == 1 {
			target = args[0]
		} else {
			selectCtx, selectCancel := context.WithTimeout(context.Background(), 5*time.Second)
			selected, err := selectAnyContainer(selectCtx, cli)
			selectCancel()
			if err != nil {
				if errors.Is(err, errNoOptions) {
					return nil
				}
				if errors.Is(err, terminal.InterruptErr) {
					printCommandCanceled()
					return nil
				}
				return err
			}
			if selected == "" {
				printWarningBox("No containers selected")
				return nil
			}
			target = selected
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		info, err := docker.InspectContainer(ctx, cli, target)
		if err != nil {
			return fmt.Errorf("failed to inspect container: %w", err)
		}

		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}

		_, err = os.Stdout.Write(data)
		if err == nil {
			_, _ = os.Stdout.Write([]byte("\n"))
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}

func selectAnyContainer(ctx context.Context, cli *client.Client) (string, error) {
	containers, err := docker.ListContainers(ctx, cli, true, false)
	if err != nil {
		return "", err
	}

	options := make([]string, 0, len(containers))
	lookup := map[string]string{}
	for _, c := range containers {
		id := c.ID
		if len(id) > 12 {
			id = id[:12]
		}
		name := strings.TrimPrefix(strings.Join(c.Names, ","), "/")
		statusText := formatStopStatus(c.Status)
		label := fmt.Sprintf("%s (%s) - %s", name, id, colorizeStatus(statusText))
		options = append(options, label)
		lookup[label] = c.ID
	}

	if len(options) == 0 {
		printWarning("No containers found")
		return "", errNoOptions
	}

	var selected string
	prompt := &survey.Select{
		Message: "Select a container to inspect:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return "", err
	}

	if id, ok := lookup[selected]; ok {
		return id, nil
	}
	return "", nil
}
