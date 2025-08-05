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
	"github.com/spf13/cobra"
)

var {{.VarName}}Cmd = &cobra.Command{
	Use:   "{{.CommandName}}",
	Short: "Descrição curta do comando {{.CommandName}}",
	Long: ` + "`" + `Descrição mais detalhada do comando {{.CommandName}}.

Este comando permite realizar operações específicas relacionadas a {{.CommandName}}.
Use --help para ver todas as opções disponíveis.` + "`" + `,
	Example: ` + "`" + `  {{.BinaryName}} {{.CommandName}}
  {{.BinaryName}} {{.CommandName}} --flag valor` + "`" + `,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implementar lógica do comando {{.CommandName}}
		fmt.Printf("Executando comando: %s\n", "{{.CommandName}}")

		// Exemplo de uso de flags
		// value, _ := cmd.Flags().GetString("flag-name")
		// fmt.Printf("Flag value: %s\n", value)
	},
}

func init() {
	rootCmd.AddCommand({{.VarName}}Cmd)

	// Adicione flags específicas do comando aqui
	// {{.VarName}}Cmd.Flags().StringP("flag-name", "f", "default", "Descrição da flag")
	// {{.VarName}}Cmd.Flags().BoolP("verbose", "v", false, "Modo verboso")
}
`

type CommandData struct {
	CommandName string
	VarName     string
	BinaryName  string
}

// Converte string para camelCase correto
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

// Valida se o nome do comando é válido
func isValidCommandName(name string) bool {
	if len(name) == 0 {
		return false
	}

	// Permite apenas letras, números e hífens
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' {
			return false
		}
	}

	return true
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run gen-command.go <nome-do-comando>")
		fmt.Println("Exemplo: go run gen-command.go user-list")
		os.Exit(1)
	}

	commandName := strings.ToLower(strings.TrimSpace(os.Args[1]))

	// Validação do nome do comando
	if !isValidCommandName(commandName) {
		fmt.Printf("Nome do comando inválido: %s\n", commandName)
		fmt.Println("Use apenas letras, números e hífens. Exemplo: user-list, create-project")
		os.Exit(1)
	}

	varName := toCamelCase(commandName)
	binaryName := "dck" // Nome do binário principal

	data := CommandData{
		CommandName: commandName,
		VarName:     varName,
		BinaryName:  binaryName,
	}

	// Cria o diretório cmd se não existir
	cmdDir := "cmd"
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		fmt.Printf("Erro ao criar diretório %s: %v\n", cmdDir, err)
		os.Exit(1)
	}

	fileName := filepath.Join(cmdDir, commandName+".go")

	// Verifica se o arquivo já existe
	if _, err := os.Stat(fileName); err == nil {
		fmt.Printf("❌ Arquivo já existe: %s\n", fileName)
		fmt.Println("Remova o arquivo existente ou use um nome diferente.")
		os.Exit(1)
	}

	// Cria o arquivo
	f, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Erro ao criar o arquivo: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	// Parse do template
	tmpl, err := template.New("command").Parse(commandTemplate)
	if err != nil {
		fmt.Printf("Erro ao criar o template: %v\n", err)
		os.Exit(1)
	}

	// Executa o template
	if err := tmpl.Execute(f, data); err != nil {
		fmt.Printf("Erro ao executar o template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Comando gerado com sucesso: %s\n", fileName)
	fmt.Printf("📝 Nome da variável: %s\n", varName)
	fmt.Printf("🔧 Próximos passos:\n")
	fmt.Printf("   1. Edite o arquivo %s para implementar a lógica\n", fileName)
	fmt.Printf("   2. Execute 'make build' para compilar\n")
	fmt.Printf("   3. Teste com './bin/%s %s'\n", binaryName, commandName)
}
