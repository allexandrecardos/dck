package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"
)

var (
	uninstallPurge bool
	uninstallYes   bool
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall dck",
	Long:  "Remove the dck binary from the installation folder.",
	Example: "  dck uninstall\n" +
		"  dck uninstall --purge",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		exe, err := os.Executable()
		if err != nil {
			return err
		}

		baseDir := filepath.Dir(exe)
		configPath := filepath.Join(baseDir, "dck-config.yml")

		if !uninstallYes {
			lines := []string{fmt.Sprintf("This will remove: %s", exe)}
			if uninstallPurge {
				lines = append(lines, fmt.Sprintf("This will also remove: %s", configPath))
			}
			if !confirmUninstall(strings.Join(lines, "\n")) {
				printCommandCanceled()
				return nil
			}
		}

		if runtime.GOOS == "windows" {
			if err := scheduleWindowsUninstall(exe, configPath, uninstallPurge); err != nil {
				return err
			}
			printInfo("Uninstall scheduled. dck will be removed after this command exits.")
			printInfo("If you added the install folder to PATH, remove it manually.")
			return nil
		}

		if err := os.Remove(exe); err != nil {
			return fmt.Errorf("failed to remove %s: %w", exe, err)
		}

		if uninstallPurge {
			if err := removeIfExists(configPath); err != nil {
				return err
			}
		}

		printInfo("dck removed")
		return nil
	},
}

func init() {
	uninstallCmd.Flags().BoolVar(&uninstallPurge, "purge", false, "Also remove dck-config.yml")
	uninstallCmd.Flags().BoolVarP(&uninstallYes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(uninstallCmd)
}

func confirmUninstall(message string) bool {
	confirm := false
	prompt := &survey.Confirm{
		Message: message + "\nProceed?",
		Default: false,
	}
	if err := survey.AskOne(prompt, &confirm); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			return false
		}
		return false
	}
	return confirm
}

func scheduleWindowsUninstall(exePath, configPath string, purge bool) error {
	tmpFile, err := os.CreateTemp("", "dck-uninstall-*.cmd")
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	lines := []string{
		"@echo off",
		"ping 127.0.0.1 -n 2 > nul",
		fmt.Sprintf("del /f /q \"%s\"", exePath),
	}
	if purge {
		lines = append(lines, fmt.Sprintf("del /f /q \"%s\"", configPath))
	}
	lines = append(lines, "del /f /q \"%~f0\"")

	if _, err := tmpFile.WriteString(strings.Join(lines, "\r\n") + "\r\n"); err != nil {
		return err
	}

	cmd := exec.Command("cmd", "/c", "start", "", "/b", tmpFile.Name())
	return cmd.Start()
}

func removeIfExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove %s: %w", path, err)
	}
	return nil
}
