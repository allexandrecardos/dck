package formatter

import (
	"fmt"
	"time"
)

func FormatSince(t time.Time) string {
	if t.IsZero() {
		return "-"
	}

	d := time.Since(t)
	if d < time.Minute {
		return "now"
	}

	if d < time.Hour {
		return formatUnit(d, time.Minute, "m")
	}

	if d < 24*time.Hour {
		return formatUnit(d, time.Hour, "h")
	}

	return formatUnit(d, 24*time.Hour, "d")
}

func formatUnit(d, unit time.Duration, suffix string) string {
	n := int(d / unit)
	if n < 0 {
		n = 0
	}
	return fmt.Sprintf("%d%s", n, suffix)
}
