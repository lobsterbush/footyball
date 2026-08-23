package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type styleSet struct {
	App        lipgloss.Style
	Header     lipgloss.Style
	Brand      lipgloss.Style
	Version    lipgloss.Style
	LeagueName lipgloss.Style
	LeagueMark lipgloss.Style
	Muted      lipgloss.Style
	Text       lipgloss.Style
	Card       lipgloss.Style
	CardFocus  lipgloss.Style
	Live       lipgloss.Style
	Final      lipgloss.Style
	Upcoming   lipgloss.Style
	Gold       lipgloss.Style
	Accent     lipgloss.Style
	Footer     lipgloss.Style
	Key        lipgloss.Style
	Star       lipgloss.Style
	Err        lipgloss.Style
}

func (m Model) styles() styleSet {
	t := m.th
	return styleSet{
		App: lipgloss.NewStyle().Background(t.BG).Foreground(t.Text),
		Header: lipgloss.NewStyle().Bold(true).Foreground(t.Text).
			Background(t.Surface).Padding(0, 1),
		Brand:      lipgloss.NewStyle().Bold(true).Foreground(t.Gold),
		Version:    lipgloss.NewStyle().Foreground(t.Muted),
		LeagueName: lipgloss.NewStyle().Bold(true).Foreground(t.Text),
		LeagueMark: lipgloss.NewStyle().Bold(true).Foreground(t.BG).Background(t.Accent).Padding(0, 1),
		Muted:      lipgloss.NewStyle().Foreground(t.Muted),
		Text:       lipgloss.NewStyle().Foreground(t.Text),
		Card:       lipgloss.NewStyle().Foreground(t.Text).Background(t.Surface).Padding(0, 1),
		CardFocus:  lipgloss.NewStyle().Foreground(t.Text).Background(t.Surface).Bold(true),
		Live:       lipgloss.NewStyle().Bold(true).Foreground(t.Live),
		Final:      lipgloss.NewStyle().Foreground(t.Muted),
		Upcoming:   lipgloss.NewStyle().Foreground(t.Accent),
		Gold:       lipgloss.NewStyle().Foreground(t.Gold),
		Accent:     lipgloss.NewStyle().Foreground(t.Accent),
		Footer:     lipgloss.NewStyle().Foreground(t.Muted).Background(t.Surface).Padding(0, 1),
		Key:        lipgloss.NewStyle().Bold(true).Foreground(t.Gold),
		Star:       lipgloss.NewStyle().Foreground(t.Gold).Bold(true),
		Err:        lipgloss.NewStyle().Foreground(t.Live).Bold(true),
	}
}

// teamColor resolves an ESPN team hex color against the theme, falling
// back to the theme's accent when ESPN gives us nothing usable (empty,
// pure black, or pure white all read as "no real color" in practice).
func teamColor(hex string, fallback lipgloss.Color) lipgloss.Color {
	hex = strings.TrimSpace(strings.ToLower(hex))
	if hex == "" || hex == "000000" || hex == "ffffff" {
		return fallback
	}
	return lipgloss.Color("#" + hex)
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n == 1 {
		return string(r[:1])
	}
	return string(r[:n-1]) + "…"
}

func padRight(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

func padLeft(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return strings.Repeat(" ", n-w) + s
}

func bar(pct float64, width int, fill, empty lipgloss.Style) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	filled := int(pct*float64(width) + 0.5)
	if filled > width {
		filled = width
	}
	return fill.Render(strings.Repeat("█", filled)) + empty.Render(strings.Repeat("░", width-filled))
}

func kickoffLabel(local string) string {
	return fmt.Sprintf(" %s ", local)
}
