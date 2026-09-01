package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/crabtree/footyball/internal/api"
	"github.com/crabtree/footyball/internal/leagues"
)

func (m Model) currentLeague() (leagues.League, bool) {
	if m.focusIdx < 0 || m.focusIdx >= len(m.leagueList) {
		return leagues.League{}, false
	}
	return m.leagueList[m.focusIdx], true
}

func (m Model) selectedEvent() (leagues.League, api.Event, bool) {
	l, ok := m.currentLeague()
	if !ok {
		return l, api.Event{}, false
	}
	events := m.leagueEvents(l)
	idx := m.cursor[l.Key]
	if idx < 0 || idx >= len(events) {
		return l, api.Event{}, false
	}
	return l, events[idx], true
}

// indexOfEvent finds an event by ID in a (possibly just re-sorted) list,
// so the cursor can keep following the same game across a favorite toggle
// instead of drifting onto whatever ends up at the old index.
func indexOfEvent(events []api.Event, id string) int {
	for i, e := range events {
		if e.ID == id {
			return i
		}
	}
	return 0
}

func (m *Model) resetCursors() {
	for k := range m.cursor {
		m.cursor[k] = 0
	}
}

func (m Model) updateDashboard(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.focusIdx > 0 {
			m.focusIdx--
		}
	case "down", "j":
		if m.focusIdx < len(m.leagueList)-1 {
			m.focusIdx++
		}
	case "tab":
		if len(m.leagueList) > 0 {
			m.focusIdx = (m.focusIdx + 1) % len(m.leagueList)
		}
	case "shift+tab":
		if len(m.leagueList) > 0 {
			m.focusIdx = (m.focusIdx - 1 + len(m.leagueList)) % len(m.leagueList)
		}
	case "left", "h":
		if l, ok := m.currentLeague(); ok {
			if c := m.cursor[l.Key]; c > 0 {
				m.cursor[l.Key] = c - 1
			}
		}
	case "right":
		if l, ok := m.currentLeague(); ok {
			events := m.leagueEvents(l)
			if c := m.cursor[l.Key]; c < len(events)-1 {
				m.cursor[l.Key] = c + 1
			}
		}
	case "enter", "l":
		if l, ev, ok := m.selectedEvent(); ok {
			m.detailLeague = l
			m.detailEvent = ev
			m.summary = nil
			m.summaryErr = nil
			m.summaryLoad = true
			m.detailScroll = 0
			m.view = viewDetail
			return m, fetchSummaryCmd(l, ev.ID)
		}
	case "v":
		m.filter = stateFilter((int(m.filter) + 1) % 4)
		m.resetCursors()
	case "V":
		m.filter = stateFilter((int(m.filter) + 3) % 4)
		m.resetCursors()
	case "f", "F":
		if l, ev, ok := m.selectedEvent(); ok {
			comp, _ := ev.Competition()
			var team api.Competitor
			var found bool
			if key == "f" {
				team, found = comp.Away()
			} else {
				team, found = comp.Home()
			}
			if found {
				m.cfg.ToggleFavorite(l.Key, team.Team.ID)
				_ = m.cfg.Save()
				verb := "Favorited"
				if !m.cfg.IsFavorite(l.Key, team.Team.ID) {
					verb = "Unfavorited"
				}
				m.cursor[l.Key] = indexOfEvent(m.leagueEvents(l), ev.ID)
				cmd := m.setStatus(fmt.Sprintf("%s %s", verb, team.Team.DisplayName))
				return m, cmd
			}
		}
	case "g", "G":
		if l, ev, ok := m.selectedEvent(); ok {
			comp, _ := ev.Competition()
			var team api.Competitor
			var found bool
			if key == "g" {
				team, found = comp.Away()
			} else {
				team, found = comp.Home()
			}
			if found {
				m.scheduleLeague = l
				m.scheduleTeam = team.Team
				m.scheduleEvents = nil
				m.scheduleErr = nil
				m.scheduleLoad = true
				m.scheduleScroll = 0
				m.prevView = viewDashboard
				m.view = viewSchedule
				return m, fetchScheduleCmd(l, team.Team)
			}
		}
	case "s":
		if l, ok := m.currentLeague(); ok {
			m.standingsIdx = m.focusIdx
			m.standingsGroups = nil
			m.standingsErr = nil
			m.standingsLoad = true
			m.standingsCursor = 0
			m.view = viewStandings
			return m, fetchStandingsCmd(l)
		}
	case "L":
		m.settingsCursor = 0
		m.view = viewSettings
	case "r":
		cmds := fetchAllCmd(m.leagueList)
		for _, l := range m.leagueList {
			m.loading[l.Key] = true
		}
		cmds = append(cmds, m.setStatus("Refreshing…"))
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

const cardWidth = 30

func formatKickoff(t time.Time) string {
	return t.Local().Format("Mon 3:04pm")
}

// periodLabel is the short period-name prefix scoring plays and the
// in-progress status fallback render (e.g. "Q3", "H2"). ESPN's period
// number is sport-agnostic (just 1, 2, 3, ...), but what it means differs:
// AFL and the NBL play quarters, while NRL, Super Rugby, and soccer play
// halves.
func periodLabel(sportSlug string) string {
	switch sportSlug {
	case "rugby-league", "rugby", "soccer":
		return "H"
	default:
		return "Q"
	}
}

func statusLabel(s api.Status, pulseOn bool, sportSlug string, st styleSet) string {
	switch s.Type.State {
	case "in":
		dot := "●"
		if !pulseOn {
			dot = "○"
		}
		clock := s.DisplayClock
		if clock == "" {
			clock = fmt.Sprintf("%s%d", periodLabel(sportSlug), s.Period)
		}
		return st.Live.Render(dot + " LIVE " + clock)
	case "post":
		return st.Final.Render("FULL TIME")
	default:
		return st.Upcoming.Render(s.Type.ShortDetail)
	}
}

func (m Model) renderCard(l leagues.League, e api.Event, focused bool) string {
	st := m.styles()
	comp, ok := e.Competition()
	if !ok {
		return ""
	}
	away, _ := comp.Away()
	home, _ := comp.Home()

	homeColor := teamColor(home.Team.Color, m.th.Accent)

	star := func(teamID string) string {
		if m.cfg.IsFavorite(l.Key, teamID) {
			return st.Star.Render("★ ")
		}
		return ""
	}

	scoreOf := func(c api.Competitor) string {
		if comp.Status.Type.State == "pre" {
			return ""
		}
		return string(c.Score)
	}

	awayLine := fmt.Sprintf("%s%-4s %s", star(away.Team.ID), truncate(away.Team.Abbreviation, 4), padLeft(scoreOf(away), 3))
	homeLine := fmt.Sprintf("%s%-4s %s", star(home.Team.ID), truncate(home.Team.Abbreviation, 4), padLeft(scoreOf(home), 3))
	if away.Winner {
		awayLine = st.Text.Bold(true).Render(awayLine)
	} else {
		awayLine = st.Text.Render(awayLine)
	}
	if home.Winner {
		homeLine = st.Text.Bold(true).Render(homeLine)
	} else {
		homeLine = st.Text.Render(homeLine)
	}

	status := statusLabel(comp.Status, m.pulseOn, l.SportSlug, st)
	if comp.Status.Type.State == "pre" {
		status = st.Upcoming.Render(formatKickoff(e.Date.Time))
	}

	body := lipgloss.JoinVertical(lipgloss.Left, status, awayLine, homeLine)

	box := lipgloss.NewStyle().
		Width(cardWidth-2).
		Background(m.th.Surface).
		Foreground(m.th.Text).
		Padding(0, 1).
		BorderStyle(lipgloss.Border{Left: "▌"}).
		BorderLeft(true).
		BorderForeground(homeColor).
		BorderBackground(m.th.Surface)

	if focused {
		box = box.Bold(true).BorderForeground(m.th.Gold)
	}

	return box.Render(body)
}

func (m Model) renderLeagueSection(idx int, l leagues.League) string {
	st := m.styles()
	focusedLeague := idx == m.focusIdx

	mark := st.LeagueMark.Render(l.Mark)
	name := st.LeagueName.Render(l.Name)
	if focusedLeague {
		name = st.LeagueName.Foreground(m.th.Gold).Render("▸ " + l.Name)
	}
	header := lipgloss.JoinHorizontal(lipgloss.Center, mark, " ", name)

	if m.loading[l.Key] && m.scoreboards[l.Key] == nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, st.Muted.Render("  loading…"))
	}
	if err, ok := m.sbErr[l.Key]; ok && m.scoreboards[l.Key] == nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, st.Err.Render("  "+err.Error()))
	}

	events := m.leagueEvents(l)
	if len(events) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, st.Muted.Render("  no "+m.filter.label()+" games"))
	}

	perRow := m.width / (cardWidth + 1)
	if perRow < 1 {
		perRow = 1
	}
	cursor := m.cursor[l.Key]

	var rows []string
	var row []string
	for i, e := range events {
		focusedCard := focusedLeague && i == cursor
		row = append(row, m.renderCard(l, e, focusedCard))
		if len(row) == perRow || i == len(events)-1 {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, row...))
			row = nil
		}
	}
	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func (m Model) viewDashboardBody() string {
	var sections []string
	for i, l := range m.leagueList {
		sections = append(sections, m.renderLeagueSection(i, l))
	}
	if len(sections) == 0 {
		return m.styles().Muted.Render("All leagues hidden — press L to bring one back.")
	}
	return strings.Join(sections, "\n\n")
}
