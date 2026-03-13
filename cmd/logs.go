package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/allexandrecardos/dck/internal/docker"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/spf13/cobra"
)

var (
	logsFollow bool
	logsTail   int
)

var logsCmd = &cobra.Command{
	Use:   "logs [container]",
	Short: "Show container logs",
	Long:  "Show logs for a container. Use -f to follow and -n to limit lines.",
	Example: "  dck logs api\n" +
		"  dck logs -f -n 200\n" +
		"  dck logs",
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
			selected, err := selectLogContainer(selectCtx, cli)
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

		ctx := context.Background()
		if !logsFollow {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
		}

		reader, err := docker.LogsContainer(ctx, cli, target, logsFollow, logsTail)
		if err != nil {
			return fmt.Errorf("failed to get logs: %w", err)
		}
		defer reader.Close()

		if _, err := stdcopy.StdCopy(os.Stdout, os.Stderr, reader); err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow logs in real time")
	logsCmd.Flags().IntVarP(&logsTail, "tail", "n", 200, "Number of lines")
	rootCmd.AddCommand(logsCmd)
}

func selectLogContainer(ctx context.Context, cli *client.Client) (string, error) {
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
		Message: "Select a container to view logs:",
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
