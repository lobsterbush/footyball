package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateSettings(key string) (tea.Model, tea.Cmd) {
	order := m.cfg.OrderedLeagues()

	switch key {
	case "up", "k":
		if m.settingsCursor > 0 {
			m.settingsCursor--
		}
	case "down", "j":
		if m.settingsCursor < len(order)-1 {
			m.settingsCursor++
		}
	case " ":
		if m.settingsCursor >= 0 && m.settingsCursor < len(order) {
			l := order[m.settingsCursor]
			m.cfg.SetHidden(l.Key, !m.cfg.IsHidden(l.Key))
			_ = m.cfg.Save()
		}
	case "K":
		m.settingsCursor = m.cfg.MoveLeague(m.settingsCursor, -1)
		_ = m.cfg.Save()
	case "J":
		m.settingsCursor = m.cfg.MoveLeague(m.settingsCursor, 1)
		_ = m.cfg.Save()
	case "0":
		m.cfg.ResetLeagueOrder()
		m.settingsCursor = 0
		_ = m.cfg.Save()
	case "esc":
		newList := m.cfg.VisibleLeagues()
		var cmds []tea.Cmd
		for _, l := range newList {
			if _, ok := m.scoreboards[l.Key]; !ok {
				m.loading[l.Key] = true
				cmds = append(cmds, fetchScoreboardCmd(l))
			}
		}
		m.leagueList = newList
		if m.focusIdx >= len(m.leagueList) {
			m.focusIdx = 0
		}
		m.view = viewDashboard
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m Model) viewSettingsBody() string {
	st := m.styles()
	var lines []string
	lines = append(lines, st.LeagueName.Render("League settings"))
	lines = append(lines, st.Muted.Render("space: show/hide   K/J: reorder   0: reset   esc: done"))
	lines = append(lines, "")

	for i, l := range m.cfg.OrderedLeagues() {
		box := "[x]"
		style := st.Text
		if m.cfg.IsHidden(l.Key) {
			box = "[ ]"
			style = st.Muted
		}
		cursor := " "
		if i == m.settingsCursor {
			cursor = "▸"
			style = style.Bold(true).Foreground(m.th.Gold)
		}
		line := fmt.Sprintf("%s %s %-12s %s", cursor, box, l.Name, l.FullName)
		lines = append(lines, style.Render(line))
	}

	return strings.Join(lines, "\n")
}
