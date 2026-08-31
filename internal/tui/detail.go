package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/crabtree/footyball/internal/api"
)

func (m Model) updateDetail(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.detailScroll > 0 {
			m.detailScroll--
		}
	case "down", "j":
		m.detailScroll++
	case "f", "F":
		comp, _ := m.detailEvent.Competition()
		var team api.Competitor
		var found bool
		if key == "f" {
			team, found = comp.Away()
		} else {
			team, found = comp.Home()
		}
		if found {
			m.cfg.ToggleFavorite(m.detailLeague.Key, team.Team.ID)
			_ = m.cfg.Save()
			return m, m.setStatus("Updated favorites")
		}
	case "g", "G":
		comp, _ := m.detailEvent.Competition()
		var team api.Competitor
		var found bool
		if key == "g" {
			team, found = comp.Away()
		} else {
			team, found = comp.Home()
		}
		if found {
			m.scheduleLeague = m.detailLeague
			m.scheduleTeam = team.Team
			m.scheduleEvents = nil
			m.scheduleErr = nil
			m.scheduleLoad = true
			m.scheduleScroll = 0
			m.prevView = viewDetail
			m.view = viewSchedule
			return m, fetchScheduleCmd(m.detailLeague, team.Team)
		}
	case "esc", "h":
		m.view = viewDashboard
	}
	return m, nil
}

// scoringPlays filters a play-by-play feed down to plays where either
// team's score changed, which reads as "scoring plays" across every sport
// without needing a per-sport list of play-type names.
func scoringPlays(plays []api.Play) []api.Play {
	var out []api.Play
	prevHome, prevAway := 0, 0
	for _, p := range plays {
		if p.HomeScore != prevHome || p.AwayScore != prevAway {
			out = append(out, p)
		}
		prevHome, prevAway = p.HomeScore, p.AwayScore
	}
	return out
}

// scoringKeyEvents filters a keyEvents feed down to non-shootout goals,
// pairing each with a running score computed by tallying one goal per team
// as they occur, since soccer's key-events feed carries no score field of
// its own, unlike the "plays" feed's awayScore/homeScore. Used as a
// fallback for sports (currently A-League Men/Women) whose /summary
// responses populate "keyEvents" instead of "plays".
func scoringKeyEvents(events []api.KeyEvent, awayTeamID, homeTeamID string) []api.Play {
	var out []api.Play
	awayScore, homeScore := 0, 0
	for _, e := range events {
		if e.Shootout || !e.ScoringPlay || e.Type.Text != "Goal" {
			continue
		}
		switch e.Team.ID {
		case awayTeamID:
			awayScore++
		case homeTeamID:
			homeScore++
		}
		out = append(out, api.Play{
			Text:      e.Text,
			AwayScore: awayScore,
			HomeScore: homeScore,
			Period:    e.Period,
			Clock:     e.Clock,
			Team:      e.Team,
		})
	}
	return out
}

// topPerformersLine renders each side's headline stat leader (top goal
// scorer, top scorer, etc.) as a single line. Not every sport/game has
// this data, so it's a graceful no-op when absent.
func topPerformersLine(leaders []api.TeamLeaders, away, home api.Competitor, st styleSet) (string, bool) {
	find := func(teamID string) (cat, player, value string, ok bool) {
		for _, t := range leaders {
			if t.Team.ID == teamID {
				return t.TopPerformer()
			}
		}
		return "", "", "", false
	}
	aCat, aPlayer, aValue, aOK := find(away.Team.ID)
	hCat, hPlayer, hValue, hOK := find(home.Team.ID)
	if !aOK && !hOK {
		return "", false
	}
	parts := []string{st.Gold.Bold(true).Render("TOP PERFORMERS")}
	if aOK {
		parts = append(parts, st.Text.Render(fmt.Sprintf("  %-4s %s — %s %s", away.Team.Abbreviation, aPlayer, aValue, aCat)))
	}
	if hOK {
		parts = append(parts, st.Text.Render(fmt.Sprintf("  %-4s %s — %s %s", home.Team.Abbreviation, hPlayer, hValue, hCat)))
	}
	return strings.Join(parts, "\n"), true
}

func findBoxTeam(bx api.Boxscore, teamID string) (api.BoxTeam, bool) {
	for _, t := range bx.Teams {
		if t.Team.ID == teamID {
			return t, true
		}
	}
	return api.BoxTeam{}, false
}

func statByName(stats []api.BoxTeamStat, name string) (api.BoxTeamStat, bool) {
	for _, s := range stats {
		if s.Name == name {
			return s, true
		}
	}
	return api.BoxTeamStat{}, false
}

func (m Model) viewDetailBody() string {
	st := m.styles()
	comp, ok := m.detailEvent.Competition()
	if !ok {
		return st.Err.Render("No competition data for this game.")
	}
	away, _ := comp.Away()
	home, _ := comp.Home()

	var lines []string

	title := fmt.Sprintf("%s  vs  %s", away.Team.DisplayName, home.Team.DisplayName)
	lines = append(lines, st.LeagueName.Render(title))
	subtitle := m.detailLeague.FullName
	if comp.Venue.FullName != "" {
		subtitle += " · " + comp.Venue.FullName
	}
	lines = append(lines, st.Muted.Render(subtitle))
	lines = append(lines, "")

	scoreLine := fmt.Sprintf("%-24s %3s", truncate(away.Team.DisplayName, 24), away.Score)
	lines = append(lines, st.Text.Render(scoreLine))
	scoreLine = fmt.Sprintf("%-24s %3s", truncate(home.Team.DisplayName, 24), home.Score)
	lines = append(lines, st.Text.Render(scoreLine))
	lines = append(lines, statusLabel(comp.Status, m.pulseOn, st))
	lines = append(lines, "")

	if m.summaryLoad {
		lines = append(lines, st.Muted.Render("Loading box score…"))
		return strings.Join(lines, "\n")
	}
	if m.summaryErr != nil {
		lines = append(lines, st.Err.Render("Couldn't load detail: "+m.summaryErr.Error()))
		return strings.Join(lines, "\n")
	}
	if m.summary == nil {
		lines = append(lines, st.Muted.Render("No detail available."))
		return strings.Join(lines, "\n")
	}

	// Top performers, one line per side, when ESPN provides them.
	if topLine, ok := topPerformersLine(m.summary.Leaders, away, home, st); ok {
		lines = append(lines, topLine, "")
	}

	// Box score comparison bars.
	awayBox, hasAway := findBoxTeam(m.summary.Boxscore, away.Team.ID)
	homeBox, hasHome := findBoxTeam(m.summary.Boxscore, home.Team.ID)
	awayStats := api.FlattenBoxStats(awayBox.Statistics)
	homeStats := api.FlattenBoxStats(homeBox.Statistics)
	if hasAway && hasHome && len(awayStats) > 0 {
		lines = append(lines, st.Gold.Bold(true).Render("TEAM STATS"))
		lines = append(lines, st.Muted.Render(padLeft(truncate(away.Team.Abbreviation, 4), 4)+strings.Repeat(" ", 22)+truncate(home.Team.Abbreviation, 4)))
		for _, as := range awayStats {
			if as.Label == "" {
				continue
			}
			hs, found := statByName(homeStats, as.Name)
			if !found {
				continue
			}
			av, aerr := strconv.ParseFloat(as.DisplayValue, 64)
			hv, herr := strconv.ParseFloat(hs.DisplayValue, 64)
			label := padRight(truncate(as.Label, 19), 20)
			if aerr == nil && herr == nil && (av > 0 || hv > 0) {
				total := av + hv
				aPct := 0.5
				if total > 0 {
					aPct = av / total
				}
				row := fmt.Sprintf("%3s %s%s %-3s",
					padLeft(as.DisplayValue, 3),
					bar(aPct, 10, st.Accent.Reverse(true), st.Muted),
					bar(1-aPct, 10, st.Gold.Reverse(true), st.Muted),
					hs.DisplayValue,
				)
				lines = append(lines, st.Text.Render(label)+row)
			} else {
				lines = append(lines, st.Text.Render(fmt.Sprintf("%s %6s   vs   %-6s", label, as.DisplayValue, hs.DisplayValue)))
			}
		}
		lines = append(lines, "")
	}

	// Scoring plays. AFL/NRL/rugby populate "plays" with a running score;
	// A-League soccer instead populates "keyEvents" with individual goal
	// events and no score field, so fall back to tallying those when
	// "plays" is empty.
	plays := scoringPlays(m.summary.Plays)
	if len(plays) == 0 {
		plays = scoringKeyEvents(m.summary.KeyEvents, away.Team.ID, home.Team.ID)
	}
	if len(plays) > 0 {
		lines = append(lines, st.Gold.Bold(true).Render("SCORING PLAYS"))
		start := m.detailScroll
		if start > len(plays)-1 {
			start = len(plays) - 1
		}
		if start < 0 {
			start = 0
		}
		end := start + 12
		if end > len(plays) {
			end = len(plays)
		}
		for _, p := range plays[start:end] {
			who := ""
			if p.Team.ID == away.Team.ID {
				who = away.Team.Abbreviation
			} else if p.Team.ID == home.Team.ID {
				who = home.Team.Abbreviation
			}
			line := fmt.Sprintf("Q%d %5s  %-4s %-3d-%-3d  %s",
				p.Period.Number, p.Clock.DisplayValue, who, p.AwayScore, p.HomeScore, truncate(p.Text, 40))
			lines = append(lines, st.Text.Render(line))
		}
		if len(plays) > end || start > 0 {
			lines = append(lines, st.Muted.Render(fmt.Sprintf("  (%d/%d — ↑/↓ to scroll)", start+1, len(plays))))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
