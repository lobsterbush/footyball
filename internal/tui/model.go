// Package tui is footyball's Bubble Tea application: a live dashboard for
// Australian sport (AFL, NRL, A-League Men, Super Rugby Pacific) drawn from
// ESPN's public scoreboard API.
package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/crabtree/footyball/internal/api"
	"github.com/crabtree/footyball/internal/config"
	"github.com/crabtree/footyball/internal/leagues"
	"github.com/crabtree/footyball/internal/theme"
)

type view int

const (
	viewDashboard view = iota
	viewDetail
	viewStandings
	viewSchedule
	viewSettings
)

type stateFilter int

const (
	filterAll stateFilter = iota
	filterLive
	filterRecent
	filterUpcoming
)

func (f stateFilter) label() string {
	switch f {
	case filterLive:
		return "live"
	case filterRecent:
		return "recent"
	case filterUpcoming:
		return "upcoming"
	default:
		return "all"
	}
}

// Model is footyball's single root Bubble Tea model.
type Model struct {
	cfg      *config.Config
	th       theme.Theme
	termDark bool
	width    int
	height   int

	view     view
	prevView view // where esc returns to from schedule
	quitting bool
	pulseOn  bool

	// dashboard state
	leagueList  []leagues.League
	scoreboards map[string]*api.ScoreboardResponse
	sbErr       map[string]error
	loading     map[string]bool
	focusIdx    int            // index into leagueList
	cursor      map[string]int // leagueKey -> selected card index (post-filter)
	filter      stateFilter
	lastRefresh time.Time

	// detail state
	detailLeague leagues.League
	detailEvent  api.Event
	summary      *api.Summary
	summaryErr   error
	summaryLoad  bool
	detailScroll int

	// standings state
	standingsIdx    int // index into leagueList
	standingsGroups []api.StandingsGroup
	standingsErr    error
	standingsLoad   bool
	standingsCursor int

	// schedule state
	scheduleLeague leagues.League
	scheduleTeam   api.Team
	scheduleEvents []api.Event
	scheduleErr    error
	scheduleLoad   bool
	scheduleScroll int

	// league settings state
	settingsCursor int

	status      string // transient footer message (e.g. "refreshed")
	bellPending bool   // rings the terminal bell once, on the next render

	// wildlife: a kangaroo or koala occasionally crosses the dashboard
	critterActive bool
	critterKind   critterKind
	critterX      int
	critterFrame  int
}

// New builds the initial model from loaded (or default) config. termDark
// reflects the terminal's actual background (lipgloss.HasDarkBackground)
// and constrains which theme variants 't' cycles through.
func New(cfg *config.Config, termDark bool) Model {
	leagueList := cfg.VisibleLeagues()

	// Pre-mark every visible league as loading. Without this, the first
	// render (before any fetch has completed, or even been dispatched)
	// finds m.loading[key] false for a key that was never set, which is
	// indistinguishable from "fetch finished, found nothing" — the
	// dashboard would flash "no games" for every league on every launch
	// until the real data arrived, instead of "loading…".
	loading := make(map[string]bool, len(leagueList))
	for _, l := range leagueList {
		loading[l.Key] = true
	}

	return Model{
		cfg:         cfg,
		th:          theme.ByKey(cfg.Theme),
		termDark:    termDark,
		view:        viewDashboard,
		leagueList:  leagueList,
		scoreboards: map[string]*api.ScoreboardResponse{},
		sbErr:       map[string]error{},
		loading:     loading,
		cursor:      map[string]int{},
		filter:      filterAll,
		pulseOn:     true,
	}
}

// Init kicks off the first data fetch and the pulse/refresh timers.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{pulseCmd(), refreshCmd(), critterSpawnCmd()}
	cmds = append(cmds, fetchAllCmd(m.leagueList)...)
	return tea.Batch(cmds...)
}
