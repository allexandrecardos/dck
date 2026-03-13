package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/allexandrecardos/dck/internal/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var startAll bool

var startCmd = &cobra.Command{
	Use:   "start [container...]",
	Short: "Start one or more containers",
	Long:  "Start stopped containers by name/ID, or select interactively when no args are provided.",
	Example: "  dck start api worker\n" +
		"  dck start\n" +
		"  dck start -a",
	SilenceUsage: true,
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

		targets := args
		if startAll {
			listCtx, listCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer listCancel()
			allTargets, err := listStoppedContainers(listCtx, cli)
			if err != nil {
				return err
			}
			if len(allTargets) == 0 {
				printWarning("No stopped containers")
				return nil
			}
			targets = allTargets
		} else if len(targets) == 0 {
			selectCtx, selectCancel := context.WithTimeout(context.Background(), 5*time.Second)
			selected, err := selectStoppedContainers(selectCtx, cli)
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
			if len(selected) == 0 {
				printWarningBox("No containers selected")
				return nil
			}
			targets = selected
		}

		for i, target := range targets {
			if i > 0 {
				fmt.Println("")
			}
			labelCtx, labelCancel := context.WithTimeout(context.Background(), 5*time.Second)
			display := resolveContainerLabel(labelCtx, cli, target)
			labelCancel()

			fmt.Printf("%s %s\n", colorize("["+display+"]", colorYellow), colorize("Starting container...", colorGray))
			opCtx, opCancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := docker.StartContainer(opCtx, cli, target); err != nil {
				opCancel()
				msg := fmt.Sprintf("[%s] Failed to start container: %v", display, err)
				printError(msg)
				continue
			}
			opCancel()
			fmt.Printf("%s %s\n", colorize("["+display+"]", colorGreen), colorize("Container started", colorGray))
		}

		return nil
	},
}

func init() {
	startCmd.Flags().BoolVarP(&startAll, "all", "a", false, "Start all stopped containers")
	rootCmd.AddCommand(startCmd)
}

func selectStoppedContainers(ctx context.Context, cli *client.Client) ([]string, error) {
	containers, err := docker.ListContainers(ctx, cli, true, false)
	if err != nil {
		return nil, err
	}

	options := make([]string, 0, len(containers))
	lookup := map[string]string{}
	for _, c := range containers {
		if !isStartable(c) {
			continue
		}
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
		printWarning("No stopped containers")
		return nil, errNoOptions
	}

	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Select containers to start:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(selected))
	for _, label := range selected {
		if id, ok := lookup[label]; ok {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func listStoppedContainers(ctx context.Context, cli *client.Client) ([]string, error) {
	containers, err := docker.ListContainers(ctx, cli, true, false)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(containers))
	for _, c := range containers {
		if isStartable(c) {
			ids = append(ids, c.ID)
		}
	}
	return ids, nil
}

func isStartable(c types.Container) bool {
	status := strings.ToLower(c.Status)
	if strings.HasPrefix(status, "up") {
		return false
	}
	if strings.Contains(status, "paused") {
		return false
	}
	return true
}
