package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

const commandTemplate = `package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var {{.VarName}}Cmd = &cobra.Command{
	Use:   "{{.CommandName}}",
	Short: "Short description of the '{{.CommandName}}' command.",
	Long: ` + "`" + `Detailed description for the '{{.CommandName}}' command.

This command allows you to perform specific operations related to '{{.CommandName}}'.
Use --help to see all available options.` + "`" + `,
	Example: ` + "`" + `  {{.BinaryName}} {{.CommandName}}
  {{.BinaryName}} {{.CommandName}} --flag value` + "`" + `,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement logic for the '{{.CommandName}}' command
		fmt.Printf("Running command: %s\n", "{{.CommandName}}")

		// Example usage of flags
		// value, _ := cmd.Flags().GetString("flag-name")
		// fmt.Printf("Flag value: %s\n", value)
	},
}

func init() {
	rootCmd.AddCommand({{.VarName}}Cmd)

	// Add command-specific flags here
	// {{.VarName}}Cmd.Flags().StringP("flag-name", "f", "default", "Description of the flag")
	// {{.VarName}}Cmd.Flags().BoolP("verbose", "v", false, "Enable verbose mode")
}
`

type CommandData struct {
	CommandName string
	VarName     string
	BinaryName  string
}

// Converts a string to camelCase
func toCamelCase(input string) string {
	words := strings.FieldsFunc(input, func(c rune) bool {
		return c == '-' || c == '_' || unicode.IsSpace(c)
	})

	if len(words) == 0 {
		return "unknown"
	}

	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		if len(words[i]) > 0 {
			result += strings.Title(strings.ToLower(words[i]))
		}
	}

	return result
}

// Validates the command name (letters, numbers, hyphens only)
func isValidCommandName(name string) bool {
	if len(name) == 0 {
		return false
	}

	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' {
			return false
		}
	}

	return true
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run gen-command.go <command-name>")
		fmt.Println("Example: go run gen-command.go user-list")
		os.Exit(1)
	}

	commandName := strings.ToLower(strings.TrimSpace(os.Args[1]))

	// Validate command name
	if !isValidCommandName(commandName) {
		fmt.Printf("Invalid command name: %s\n", commandName)
		fmt.Println("Use only letters, numbers, and hyphens. Example: user-list, create-project")
		os.Exit(1)
	}

	varName := toCamelCase(commandName)
	binaryName := "dck" // Change this to match your CLI binary name

	data := CommandData{
		CommandName: commandName,
		VarName:     varName,
		BinaryName:  binaryName,
	}

	cmdDir := "cmd"
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		fmt.Printf("Failed to create directory %s: %v\n", cmdDir, err)
		os.Exit(1)
	}

	fileName := filepath.Join(cmdDir, commandName+".go")

	if _, err := os.Stat(fileName); err == nil {
		fmt.Printf("❌ File already exists: %s\n", fileName)
		fmt.Println("Delete the existing file or use a different command name.")
		os.Exit(1)
	}

	f, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	tmpl, err := template.New("command").Parse(commandTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template: %v\n", err)
		os.Exit(1)
	}

	if err := tmpl.Execute(f, data); err != nil {
		fmt.Printf("Failed to execute template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Command generated successfully: %s\n", fileName)
	fmt.Printf("📝 Variable name: %s\n", varName)
	fmt.Println("🔧 Next steps:")
	fmt.Printf("   1. Edit the file %s to implement the logic.\n", fileName)
	fmt.Println("   2. Run 'make build' to compile.")
	fmt.Printf("   3. Test it using './bin/%s %s'\n", binaryName, commandName)
}
