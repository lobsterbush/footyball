package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/crabtree/footyball/internal/api"
	"github.com/crabtree/footyball/internal/leagues"
)

func fetchScoreboardCmd(l leagues.League) tea.Cmd {
	return func() tea.Msg {
		resp, err := api.FetchScoreboard(l)
		return scoreboardMsg{key: l.Key, resp: resp, err: err}
	}
}

func fetchAllCmd(ls []leagues.League) []tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(ls))
	for _, l := range ls {
		cmds = append(cmds, fetchScoreboardCmd(l))
	}
	return cmds
}

func fetchSummaryCmd(l leagues.League, eventID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := api.FetchSummary(l, eventID)
		return summaryMsg{resp: resp, err: err}
	}
}

func fetchStandingsCmd(l leagues.League) tea.Cmd {
	return func() tea.Msg {
		groups, err := api.FetchStandings(l)
		return standingsMsg{groups: groups, err: err}
	}
}

func fetchScheduleCmd(l leagues.League, team api.Team) tea.Cmd {
	return func() tea.Msg {
		resp, err := api.FetchTeamSchedule(l, team.ID)
		if err != nil {
			return scheduleMsg{team: team, err: err}
		}
		return scheduleMsg{team: team, events: resp.Events}
	}
}

func pulseCmd() tea.Cmd {
	return tea.Tick(650*time.Millisecond, func(t time.Time) tea.Msg {
		return pulseMsg(t)
	})
}

func refreshCmd() tea.Cmd {
	return tea.Tick(60*time.Second, func(t time.Time) tea.Msg {
		return refreshMsg(t)
	})
}

func clearStatusCmd() tea.Cmd {
	return tea.Tick(2500*time.Millisecond, func(time.Time) tea.Msg {
		return statusClearMsg{}
	})
}
