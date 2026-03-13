package cmd

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/allexandrecardos/dck/internal/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var (
	rmTypeContainer bool
	rmTypeImage     bool
	rmTypeVolume    bool
	rmTypeNetwork   bool
	rmDeep          bool
	rmYes           bool
)

var rmCmd = &cobra.Command{
	Use:   "rm [resource]",
	Short: "Remove Docker resources",
	Long:  "Remove containers, images, volumes, or networks with safe, guided flow.",
	Example: "  dck rm api\n" +
		"  dck rm -i nginx:latest\n" +
		"  dck rm --deep api\n" +
		"  dck rm",
	SilenceUsage: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("provide at most one resource identifier")
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

		if len(args) == 0 {
			return rmInteractive(cli)
		}

		id := strings.TrimSpace(args[0])
		if id == "" {
			return errors.New("invalid resource identifier")
		}

		if rmDeep {
			return removeDeepByIdentifier(cli, id)
		}

		return removeByIdentifier(cli, id)
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&rmTypeContainer, "container", "c", false, "Remove container only")
	rmCmd.Flags().BoolVarP(&rmTypeImage, "image", "i", false, "Remove image only")
	rmCmd.Flags().BoolVarP(&rmTypeVolume, "volume", "v", false, "Remove volume only")
	rmCmd.Flags().BoolVarP(&rmTypeNetwork, "network", "n", false, "Remove network only")
	rmCmd.Flags().BoolVar(&rmDeep, "deep", false, "Deep remove (container and related resources)")
	rmCmd.Flags().BoolVar(&rmDeep, "purge", false, "Alias for --deep")
	rmCmd.Flags().BoolVarP(&rmYes, "yes", "y", false, "Skip confirmation prompts")
	rootCmd.AddCommand(rmCmd)
}

func rmInteractive(cli *client.Client) error {
	hasTypeFlag := rmTypeContainer || rmTypeImage || rmTypeVolume || rmTypeNetwork
	if hasTypeFlag && !rmDeep {
		if rmTypeContainer {
			if err := rmInteractiveContainer(cli, false); err != nil {
				return err
			}
		}
		if rmTypeImage {
			if err := rmInteractiveImage(cli); err != nil {
				return err
			}
		}
		if rmTypeVolume {
			if err := rmInteractiveVolume(cli); err != nil {
				return err
			}
		}
		if rmTypeNetwork {
			if err := rmInteractiveNetwork(cli); err != nil {
				return err
			}
		}
		return nil
	}

	options := []string{
		"container",
		"image",
		"volume",
		"network",
		"purge (deep remove)",
	}

	var selected string
	if err := survey.AskOne(&survey.Select{
		Message: "Select resource type to remove:",
		Options: options,
	}, &selected); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			printCommandCanceled()
			return nil
		}
		return err
	}

	switch selected {
	case "container":
		return rmInteractiveContainer(cli, false)
	case "image":
		return rmInteractiveImage(cli)
	case "volume":
		return rmInteractiveVolume(cli)
	case "network":
		return rmInteractiveNetwork(cli)
	case "purge (deep remove)":
		return rmInteractiveContainer(cli, true)
	default:
		return nil
	}
}

func rmInteractiveContainer(cli *client.Client, deep bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	containers, err := docker.ListContainers(ctx, cli, true, false)
	if err != nil {
		return err
	}

	options := make([]string, 0, len(containers))
	lookup := map[string]string{}
	for _, c := range containers {
		id := shortID(c.ID)
		name := strings.TrimPrefix(strings.Join(c.Names, ","), "/")
		statusText := formatStopStatus(c.Status)
		label := fmt.Sprintf("%s (%s) - %s", name, id, colorizeStatus(statusText))
		options = append(options, label)
		lookup[label] = c.ID
	}

	if len(options) == 0 {
		printWarning("No containers found")
		return nil
	}

	var selected []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select a container to remove:",
		Options: options,
	}, &selected); err != nil {
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

	for _, label := range selected {
		id, ok := lookup[label]
		if !ok || id == "" {
			continue
		}
		if deep {
			if err := removeDeepByIdentifier(cli, id); err != nil {
				return err
			}
			continue
		}
		if err := removeContainerWithConfirm(cli, id, label); err != nil {
			return err
		}
	}
	return nil
}

func rmInteractiveImage(cli *client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return err
	}

	options := make([]string, 0, len(images))
	lookup := map[string]string{}
	for _, img := range images {
		label := imageLabel(img)
		options = append(options, label)
		lookup[label] = img.ID
	}

	if len(options) == 0 {
		printWarning("No images found")
		return nil
	}

	var selected []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select an image to remove:",
		Options: options,
	}, &selected); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			printCommandCanceled()
			return nil
		}
		return err
	}

	if len(selected) == 0 {
		printWarningBox("No images selected")
		return nil
	}

	for _, label := range selected {
		id, ok := lookup[label]
		if !ok || id == "" {
			continue
		}
		if err := removeImageWithConfirm(cli, id, label); err != nil {
			return err
		}
	}
	return nil
}

func rmInteractiveVolume(cli *client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	vols, err := cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return err
	}

	options := make([]string, 0, len(vols.Volumes))
	for _, v := range vols.Volumes {
		options = append(options, v.Name)
	}
	sort.Strings(options)

	if len(options) == 0 {
		printWarning("No volumes found")
		return nil
	}

	var selected []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select a volume to remove:",
		Options: options,
	}, &selected); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			printCommandCanceled()
			return nil
		}
		return err
	}

	if len(selected) == 0 {
		printWarningBox("No volumes selected")
		return nil
	}

	for _, name := range selected {
		if strings.TrimSpace(name) == "" {
			continue
		}
		if err := removeVolumeWithConfirm(cli, name); err != nil {
			return err
		}
	}
	return nil
}

func rmInteractiveNetwork(cli *client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	nets, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return err
	}

	options := make([]string, 0, len(nets))
	lookup := map[string]string{}
	for _, n := range nets {
		label := fmt.Sprintf("%s (%s)", n.Name, shortID(n.ID))
		options = append(options, label)
		lookup[label] = n.ID
	}
	sort.Strings(options)

	if len(options) == 0 {
		printWarning("No networks found")
		return nil
	}

	var selected []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select a network to remove:",
		Options: options,
	}, &selected); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			printCommandCanceled()
			return nil
		}
		return err
	}

	if len(selected) == 0 {
		printWarningBox("No networks selected")
		return nil
	}

	for _, label := range selected {
		id, ok := lookup[label]
		if !ok || id == "" {
			continue
		}
		if err := removeNetworkWithConfirm(cli, id, label); err != nil {
			return err
		}
	}
	return nil
}

func removeByIdentifier(cli *client.Client, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	matches, err := resolveMatches(ctx, cli, id)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		printWarning("No matching resources found")
		return nil
	}

	if len(matches) == 1 {
		return removeMatch(cli, matches[0], false)
	}

	var options []string
	for _, m := range matches {
		options = append(options, m.label)
	}

	var selected string
	if err := survey.AskOne(&survey.Select{
		Message: "Multiple resources found. Select one to remove:",
		Options: options,
	}, &selected); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			printCommandCanceled()
			return nil
		}
		return err
	}

	for _, m := range matches {
		if m.label == selected {
			return removeMatch(cli, m, true)
		}
	}
	return nil
}

func removeDeepByIdentifier(cli *client.Client, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	containers, err := docker.ListContainers(ctx, cli, true, false)
	if err != nil {
		return err
	}

	var match *types.Container
	for _, c := range containers {
		name := strings.TrimPrefix(strings.Join(c.Names, ","), "/")
		if strings.HasPrefix(c.ID, id) || name == id || strings.Contains(name, id) {
			match = &c
			break
		}
	}

	if match == nil {
		printWarning("No container found for deep remove")
		return nil
	}

	label := fmt.Sprintf("%s (%s)", strings.TrimPrefix(strings.Join(match.Names, ","), "/"), shortID(match.ID))
	if !confirmAction(fmt.Sprintf("Deep remove %s?", label)) {
		printCommandCanceled()
		return nil
	}

	return deepRemoveContainer(cli, match.ID)
}

type match struct {
	kind  string
	id    string
	label string
}

func resolveMatches(ctx context.Context, cli *client.Client, id string) ([]match, error) {
	typeFilters := map[string]bool{
		"container": rmTypeContainer,
		"image":     rmTypeImage,
		"volume":    rmTypeVolume,
		"network":   rmTypeNetwork,
	}

	hasTypeFlag := rmTypeContainer || rmTypeImage || rmTypeVolume || rmTypeNetwork

	var matches []match

	if !hasTypeFlag || typeFilters["container"] {
		containers, err := docker.ListContainers(ctx, cli, true, false)
		if err != nil {
			return nil, err
		}
		for _, c := range containers {
			name := strings.TrimPrefix(strings.Join(c.Names, ","), "/")
			if strings.HasPrefix(c.ID, id) || name == id || strings.Contains(name, id) {
				label := fmt.Sprintf("container: %s (%s)", name, shortID(c.ID))
				matches = append(matches, match{kind: "container", id: c.ID, label: label})
			}
		}
	}

	if !hasTypeFlag || typeFilters["image"] {
		images, err := cli.ImageList(ctx, types.ImageListOptions{})
		if err != nil {
			return nil, err
		}
		for _, img := range images {
			if imageMatches(img, id) {
				label := fmt.Sprintf("image: %s", imageLabel(img))
				matches = append(matches, match{kind: "image", id: img.ID, label: label})
			}
		}
	}

	if !hasTypeFlag || typeFilters["volume"] {
		vols, err := cli.VolumeList(ctx, volume.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, v := range vols.Volumes {
			if strings.HasPrefix(v.Name, id) || v.Name == id {
				label := fmt.Sprintf("volume: %s", v.Name)
				matches = append(matches, match{kind: "volume", id: v.Name, label: label})
			}
		}
	}

	if !hasTypeFlag || typeFilters["network"] {
		nets, err := cli.NetworkList(ctx, types.NetworkListOptions{})
		if err != nil {
			return nil, err
		}
		for _, n := range nets {
			if strings.HasPrefix(n.ID, id) || n.Name == id || strings.Contains(n.Name, id) {
				label := fmt.Sprintf("network: %s (%s)", n.Name, shortID(n.ID))
				matches = append(matches, match{kind: "network", id: n.ID, label: label})
			}
		}
	}

	return matches, nil
}

func removeMatch(cli *client.Client, m match, confirm bool) error {
	switch m.kind {
	case "container":
		return removeContainerWithConfirm(cli, m.id, m.label, confirm)
	case "image":
		return removeImageWithConfirm(cli, m.id, m.label, confirm)
	case "volume":
		return removeVolumeWithConfirm(cli, m.id, confirm)
	case "network":
		return removeNetworkWithConfirm(cli, m.id, m.label, confirm)
	default:
		return nil
	}
}

func removeContainerWithConfirm(cli *client.Client, id, label string, confirmOverride ...bool) error {
	confirm := true
	if len(confirmOverride) > 0 {
		confirm = confirmOverride[0]
	}
	if confirm && !confirmAction(fmt.Sprintf("Remove %s?", label)) {
		printCommandCanceled()
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := docker.RemoveContainer(ctx, cli, id, true, false); err != nil {
		return err
	}
	printInfo(fmt.Sprintf("Removed %s", label))
	return nil
}

func removeImageWithConfirm(cli *client.Client, id, label string, confirmOverride ...bool) error {
	confirm := true
	if len(confirmOverride) > 0 {
		confirm = confirmOverride[0]
	}
	if confirm && !confirmAction(fmt.Sprintf("Remove %s?", label)) {
		printCommandCanceled()
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := docker.RemoveImage(ctx, cli, id); err != nil {
		return err
	}
	printInfo(fmt.Sprintf("Removed %s", label))
	return nil
}

func removeVolumeWithConfirm(cli *client.Client, name string, confirmOverride ...bool) error {
	confirm := true
	if len(confirmOverride) > 0 {
		confirm = confirmOverride[0]
	}
	if confirm && !confirmAction(fmt.Sprintf("Remove volume %s?", name)) {
		printCommandCanceled()
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := docker.RemoveVolume(ctx, cli, name); err != nil {
		return err
	}
	printInfo(fmt.Sprintf("Removed volume %s", name))
	return nil
}

func removeNetworkWithConfirm(cli *client.Client, id, label string, confirmOverride ...bool) error {
	confirm := true
	if len(confirmOverride) > 0 {
		confirm = confirmOverride[0]
	}
	if confirm && !confirmAction(fmt.Sprintf("Remove %s?", label)) {
		printCommandCanceled()
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := docker.RemoveNetwork(ctx, cli, id); err != nil {
		return err
	}
	printInfo(fmt.Sprintf("Removed %s", label))
	return nil
}

func deepRemoveContainer(cli *client.Client, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := docker.StopContainerIfRunning(ctx, cli, id); err != nil {
		// ignore stop errors; continue removal
	}

	info, err := docker.InspectContainer(ctx, cli, id)
	if err != nil {
		return err
	}

	if err := docker.RemoveContainer(ctx, cli, id, true, false); err != nil {
		return err
	}

	volumes := extractVolumes(info)
	for _, v := range volumes {
		_ = docker.RemoveVolume(ctx, cli, v)
	}

	networks := extractCustomNetworks(info)
	for _, n := range networks {
		_ = docker.RemoveNetwork(ctx, cli, n)
	}

	if imageRemovable(cli, info.Image) {
		_ = docker.RemoveImage(ctx, cli, info.Image)
	}

	printInfo("Deep remove completed")
	return nil
}

func extractVolumes(info types.ContainerJSON) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range info.Mounts {
		if m.Type == "volume" && m.Name != "" && !seen[m.Name] {
			seen[m.Name] = true
			out = append(out, m.Name)
		}
	}
	return out
}

func extractCustomNetworks(info types.ContainerJSON) []string {
	seen := map[string]bool{}
	var out []string
	for name := range info.NetworkSettings.Networks {
		if isDefaultNetwork(name) {
			continue
		}
		if !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	return out
}

func isDefaultNetwork(name string) bool {
	switch name {
	case "bridge", "host", "none":
		return true
	default:
		return false
	}
}

func imageRemovable(cli *client.Client, imageID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	containers, err := docker.ListContainers(ctx, cli, true, false)
	if err != nil {
		return false
	}
	for _, c := range containers {
		if c.ImageID == imageID {
			return false
		}
	}
	return true
}

func imageMatches(img types.ImageSummary, id string) bool {
	if strings.HasPrefix(img.ID, id) {
		return true
	}
	for _, tag := range img.RepoTags {
		if tag == id || strings.Contains(tag, id) {
			return true
		}
	}
	return false
}

func imageLabel(img types.ImageSummary) string {
	tag := "<none>"
	if len(img.RepoTags) > 0 {
		tag = img.RepoTags[0]
	}
	return fmt.Sprintf("%s (%s)", tag, shortID(img.ID))
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func confirmAction(message string) bool {
	if rmYes {
		return true
	}
	confirm := false
	prompt := &survey.Confirm{
		Message: message,
		Default: false,
	}
	if err := survey.AskOne(prompt, &confirm); err != nil {
		return false
	}
	return confirm
}
