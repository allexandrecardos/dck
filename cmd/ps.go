package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/allexandrecardos/dck/internal/config"
	"github.com/allexandrecardos/dck/internal/docker"
	"github.com/allexandrecardos/dck/internal/formatter"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
)

var psAll bool

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List Docker containers",
	Long:  "List containers in a clean table. Columns can be customized in dck-config.yml.",
	Example: "  dck ps\n" +
		"  dck ps -a",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cli, err := docker.New()
		if err != nil {
			return err
		}

		if err := docker.Ping(ctx, cli); err != nil {
			return fmt.Errorf("docker unavailable: %w", err)
		}

		cfgPath := resolveConfigPath()
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}

		columns := resolveColumns(cfg)

		containers, err := docker.ListContainers(ctx, cli, psAll, needsSize(columns))
		if err != nil {
			return err
		}

		stats := map[string]docker.StatsSnapshot{}
		if needsStats(columns) && len(containers) > 0 {
			stats = collectStats(ctx, cli, containers)
		}

		printContainers(containers, columns, stats)
		return nil
	},
}

func init() {
	psCmd.Flags().BoolVarP(&psAll, "all", "a", false, "Show all containers")
	rootCmd.AddCommand(psCmd)
}

type columnDef struct {
	key   string
	title string
}

type psRow struct {
	id      string
	name    string
	image   string
	status  string
	ports   string
	created string
	cpu     string
	mem     string
	network string
	size    string
	command string
}

func printContainers(containers []types.Container, columns []columnDef, stats map[string]docker.StatsSnapshot) {
	r := renderer.NewBlueprint(tw.Rendition{
		Borders: tw.Border{Left: tw.On, Right: tw.On, Top: tw.On, Bottom: tw.On},
		Symbols: tw.NewSymbols(tw.StyleRounded),
		Settings: tw.Settings{
			Separators: tw.Separators{BetweenColumns: tw.On},
			Lines: tw.Lines{
				ShowTop:        tw.On,
				ShowBottom:     tw.On,
				ShowHeaderLine: tw.On,
			},
		},
	})

	rowPad := tw.Padding{
		Left:      " ",
		Right:     " ",
		Top:       " ",
		Bottom:    "",
		Overwrite: true,
	}

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithRenderer(r),
		tablewriter.WithConfig(tablewriter.Config{
			Behavior: tw.Behavior{TrimSpace: tw.On},
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignCenter},
				Padding:   tw.CellPadding{Global: tw.PaddingDefault},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignLeft},
				Padding:   tw.CellPadding{Global: rowPad},
			},
		}),
	)

	header := make([]string, 0, len(columns))
	for _, col := range columns {
		header = append(header, col.title)
	}
	table.Header(header)

	for _, c := range containers {
		row := buildRow(c, stats[c.ID])
		_ = table.Append(projectRow(row, columns))
	}

	if len(containers) > 0 {
		_ = table.Append(make([]string, len(columns)))
	}

	_ = table.Render()
}

func buildRow(c types.Container, stat docker.StatsSnapshot) psRow {
	id := c.ID
	if len(id) > 12 {
		id = id[:12]
	}

	name := strings.TrimPrefix(strings.Join(c.Names, ","), "/")
	ports := formatter.FormatPorts(c.Ports)
	status := formatStatus(c.Status)
	created := formatter.FormatSince(time.Unix(c.Created, 0))
	network := formatNetworks(c.NetworkSettings)
	size := formatSize(c)
	command := formatCommand(c.Command)
	cpu, mem := formatStats(stat)

	return psRow{
		id:      id,
		name:    name,
		image:   c.Image,
		status:  status,
		ports:   ports,
		created: created,
		cpu:     cpu,
		mem:     mem,
		network: network,
		size:    size,
		command: command,
	}
}

func projectRow(row psRow, columns []columnDef) []string {
	values := make([]string, 0, len(columns))
	for _, col := range columns {
		switch col.key {
		case "id":
			values = append(values, row.id)
		case "name":
			values = append(values, row.name)
		case "image":
			values = append(values, row.image)
		case "status":
			values = append(values, row.status)
		case "ports":
			values = append(values, row.ports)
		case "created":
			values = append(values, row.created)
		case "cpu":
			values = append(values, row.cpu)
		case "mem":
			values = append(values, row.mem)
		case "network":
			values = append(values, row.network)
		case "size":
			values = append(values, row.size)
		case "command":
			values = append(values, row.command)
		default:
			values = append(values, "")
		}
	}
	return values
}

func resolveColumns(cfg config.Config) []columnDef {
	defaults := []columnDef{
		{key: "id", title: "ID"},
		{key: "name", title: "NAME"},
		{key: "image", title: "IMAGE"},
		{key: "status", title: "STATUS"},
		{key: "ports", title: "PORTS"},
		{key: "command", title: "COMMAND"},
	}

	all := []columnDef{
		{key: "id", title: "ID"},
		{key: "name", title: "NAME"},
		{key: "image", title: "IMAGE"},
		{key: "status", title: "STATUS"},
		{key: "ports", title: "PORTS"},
		{key: "created", title: "CREATED"},
		{key: "cpu", title: "CPU%"},
		{key: "mem", title: "MEM"},
		{key: "network", title: "NETWORK"},
		{key: "size", title: "SIZE"},
		{key: "command", title: "COMMAND"},
	}

	if len(cfg.PS.Columns) == 0 {
		return defaults
	}

	lookup := map[string]columnDef{}
	for _, col := range all {
		lookup[col.key] = col
	}

	cols := make([]columnDef, 0, len(cfg.PS.Columns))
	for _, key := range cfg.PS.Columns {
		k := strings.ToLower(strings.TrimSpace(key))
		if def, ok := lookup[k]; ok {
			cols = append(cols, def)
		}
	}

	if len(cols) == 0 {
		return defaults
	}
	return cols
}

func resolveConfigPath() string {
	exe, err := os.Executable()
	if err == nil {
		baseDir := filepath.Dir(exe)
		path := filepath.Join(baseDir, "dck-config.yml")
		if _, statErr := os.Stat(path); statErr == nil {
			return path
		}
	}

	// fallback to current working directory
	return filepath.Join(".", "dck-config.yml")
}

func needsStats(columns []columnDef) bool {
	for _, col := range columns {
		if col.key == "cpu" || col.key == "mem" {
			return true
		}
	}
	return false
}

func needsSize(columns []columnDef) bool {
	for _, col := range columns {
		if col.key == "size" {
			return true
		}
	}
	return false
}

func collectStats(parent context.Context, cli *client.Client, containers []types.Container) map[string]docker.StatsSnapshot {
	stats := map[string]docker.StatsSnapshot{}
	for _, c := range containers {
		ctx, cancel := context.WithTimeout(parent, 800*time.Millisecond)
		snapshot, err := docker.StatsOnce(ctx, cli, c.ID)
		cancel()
		if err == nil {
			stats[c.ID] = snapshot
		}
	}
	return stats
}

func formatStats(stat docker.StatsSnapshot) (string, string) {
	if stat.MemLimit == 0 && stat.MemUsage == 0 && stat.CPUPercent == 0 {
		return "-", "-"
	}

	cpu := fmt.Sprintf("%.1f%%", stat.CPUPercent)
	mem := "-"
	if stat.MemLimit > 0 {
		mem = fmt.Sprintf("%s / %s", formatter.FormatBytes(stat.MemUsage), formatter.FormatBytes(stat.MemLimit))
	} else if stat.MemUsage > 0 {
		mem = formatter.FormatBytes(stat.MemUsage)
	}
	return cpu, mem
}

func formatNetworks(netSettings *types.SummaryNetworkSettings) string {
	if netSettings == nil || len(netSettings.Networks) == 0 {
		return "-"
	}

	names := make([]string, 0, len(netSettings.Networks))
	for name := range netSettings.Networks {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ",")
}

func formatSize(c types.Container) string {
	if c.SizeRw == 0 && c.SizeRootFs == 0 {
		return "-"
	}
	if c.SizeRootFs > 0 {
		return fmt.Sprintf("%s (%s)", formatter.FormatBytes(uint64(c.SizeRw)), formatter.FormatBytes(uint64(c.SizeRootFs)))
	}
	return formatter.FormatBytes(uint64(c.SizeRw))
}

func formatCommand(cmd string) string {
	if cmd == "" {
		return "-"
	}
	if len(cmd) > 40 {
		return cmd[:37] + "..."
	}
	return cmd
}

func formatStatus(status string) string {
	s := strings.ToLower(status)
	icon := "[ ]"
	color := colorGray
	label := "unknown"
	timeText := extractStatusTime(status)

	switch {
	case strings.Contains(s, "paused"):
		icon = "[~]"
		color = colorYellow
		label = "paused"
	case strings.Contains(s, "restarting"):
		icon = "[~]"
		color = colorYellow
		label = "restarting"
	case strings.Contains(s, "unhealthy"):
		icon = "[x]"
		color = colorRed
		label = "unhealthy"
	case strings.Contains(s, "healthy"):
		icon = "[+]"
		color = colorGreen
		label = "healthy"
	case strings.HasPrefix(s, "up"):
		icon = "[+]"
		color = colorGreen
		label = "running"
	case strings.Contains(s, "exited"):
		icon = "[x]"
		color = colorRed
		label = "exited"
	case strings.Contains(s, "created"):
		icon = "[~]"
		color = colorYellow
		label = "created"
	case strings.Contains(s, "dead"):
		icon = "[x]"
		color = colorRed
		label = "dead"
	default:
		icon = "[ ]"
		color = colorGray
		label = "unknown"
	}

	text := fmt.Sprintf("%s %s", icon, label)
	if timeText != "" {
		text = fmt.Sprintf("%s [%s]", text, timeText)
	}
	return colorize(text, color)
}

func extractStatusTime(status string) string {
	s := strings.TrimSpace(status)
	if s == "" {
		return ""
	}
	lower := strings.ToLower(s)

	if strings.HasPrefix(lower, "up ") {
		rest := strings.TrimSpace(s[3:])
		if idx := strings.Index(rest, "("); idx >= 0 {
			rest = strings.TrimSpace(rest[:idx])
		}
		return rest
	}

	if idx := strings.Index(lower, "exited"); idx >= 0 {
		rest := strings.TrimSpace(s[idx+len("exited"):])
		if paren := strings.Index(rest, ")"); paren >= 0 {
			rest = strings.TrimSpace(rest[paren+1:])
		}
		return rest
	}

	return ""
}

const (
	colorReset  = "\x1b[0m"
	colorRed    = "\x1b[31m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
	colorGray   = "\x1b[90m"
	colorBlue   = "\x1b[94m"
)

func colorize(text, color string) string {
	if os.Getenv("NO_COLOR") != "" {
		return text
	}
	return color + text + colorReset
}
