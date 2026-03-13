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

var stopTimeout int
var stopAll bool

var stopCmd = &cobra.Command{
	Use:   "stop [container...]",
	Short: "Stop one or more containers",
	Long:  "Stop containers by name/ID, or select interactively when no args are provided.",
	Example: "  dck stop api worker\n" +
		"  dck stop\n" +
		"  dck stop -a",
	SilenceUsage: true,
	Args: func(cmd *cobra.Command, args []string) error {
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

		targets := args
		if stopAll {
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
			defer selectCancel()
			selected, err := selectContainers(selectCtx, cli)
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

		timeout := time.Duration(stopTimeout) * time.Second
		var stopErrs []string
		for i, target := range targets {
			if i > 0 {
				fmt.Println("")
			}

			labelCtx, labelCancel := context.WithTimeout(context.Background(), 5*time.Second)
			display := resolveContainerLabel(labelCtx, cli, target)
			labelCancel()

			fmt.Printf("%s %s\n", colorize("["+display+"]", colorBlue), colorize("Stopping container...", colorGray))
			stopCtx, stopCancel := context.WithTimeout(context.Background(), timeout+5*time.Second)
			if err := docker.StopContainer(stopCtx, cli, target, timeout); err != nil {
				stopCancel()
				msg := fmt.Sprintf("[%s] Failed to stop container: %v", display, err)
				printError(msg)
				stopErrs = append(stopErrs, msg)
				continue
			}
			stopCancel()
			fmt.Printf("%s %s\n", colorize("["+display+"]", colorGreen), colorize("Container stopped", colorGray))
		}

		return nil
	},
}

func init() {
	stopCmd.Flags().IntVarP(&stopTimeout, "time", "t", 10, "Timeout in seconds before force")
	stopCmd.Flags().BoolVarP(&stopAll, "all", "a", false, "Stop all running containers")
	rootCmd.AddCommand(stopCmd)
}

func selectContainers(ctx context.Context, cli *client.Client) ([]string, error) {
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
		Message: "Select containers to stop:",
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

func isRunning(c types.Container) bool {
	return strings.HasPrefix(strings.ToLower(c.Status), "up")
}

func listRunningContainers(ctx context.Context, cli *client.Client) ([]string, error) {
	containers, err := docker.ListContainers(ctx, cli, true, false)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(containers))
	for _, c := range containers {
		if isRunning(c) {
			ids = append(ids, c.ID)
		}
	}
	return ids, nil
}

func resolveContainerLabel(ctx context.Context, cli *client.Client, idOrName string) string {
	containers, err := docker.ListContainers(ctx, cli, true, false)
	if err != nil {
		return idOrName
	}

	for _, c := range containers {
		name := strings.TrimPrefix(strings.Join(c.Names, ","), "/")
		if strings.HasPrefix(c.ID, idOrName) || name == idOrName || strings.Contains(name, idOrName) {
			if name == "" {
				return idOrName
			}
			return name
		}
	}

	return idOrName
}

func formatStopStatus(status string) string {
	s := strings.ToLower(status)
	switch {
	case strings.Contains(s, "paused"):
		return "paused"
	case strings.Contains(s, "restarting"):
		return "restarting"
	case strings.Contains(s, "unhealthy"):
		return "unhealthy"
	case strings.Contains(s, "healthy"):
		return "healthy"
	case strings.HasPrefix(s, "up"):
		return "running"
	case strings.Contains(s, "exited"):
		return "exited"
	case strings.Contains(s, "created"):
		return "created"
	case strings.Contains(s, "dead"):
		return "dead"
	default:
		return "unknown"
	}
}

func colorizeStatus(status string) string {
	switch status {
	case "running", "healthy":
		return colorize(status, colorGreen)
	case "paused", "restarting", "created":
		return colorize(status, colorYellow)
	case "unhealthy", "exited", "dead":
		return colorize(status, colorRed)
	default:
		return colorize(status, colorGray)
	}
}
