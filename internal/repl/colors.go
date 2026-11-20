package repl

import (
	"fmt"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

func hpBar(hp, maxHP int) string {
	if maxHP <= 0 {
		return "[----------] 0/0"
	}
	if hp < 0 {
		hp = 0
	}
	width := 10
	filled := (hp * width) / maxHP
	if filled > width {
		filled = width
	}

	ratio := float64(hp) / float64(maxHP)
	color := colorGreen
	if ratio <= 0.3 {
		color = colorRed
	} else if ratio <= 0.6 {
		color = colorYellow
	}

	bar := strings.Repeat("#", filled) + strings.Repeat("-", width-filled)
	return fmt.Sprintf("%s[%s] %d/%d HP%s", color, bar, hp, maxHP, colorReset)
}

func dangerBadge(level int) string {
	switch {
	case level <= 1:
		return fmt.Sprintf("%sDanger %d (Low)%s", colorGreen, level, colorReset)
	case level == 2:
		return fmt.Sprintf("%sDanger %d (Moderate)%s", colorYellow, level, colorReset)
	default:
		return fmt.Sprintf("%s%sDanger %d (Deadly)%s", colorBold, colorRed, level, colorReset)
	}
}
