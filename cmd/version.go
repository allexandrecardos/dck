package cmd

import "fmt"

var (
	// These values can be injected at build time using -ldflags
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const repoSlug = "allexandrecardos/dck"

func buildVersion() string {
	return version
}

func formatVersionInfo() string {
	return fmt.Sprintf("dck %s (commit %s, built %s)", version, commit, date)
}
