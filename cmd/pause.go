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
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var pauseAll bool

var pauseCmd = &cobra.Command{
	Use:   "pause [container...]",
	Short: "Pause one or more containers",
	Long:  "Pause running containers by name/ID, or select interactively when no args are provided.",
	Example: "  dck pause api worker\n" +
		"  dck pause\n" +
		"  dck pause -a",
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
		if pauseAll {
			listCtx, listCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer listCancel()
			allTargets, err := listRunningContainers(listCtx, cli)
			if err != nil {
				return err
			}
			if len(allTargets) == 0 {
				printWarning("No running containers")
				return nil
			}
			targets = allTargets
		} else if len(targets) == 0 {
			selectCtx, selectCancel := context.WithTimeout(context.Background(), 5*time.Second)
			selected, err := selectRunningContainers(selectCtx, cli)
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

			fmt.Printf("%s %s\n", colorize("["+display+"]", colorYellow), colorize("Pausing container...", colorGray))
			opCtx, opCancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := docker.PauseContainer(opCtx, cli, target); err != nil {
				opCancel()
				msg := fmt.Sprintf("[%s] Failed to pause container: %v", display, err)
				printError(msg)
				continue
			}
			opCancel()
			fmt.Printf("%s %s\n", colorize("["+display+"]", colorGreen), colorize("Container paused", colorGray))
		}

		return nil
	},
}

func init() {
	pauseCmd.Flags().BoolVarP(&pauseAll, "all", "a", false, "Pause all running containers")
	rootCmd.AddCommand(pauseCmd)
}

func selectRunningContainers(ctx context.Context, cli *client.Client) ([]string, error) {
	containers, err := docker.ListContainers(ctx, cli, true, false)
	if err != nil {
		return nil, err
	}

	options := make([]string, 0, len(containers))
	lookup := map[string]string{}
	for _, c := range containers {
		if !isRunning(c) {
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
		printWarning("No running containers")
		return nil, errNoOptions
	}

	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Select containers to pause:",
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
