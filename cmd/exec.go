package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/allexandrecardos/dck/internal/docker"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var execShell string
var execCmdText string
var execWorkdir string
var execUser string

var execCmd = &cobra.Command{
	Use:   "exec [container]",
	Short: "Enter a container shell",
	Long:  "Open an interactive shell inside a container with auto-detection of available shells.",
	Example: "  dck exec api\n" +
		"  dck exec -s /bin/sh api\n" +
		"  dck exec -c \"ls -la\" api",
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
			selected, err := selectRunningContainerPrompt(selectCtx, cli)
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

		shell := execShell
		if execCmdText != "" {
			if shell == "" {
				shell = "/bin/sh"
			}
			return runDockerExec(target, buildExecOptions(), shell, "-c", execCmdText)
		}

		if shell == "" {
			found, err := detectShell(target)
			if err != nil {
				return err
			}
			shell = found
		}

		if shell == "" {
			printWarning("No known shell found in this container.")
			fmt.Print("Enter a command to run: ")
			command, err := readLine(os.Stdin)
			if err != nil {
				return err
			}
			command = strings.TrimSpace(command)
			if command == "" {
				printWarningBox("No command provided")
				return nil
			}
			return runDockerExec(target, buildExecOptions(), strings.Fields(command)...)
		}

		printInfo(fmt.Sprintf("Using shell: %s", shell))
		return runDockerExec(target, buildExecOptions(), shell)
	},
}

func init() {
	execCmd.Flags().StringVarP(&execShell, "shell", "s", "", "Shell to use (skip auto-detection)")
	execCmd.Flags().StringVarP(&execCmdText, "cmd", "c", "", "Command to run inside the container")
	execCmd.Flags().StringVarP(&execWorkdir, "workdir", "w", "", "Working directory inside the container")
	execCmd.Flags().StringVarP(&execUser, "user", "u", "", "User to run the command as")
	rootCmd.AddCommand(execCmd)
}

func selectRunningContainerPrompt(ctx context.Context, cli *client.Client) (string, error) {
	containers, err := docker.ListContainers(ctx, cli, true, false)
	if err != nil {
		return "", err
	}

	items := make([]string, 0, len(containers))
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
		items = append(items, label)
		lookup[label] = c.ID
	}

	if len(items) == 0 {
		printWarning("No running containers")
		return "", errNoOptions
	}

	var selected string
	prompt := &survey.Select{
		Message: "Select a container:",
		Options: items,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			return "", nil
		}
		return "", err
	}

	if id, ok := lookup[selected]; ok {
		return id, nil
	}
	return "", nil
}

func detectShell(container string) (string, error) {
	candidates := []string{"/bin/bash", "/bin/sh", "sh"}
	for _, shell := range candidates {
		if testShell(container, shell) {
			return shell, nil
		}
	}
	return "", nil
}

func testShell(container, shell string) bool {
	cmd := exec.Command("docker", "exec", container, shell, "-c", "exit")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

type execOptions struct {
	Workdir string
	User    string
}

func buildExecOptions() execOptions {
	return execOptions{
		Workdir: execWorkdir,
		User:    execUser,
	}
}

func runDockerExec(container string, opts execOptions, command ...string) error {
	if len(command) == 0 {
		return errors.New("no command provided")
	}

	args := []string{"exec", "-it"}
	if opts.User != "" {
		args = append(args, "-u", opts.User)
	}
	if opts.Workdir != "" {
		args = append(args, "-w", opts.Workdir)
	}
	args = append(args, container)
	args = append(args, command...)
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func readLine(r io.Reader) (string, error) {
	reader := bufio.NewReader(r)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
