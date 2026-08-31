// Command footyball is a terminal dashboard for Australian sport: AFL, NRL,
// A-League Men, A-League Women, NBL, and Super Rugby Pacific, drawn live
// from ESPN's public scoreboard API. No API key required.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/crabtree/footyball/internal/config"
	"github.com/crabtree/footyball/internal/tui"
)

const version = "0.1.0"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("footyball " + version)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "footyball: couldn't load config:", err)
		cfg = config.Default()
	}

	termDark := lipgloss.HasDarkBackground()
	m := tui.New(cfg, termDark)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "footyball:", err)
		os.Exit(1)
	}
}
