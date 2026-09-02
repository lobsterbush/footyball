package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Version is the single source of truth for footyball's version string,
// shared with main.go's --version flag so the two can't drift apart.
const Version = "v0.1.0"

// View renders the full frame for the current model state.
func (m Model) View() string {
	if m.quitting {
		return ""
	}
	st := m.styles()
	width := m.width
	if width <= 0 {
		width = 100
	}

	header := m.viewHeader(st, width)

	var body string
	switch m.view {
	case viewDashboard:
		body = m.viewDashboardBody()
	case viewDetail:
		body = m.viewDetailBody()
	case viewStandings:
		body = m.viewStandingsBody()
	case viewSchedule:
		body = m.viewScheduleBody()
	case viewSettings:
		body = m.viewSettingsBody()
	}

	footer := m.viewFooter(st, width)

	sections := []string{header, lipgloss.NewStyle().Padding(1, 2).Render(body)}
	if critter := m.viewCritter(); critter != "" {
		sections = append(sections, critter)
	}
	sections = append(sections, footer)

	frame := lipgloss.JoinVertical(lipgloss.Left, sections...)
	out := st.App.Width(width).Render(frame)
	if m.bellPending {
		out = "\a" + out
	}
	return out
}

func (m Model) viewHeader(st styleSet, width int) string {
	cross := st.Gold.Render("✦ ✦ ✧ ✦")
	brand := st.Brand.Render("footyball")
	ver := st.Version.Render(" " + Version + " ")
	left := cross + "  " + brand + ver

	var right string
	switch m.view {
	case viewDashboard:
		right = fmt.Sprintf("filter: %s   theme: %s", m.filter.label(), m.th.Name)
	default:
		right = m.th.Name
	}
	rightStyled := st.Version.Render(right)

	gap := width - lipgloss.Width(left) - lipgloss.Width(rightStyled) - 2
	if gap < 1 {
		gap = 1
	}
	line := left + strings.Repeat(" ", gap) + rightStyled
	return st.Header.Width(width).Render(line)
}

func (m Model) viewFooter(st styleSet, width int) string {
	hints := m.footerHints()
	line := hints
	if m.status != "" {
		line = m.status + "   " + hints
	}
	return st.Footer.Width(width).Render(truncate(line, width-2))
}

func (m Model) footerHints() string {
	switch m.view {
	case viewDetail:
		return "↑/↓ scroll · f/F favorite · g/G schedule · esc back · t theme · q quit"
	case viewStandings:
		return "↑/↓ move · tab league · enter schedule · f favorite · esc back · t theme · q quit"
	case viewSchedule:
		return "↑/↓ scroll · f favorite · esc back · t theme · q quit"
	case viewSettings:
		return "↑/↓ move · space show/hide · K/J reorder · 0 reset · esc done"
	default:
		return "↑↓←→ move · enter open · v/V filter · f/F fav · g/G schedule · s standings · L leagues · r refresh · t theme · q quit"
	}
}
