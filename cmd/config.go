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
	"github.com/allexandrecardos/dck/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:          "config",
	Short:        "Open dck-config.yml",
	Long:         "Open (or create) dck-config.yml in the CLI install folder.",
	Example:      "  dck config",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		baseDir := filepath.Dir(exe)
		configPath := filepath.Join(baseDir, "dck-config.yml")

		if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
			if err := os.WriteFile(configPath, []byte(defaultConfigTemplate()), 0644); err != nil {
				return err
			}
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}

		editor := strings.TrimSpace(os.Getenv("EDITOR"))
		if editor == "" {
			editor = strings.TrimSpace(cfg.Editor)
		}

		if editor == "" {
			choice, err := promptEditor()
			if err != nil {
				if errors.Is(err, terminal.InterruptErr) {
					printCommandCanceled()
					return nil
				}
				return err
			}
			editor = choice
			if err := validateEditor(editor); err != nil {
				printEditorError(editor, err)
				return nil
			}
			cfg.Editor = editor
			if err := config.Save(configPath, cfg); err != nil {
				return err
			}
		}

		if editor == "" {
			printInfo(fmt.Sprintf("Config file: %s", configPath))
			printInfo("Set EDITOR to open it automatically (e.g., set EDITOR=code or notepad).")
			return nil
		}

		if err := validateEditor(editor); err != nil {
			printEditorError(editor, err)
			return nil
		}

		if err := runEditor(editor, configPath); err != nil {
			printEditorRunError(editor, err)
			return nil
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func defaultConfigTemplate() string {
	return "editor: \"\"\nps:\n  columns:\n    - id\n    - name\n    - image\n    - status\n    - created\n    - cpu\n    - mem\n    - network\n    - size\n    - command\n    - ports\n"
}

func runEditor(editor, path string) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", editor, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return errors.New("invalid EDITOR value")
	}
	bin := parts[0]
	args := append(parts[1:], path)
	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func validateEditor(editor string) error {
	bin := editorBinary(editor)
	if bin == "" {
		return errors.New("empty editor command")
	}
	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("editor not found: %s", bin)
	}
	return nil
}

func editorBinary(editor string) string {
	editor = strings.TrimSpace(editor)
	if editor == "" {
		return ""
	}
	if strings.HasPrefix(editor, "\"") {
		if end := strings.Index(editor[1:], "\""); end >= 0 {
			return editor[1 : 1+end]
		}
	}
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func printEditorError(editor string, err error) {
	msg := fmt.Sprintf("Editor not found: %s. Set EDITOR or choose another editor.", editor)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "empty") {
		msg = "Editor command is empty. Set EDITOR or choose another editor."
	}
	printWarning(msg)
}

func printEditorRunError(editor string, err error) {
	msg := fmt.Sprintf("Failed to open editor: %s.", editor)
	if err != nil {
		msg = fmt.Sprintf("Failed to open editor: %s. %v", editor, err)
	}
	printError(msg)
}

func promptEditor() (string, error) {
	options := editorOptionsByOS()

	var selected string
	prompt := &survey.Select{
		Message: "Choose the editor to open the config file:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return "", err
	}

	switch selected {
	case "VS Code (code)":
		return "code", nil
	case "Notepad (notepad)":
		return "notepad", nil
	case "Notepad++ (notepad++)":
		return "notepad++", nil
	case "Sublime (subl)":
		return "subl", nil
	case "Nano (nano)":
		return "nano", nil
	case "Vim (vim)":
		return "vim", nil
	case "Vi (vi)":
		return "vi", nil
	case "Neovim (nvim)":
		return "nvim", nil
	case "Micro (micro)":
		return "micro", nil
	case "Emacs (emacs)":
		return "emacs", nil
	case "Joe (joe)":
		return "joe", nil
	default:
		var custom string
		input := &survey.Input{
			Message: "Enter the editor command:",
		}
		if err := survey.AskOne(input, &custom); err != nil {
			return "", err
		}
		return strings.TrimSpace(custom), nil
	}
}

func editorOptionsByOS() []string {
	common := []string{
		"VS Code (code)",
		"Sublime (subl)",
	}

	if runtime.GOOS == "windows" {
		return append([]string{
			"Notepad (notepad)",
			"Notepad++ (notepad++)",
		}, append(common, "Other")...)
	}

	unixEditors := []string{
		"Nano (nano)",
		"Vim (vim)",
		"Vi (vi)",
		"Neovim (nvim)",
		"Micro (micro)",
		"Emacs (emacs)",
		"Joe (joe)",
	}

	options := append(common, unixEditors...)
	options = append(options, "Other")
	return options
}
