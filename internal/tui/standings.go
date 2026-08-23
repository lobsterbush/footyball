package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/crabtree/footyball/internal/api"
)

func (m Model) flattenStandings() []api.StandingsEntry {
	var out []api.StandingsEntry
	for _, g := range m.standingsGroups {
		out = append(out, g.Entries...)
	}
	return out
}

func (m Model) updateStandings(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.standingsCursor > 0 {
			m.standingsCursor--
		}
	case "down", "j":
		if m.standingsCursor < len(m.flattenStandings())-1 {
			m.standingsCursor++
		}
	case "tab", "shift+tab":
		if len(m.leagueList) == 0 {
			return m, nil
		}
		if key == "tab" {
			m.standingsIdx = (m.standingsIdx + 1) % len(m.leagueList)
		} else {
			m.standingsIdx = (m.standingsIdx - 1 + len(m.leagueList)) % len(m.leagueList)
		}
		m.standingsGroups = nil
		m.standingsErr = nil
		m.standingsLoad = true
		m.standingsCursor = 0
		return m, fetchStandingsCmd(m.leagueList[m.standingsIdx])
	case "enter":
		entries := m.flattenStandings()
		if m.standingsCursor >= 0 && m.standingsCursor < len(entries) {
			team := entries[m.standingsCursor].Team
			m.scheduleLeague = m.leagueList[m.standingsIdx]
			m.scheduleTeam = team
			m.scheduleEvents = nil
			m.scheduleErr = nil
			m.scheduleLoad = true
			m.scheduleScroll = 0
			m.prevView = viewStandings
			m.view = viewSchedule
			return m, fetchScheduleCmd(m.leagueList[m.standingsIdx], team)
		}
	case "f":
		entries := m.flattenStandings()
		if m.standingsCursor >= 0 && m.standingsCursor < len(entries) {
			l := m.leagueList[m.standingsIdx]
			team := entries[m.standingsCursor].Team
			m.cfg.ToggleFavorite(l.Key, team.ID)
			_ = m.cfg.Save()
			return m, m.setStatus("Updated favorites")
		}
	case "esc", "s":
		m.view = viewDashboard
	}
	return m, nil
}

func (m Model) viewStandingsBody() string {
	st := m.styles()
	if len(m.leagueList) == 0 {
		return st.Muted.Render("No leagues visible.")
	}
	l := m.leagueList[m.standingsIdx]
	var lines []string
	lines = append(lines, st.LeagueName.Render(l.FullName+" standings"))
	lines = append(lines, st.Muted.Render("tab/shift+tab: switch league   enter: team schedule   f: favorite"))
	lines = append(lines, "")

	if m.standingsLoad {
		lines = append(lines, st.Muted.Render("Loading…"))
		return strings.Join(lines, "\n")
	}
	if m.standingsErr != nil {
		lines = append(lines, st.Err.Render(m.standingsErr.Error()))
		return strings.Join(lines, "\n")
	}

	cursor := 0
	for _, g := range m.standingsGroups {
		if len(m.standingsGroups) > 1 {
			lines = append(lines, st.Gold.Bold(true).Render(g.Name))
		}

		cols := activeStandingsColumns(g.Entries)

		headerRow := fmt.Sprintf("%-3s %-24s", "#", "TEAM")
		for _, c := range cols {
			headerRow += " " + padLeft(c.header, c.width)
		}
		lines = append(lines, st.Muted.Render(headerRow))

		for _, e := range g.Entries {
			rank, _ := e.Stat("rank", "playoffSeed")

			star := ""
			if m.cfg.IsFavorite(l.Key, e.Team.ID) {
				star = "★"
			}
			name := star + e.Team.DisplayName

			row := fmt.Sprintf("%-3s %-24s", rank, truncate(name, 24))
			for _, c := range cols {
				v, _ := e.Stat(c.candidates...)
				row += " " + padLeft(v, c.width)
			}

			style := st.Text
			if cursor == m.standingsCursor {
				style = st.Text.Bold(true).Foreground(m.th.Gold)
				row = "▸" + row
			} else {
				row = " " + row
			}
			lines = append(lines, style.Render(row))
			cursor++
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

type standingsColumn struct {
	header     string
	width      int
	candidates []string
}

var standingsColumnSpecs = []standingsColumn{
	{"GP", 3, []string{"gamesPlayed"}},
	{"W", 3, []string{"gamesWon", "wins"}},
	{"D", 3, []string{"gamesDrawn", "ties"}},
	{"L", 3, []string{"gamesLost", "losses"}},
	{"PCT", 5, []string{"winPercent"}},
	{"PTS", 5, []string{"points"}},
	{"DIFF", 6, []string{"pointsDifference", "pointDifferential", "differential"}},
}

// activeStandingsColumns picks only the columns this league's entries
// actually populate — ESPN uses different stat sets per sport (e.g.
// basketball has no draws and no ladder "points", so it falls back to W/L
// win percentage instead).
func activeStandingsColumns(entries []api.StandingsEntry) []standingsColumn {
	if len(entries) == 0 {
		return nil
	}
	var cols []standingsColumn
	for _, c := range standingsColumnSpecs {
		if _, ok := entries[0].Stat(c.candidates...); ok {
			cols = append(cols, c)
		}
	}
	return cols
}
