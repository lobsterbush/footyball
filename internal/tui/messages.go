package tui

import (
	"time"

	"github.com/crabtree/footyball/internal/api"
)

type scoreboardMsg struct {
	key  string
	resp *api.ScoreboardResponse
	err  error
}

type summaryMsg struct {
	resp *api.Summary
	err  error
}

type standingsMsg struct {
	groups []api.StandingsGroup
	err    error
}

type scheduleMsg struct {
	team   api.Team
	events []api.Event
	err    error
}

type pulseMsg time.Time

type refreshMsg time.Time

type statusClearMsg struct{}
