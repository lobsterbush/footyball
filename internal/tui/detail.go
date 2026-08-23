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
	lines = append(lines, st.Muted.Render(m.detailLeague.FullName+" · "+comp.Venue.FullName))
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

	// Box score comparison bars.
	awayBox, hasAway := findBoxTeam(m.summary.Boxscore, away.Team.ID)
	homeBox, hasHome := findBoxTeam(m.summary.Boxscore, home.Team.ID)
	if hasAway && hasHome && len(awayBox.Statistics) > 0 {
		lines = append(lines, st.Gold.Bold(true).Render("TEAM STATS"))
		lines = append(lines, st.Muted.Render(padLeft(truncate(away.Team.Abbreviation, 4), 4)+strings.Repeat(" ", 22)+truncate(home.Team.Abbreviation, 4)))
		for _, as := range awayBox.Statistics {
			hs, found := statByName(homeBox.Statistics, as.Name)
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

	// Scoring plays.
	plays := scoringPlays(m.summary.Plays)
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
