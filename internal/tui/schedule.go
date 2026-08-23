package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/crabtree/footyball/internal/api"
)

func (m Model) updateSchedule(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.scheduleScroll > 0 {
			m.scheduleScroll--
		}
	case "down", "j":
		m.scheduleScroll++
	case "f", "F":
		m.cfg.ToggleFavorite(m.scheduleLeague.Key, m.scheduleTeam.ID)
		_ = m.cfg.Save()
		return m, m.setStatus("Updated favorites")
	case "esc", "h":
		m.view = m.prevView
	}
	return m, nil
}

func (m Model) viewScheduleBody() string {
	st := m.styles()
	var lines []string
	star := ""
	if m.cfg.IsFavorite(m.scheduleLeague.Key, m.scheduleTeam.ID) {
		star = "★ "
	}
	lines = append(lines, st.LeagueName.Render(star+m.scheduleTeam.DisplayName+" — "+m.scheduleLeague.FullName))
	lines = append(lines, st.Muted.Render("f: favorite this team   esc: back"))
	lines = append(lines, "")

	if m.scheduleLoad {
		lines = append(lines, st.Muted.Render("Loading…"))
		return strings.Join(lines, "\n")
	}
	if m.scheduleErr != nil {
		lines = append(lines, st.Err.Render(m.scheduleErr.Error()))
		return strings.Join(lines, "\n")
	}
	if len(m.scheduleEvents) == 0 {
		lines = append(lines, st.Muted.Render("No fixtures found."))
		return strings.Join(lines, "\n")
	}

	start := m.scheduleScroll
	if start > len(m.scheduleEvents)-1 {
		start = len(m.scheduleEvents) - 1
	}
	if start < 0 {
		start = 0
	}
	end := start + 18
	if end > len(m.scheduleEvents) {
		end = len(m.scheduleEvents)
	}

	for _, e := range m.scheduleEvents[start:end] {
		comp, ok := e.Competition()
		if !ok {
			continue
		}
		var opp api.Competitor
		var mine api.Competitor
		home, _ := comp.Home()
		away, _ := comp.Away()
		if home.Team.ID == m.scheduleTeam.ID {
			mine, opp = home, away
		} else {
			mine, opp = away, home
		}
		versus := "vs"
		if mine.HomeAway == "away" {
			versus = "@"
		}

		var result string
		switch comp.Status.Type.State {
		case "pre":
			result = st.Upcoming.Render(formatKickoff(e.Date.Time))
		case "in":
			result = statusLabel(comp.Status, m.pulseOn, st)
		default:
			outcome := "L"
			style := st.Err
			if mine.Winner {
				outcome = "W"
				style = st.Live.Foreground(m.th.Win)
			} else if mine.Score == opp.Score {
				outcome = "D"
				style = st.Muted
			}
			result = style.Render(fmt.Sprintf("%s %s-%s", outcome, mine.Score, opp.Score))
		}

		line := fmt.Sprintf("%-3s %-24s %s", versus, truncate(opp.Team.DisplayName, 24), result)
		lines = append(lines, st.Text.Render(line))
	}
	if start > 0 || end < len(m.scheduleEvents) {
		lines = append(lines, st.Muted.Render(fmt.Sprintf("  (%d-%d of %d — ↑/↓ to scroll)", start+1, end, len(m.scheduleEvents))))
	}

	return strings.Join(lines, "\n")
}
