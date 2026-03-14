package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var updateCheckOnly bool

var versionCmd = &cobra.Command{
	Use:          "version",
	Short:        "Show dck version",
	Long:         "Print the installed DCK version information.",
	Example:      "  dck version",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		printInfo(formatVersionInfo())
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates",
	Long:  "Check GitHub releases to see if a newer version is available.",
	Example: "  dck update\n" +
		"  dck update --check",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()

		latest, url, err := fetchLatestRelease(ctx)
		if err != nil {
			return err
		}

		if updateCheckOnly {
			printUpdateStatus(latest, url)
			return nil
		}

		if version != "dev" && !isNewerVersion(latest, version) {
			printInfo(fmt.Sprintf("You are up to date (%s)", version))
			return nil
		}

		if err := runSelfUpdate(ctx, latest); err != nil {
			return err
		}
		printInfo(fmt.Sprintf("Updated to %s", latest))
		return nil
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheckOnly, "check", false, "Only check for updates")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
}

type releaseResponse struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func fetchLatestRelease(ctx context.Context) (string, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repoSlug)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "dck-cli")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", "", fmt.Errorf("failed to check updates: %s", resp.Status)
	}

	var data releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", err
	}

	tag := strings.TrimSpace(data.TagName)
	if tag == "" {
		return "", "", errors.New("latest release not found")
	}

	return tag, data.HTMLURL, nil
}

func runSelfUpdate(ctx context.Context, latest string) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return err
	}

	asset, err := resolveAssetName(latest)
	if err != nil {
		return err
	}
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repoSlug, latest, asset)

	tmpFile, err := downloadToTemp(ctx, downloadURL)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		if err := scheduleWindowsReplace(exePath, tmpFile); err != nil {
			return err
		}
		return nil
	}

	if err := os.Chmod(tmpFile, 0755); err != nil {
		return err
	}
	if err := replaceBinary(tmpFile, exePath); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("failed to replace binary: %w (try running with elevated permissions or reinstall in a user-writable folder)", err)
		}
		return err
	}
	return nil
}

func resolveAssetName(tag string) (string, error) {
	versionPart := normalizeVersion(tag)
	if versionPart == "" {
		return "", errors.New("invalid release tag")
	}
	versionPart = "v" + versionPart

	osPart := runtime.GOOS
	archPart := runtime.GOARCH

	switch osPart {
	case "windows":
		if archPart != "amd64" {
			return "", fmt.Errorf("unsupported architecture: %s/%s", osPart, archPart)
		}
		return fmt.Sprintf("dck_%s_windows_amd64.exe", versionPart), nil
	case "linux":
		if archPart != "amd64" && archPart != "arm64" {
			return "", fmt.Errorf("unsupported architecture: %s/%s", osPart, archPart)
		}
		return fmt.Sprintf("dck_%s_linux_%s", versionPart, archPart), nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", osPart)
	}
}

func downloadToTemp(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "dck-cli")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}

	tmp, err := os.CreateTemp("", "dck-update-*")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return "", err
	}

	return tmp.Name(), nil
}

func scheduleWindowsReplace(exePath, newPath string) error {
	script, err := os.CreateTemp("", "dck-update-*.cmd")
	if err != nil {
		return err
	}
	defer script.Close()

	lines := []string{
		"@echo off",
		"ping 127.0.0.1 -n 2 > nul",
		fmt.Sprintf("move /y \"%s\" \"%s\"", newPath, exePath),
		"del /f /q \"%~f0\"",
	}

	if _, err := script.WriteString(strings.Join(lines, "\r\n") + "\r\n"); err != nil {
		return err
	}

	cmd := exec.Command("cmd", "/c", "start", "", "/b", script.Name())
	return cmd.Start()
}

func replaceBinary(tmpPath, exePath string) error {
	dir := filepath.Dir(exePath)
	staged, err := os.CreateTemp(dir, "dck-update-*")
	if err != nil {
		return fmt.Errorf("failed to stage update: %w", err)
	}
	stagedPath := staged.Name()
	defer staged.Close()

	src, err := os.Open(tmpPath)
	if err != nil {
		return err
	}
	defer src.Close()

	if _, err := io.Copy(staged, src); err != nil {
		return err
	}
	if err := staged.Chmod(0755); err != nil {
		return err
	}

	if err := os.Rename(stagedPath, exePath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}
	_ = os.Remove(tmpPath)
	return nil
}

func printUpdateStatus(latest, url string) {
	if version == "dev" {
		printInfo(fmt.Sprintf("Latest version: %s", latest))
		if url != "" {
			printInfo(fmt.Sprintf("Release: %s", url))
		}
		return
	}

	if isNewerVersion(latest, version) {
		printWarning(fmt.Sprintf("Update available: %s (current %s)", latest, version))
		if url != "" {
			printInfo(fmt.Sprintf("Release: %s", url))
		}
		return
	}

	printInfo(fmt.Sprintf("You are up to date (%s)", version))
}

func checkForUpdatesSilent() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	latest, url, err := fetchLatestRelease(ctx)
	if err != nil {
		return
	}

	if version == "dev" {
		return
	}

	if isNewerVersion(latest, version) {
		printWarning(fmt.Sprintf("Update available: %s (current %s)", latest, version))
		if url != "" {
			printInfo(fmt.Sprintf("Release: %s", url))
		}
	}
}

func isNewerVersion(latest, current string) bool {
	l := normalizeVersion(latest)
	c := normalizeVersion(current)
	if l == "" || c == "" {
		return latest != current
	}

	la := strings.Split(l, ".")
	ca := strings.Split(c, ".")
	for len(la) < 3 {
		la = append(la, "0")
	}
	for len(ca) < 3 {
		ca = append(ca, "0")
	}

	for i := 0; i < 3; i++ {
		if la[i] == ca[i] {
			continue
		}
		return comparePart(la[i], ca[i]) > 0
	}
	return false
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	if v == "" {
		return ""
	}
	return v
}

func comparePart(a, b string) int {
	ai := toInt(a)
	bi := toInt(b)
	switch {
	case ai > bi:
		return 1
	case ai < bi:
		return -1
	default:
		return 0
	}
}

func toInt(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}
