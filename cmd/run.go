package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/allexandrecardos/dck/internal/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
)

var (
	runName    string
	runDetach  bool
	runAttach  bool
	runPorts   []string
	runVolumes []string
	runEnvs    []string
)

var runCmd = &cobra.Command{
	Use:   "run <image> [command...]",
	Short: "Run a container with safe defaults",
	Long: `Run a container with a predictable name and safe defaults.

If the image is missing locally, it will be pulled automatically.
When no name is provided, DCK generates <image>-<n> (e.g. nginx-1).
Runs detached by default; use -a to run in foreground.`,
	Example: `  dck run nginx
  dck run nginx -p 8080:80
  dck run postgres:16 --name db -e POSTGRES_PASSWORD=secret
  dck run ubuntu bash`,
	SilenceUsage: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("provide an image to run")
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

		imageInput := strings.TrimSpace(args[0])
		if imageInput == "" {
			return errors.New("image cannot be empty")
		}

		detach := runDetach
		if runAttach {
			detach = false
		}

		image, err := resolveImage(cli, imageInput)
		if err != nil {
			return err
		}
		if image == "" {
			return nil
		}

		name, err := resolveRunName(cli, image, runName)
		if err != nil {
			return err
		}
		if name == "" {
			printCommandCanceled()
			return nil
		}

		ports, portBindings, err := parsePorts(runPorts)
		if err != nil {
			return err
		}

		binds, err := parseVolumes(runVolumes)
		if err != nil {
			return err
		}

		envs, err := parseEnvs(runEnvs)
		if err != nil {
			return err
		}

		cmdArgs := args[1:]

		config := &container.Config{
			Image:        image,
			Cmd:          cmdArgs,
			Env:          envs,
			Tty:          !detach,
			OpenStdin:    !detach,
			AttachStdout: true,
			AttachStderr: true,
			AttachStdin:  !detach,
			ExposedPorts: ports,
		}
		hostConfig := &container.HostConfig{
			PortBindings: portBindings,
			Binds:        binds,
		}

		createCtx, createCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer createCancel()
		resp, err := cli.ContainerCreate(createCtx, config, hostConfig, nil, nil, name)
		if err != nil {
			return err
		}

		if detach {
			if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
				return err
			}
			printRunSummary(name, image, detach, runPorts)
			return nil
		}

		attach, err := cli.ContainerAttach(context.Background(), resp.ID, container.AttachOptions{
			Stream: true,
			Stdin:  true,
			Stdout: true,
			Stderr: true,
		})
		if err != nil {
			return err
		}
		defer attach.Close()

		if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
			return err
		}

		printRunSummary(name, image, detach, runPorts)

		go func() {
			_, _ = io.Copy(attach.Conn, os.Stdin)
		}()

		if config.Tty {
			_, _ = io.Copy(os.Stdout, attach.Reader)
			return nil
		}

		_, _ = stdcopy.StdCopy(os.Stdout, os.Stderr, attach.Reader)
		return nil
	},
}

func init() {
	runCmd.Flags().StringVar(&runName, "name", "", "Container name")
	runCmd.Flags().BoolVarP(&runDetach, "detach", "d", true, "Run in background (default: true)")
	runCmd.Flags().BoolVarP(&runAttach, "attach", "a", false, "Run in foreground")
	runCmd.Flags().StringArrayVarP(&runPorts, "publish", "p", nil, "Publish a container port (host:container[/proto])")
	runCmd.Flags().StringArrayVarP(&runVolumes, "volume", "v", nil, "Bind mount a volume (host:container[:ro])")
	runCmd.Flags().StringArrayVarP(&runEnvs, "env", "e", nil, "Set environment variables (KEY=VALUE)")
	rootCmd.AddCommand(runCmd)
}

func resolveImage(cli *client.Client, input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", errors.New("image cannot be empty")
	}

	if !strings.ContainsAny(input, ":@") {
		selected, ok, err := promptImageTag(cli, input)
		if err != nil {
			if errors.Is(err, terminal.InterruptErr) {
				printCommandCanceled()
				return "", nil
			}
			return "", err
		}
		if ok {
			return selected, nil
		}
	}

	if exists, err := imageExists(cli, input); err == nil && exists {
		return input, nil
	}

	printInfo(fmt.Sprintf("Pulling image: %s", input))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	reader, err := cli.ImagePull(ctx, input, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	defer reader.Close()
	_, _ = io.Copy(io.Discard, reader)

	return input, nil
}

func promptImageTag(cli *client.Client, image string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return "", false, err
	}

	var matches []string
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if strings.HasPrefix(tag, image+":") {
				matches = append(matches, tag)
			}
		}
	}

	if len(matches) <= 1 {
		return "", false, nil
	}

	var selected string
	prompt := &survey.Select{
		Message: "Multiple local tags found. Choose one:",
		Options: matches,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return "", false, err
	}

	if selected == "" {
		return "", false, nil
	}
	return selected, true, nil
}

func imageExists(cli *client.Client, image string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := cli.ImageInspectWithRaw(ctx, image)
	if err == nil {
		return true, nil
	}
	if client.IsErrNotFound(err) {
		return false, nil
	}
	return false, err
}

func resolveRunName(cli *client.Client, image, requested string) (string, error) {
	req := strings.TrimSpace(requested)
	if req != "" {
		if !isValidContainerName(req) {
			return "", fmt.Errorf("invalid container name: %s", req)
		}
		if exists, err := containerNameExists(cli, req); err != nil {
			return "", err
		} else if exists {
			var replacement string
			prompt := &survey.Input{Message: "Name already exists. Enter a new name (leave empty to cancel):"}
			if err := survey.AskOne(prompt, &replacement); err != nil {
				if errors.Is(err, terminal.InterruptErr) {
					printCommandCanceled()
					return "", nil
				}
				return "", err
			}
			replacement = strings.TrimSpace(replacement)
			if replacement == "" {
				return "", nil
			}
			return resolveRunName(cli, image, replacement)
		}
		return req, nil
	}

	base := baseNameFromImage(image)
	if base == "" {
		base = "container"
	}

	used, err := listContainerNames(cli)
	if err != nil {
		return "", err
	}

	for i := 1; i < 1000; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !used[candidate] {
			return candidate, nil
		}
	}

	return "", errors.New("failed to generate a container name")
}

func baseNameFromImage(image string) string {
	noDigest := strings.Split(image, "@")[0]
	lastSlash := strings.LastIndex(noDigest, "/")
	namePart := noDigest
	if lastSlash >= 0 {
		namePart = noDigest[lastSlash+1:]
	}
	lastColon := strings.LastIndex(namePart, ":")
	if lastColon >= 0 {
		namePart = namePart[:lastColon]
	}
	namePart = strings.ToLower(namePart)
	sanitized := sanitizeName(namePart)
	return sanitized
}

func sanitizeName(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	result := strings.Trim(b.String(), "-._")
	return result
}

var containerNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]+$`)

func isValidContainerName(name string) bool {
	return containerNameRe.MatchString(name)
}

func containerNameExists(cli *client.Client, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	containers, err := docker.ListContainers(ctx, cli, true, false)
	if err != nil {
		return false, err
	}
	for _, c := range containers {
		for _, n := range c.Names {
			if strings.TrimPrefix(n, "/") == name {
				return true, nil
			}
		}
	}
	return false, nil
}

func listContainerNames(cli *client.Client) (map[string]bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	containers, err := docker.ListContainers(ctx, cli, true, false)
	if err != nil {
		return nil, err
	}
	used := map[string]bool{}
	for _, c := range containers {
		for _, n := range c.Names {
			name := strings.TrimPrefix(n, "/")
			if name != "" {
				used[name] = true
			}
		}
	}
	return used, nil
}

func parsePorts(inputs []string) (nat.PortSet, nat.PortMap, error) {
	ports := nat.PortSet{}
	bindings := nat.PortMap{}
	for _, raw := range inputs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		port, binding, err := parsePort(raw)
		if err != nil {
			return nil, nil, err
		}
		ports[port] = struct{}{}
		bindings[port] = append(bindings[port], binding)
	}
	return ports, bindings, nil
}

func parsePort(raw string) (nat.Port, nat.PortBinding, error) {
	proto := "tcp"
	main := raw
	if strings.Contains(raw, "/") {
		parts := strings.Split(raw, "/")
		if len(parts) != 2 {
			return "", nat.PortBinding{}, fmt.Errorf("invalid port mapping: %s", raw)
		}
		main = parts[0]
		proto = parts[1]
	}

	segments := strings.Split(main, ":")
	var hostIP, hostPort, containerPort string
	if len(segments) == 2 {
		hostPort = segments[0]
		containerPort = segments[1]
	} else if len(segments) == 3 {
		hostIP = segments[0]
		hostPort = segments[1]
		containerPort = segments[2]
	} else {
		return "", nat.PortBinding{}, fmt.Errorf("invalid port mapping: %s", raw)
	}

	if _, err := strconv.Atoi(hostPort); err != nil {
		return "", nat.PortBinding{}, fmt.Errorf("invalid host port: %s", hostPort)
	}
	if _, err := strconv.Atoi(containerPort); err != nil {
		return "", nat.PortBinding{}, fmt.Errorf("invalid container port: %s", containerPort)
	}

	port, err := nat.NewPort(proto, containerPort)
	if err != nil {
		return "", nat.PortBinding{}, err
	}

	binding := nat.PortBinding{HostIP: hostIP, HostPort: hostPort}
	return port, binding, nil
}

func parseVolumes(inputs []string) ([]string, error) {
	var binds []string
	for _, raw := range inputs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		parts := strings.Split(raw, ":")
		if len(parts) < 2 || len(parts) > 3 {
			return nil, fmt.Errorf("invalid volume mapping: %s", raw)
		}
		if parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid volume mapping: %s", raw)
		}
		binds = append(binds, raw)
	}
	return binds, nil
}

func parseEnvs(inputs []string) ([]string, error) {
	var out []string
	for _, raw := range inputs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if strings.Contains(raw, "=") {
			parts := strings.SplitN(raw, "=", 2)
			if strings.TrimSpace(parts[0]) == "" {
				return nil, fmt.Errorf("invalid env: %s", raw)
			}
			out = append(out, raw)
			continue
		}
		out = append(out, raw+"=")
	}
	return out, nil
}

func printRunSummary(name, image string, detached bool, ports []string) {
	mode := "attached"
	if detached {
		mode = "detached"
	}
	portsText := "none"
	if len(ports) > 0 {
		portsText = strings.Join(ports, ", ")
	}

	printInfo("Container created and started")
	printInfoCompact(fmt.Sprintf("Name: %s", name))
	printInfoCompact(fmt.Sprintf("Image: %s", image))
	printInfoCompact(fmt.Sprintf("Ports: %s", portsText))
	printInfoCompact(fmt.Sprintf("Mode: %s", mode))
}
