package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/crabtree/footyball/internal/theme"
)

// Update is the single entry point Bubble Tea calls for every message.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case scoreboardMsg:
		m.loading[msg.key] = false
		if msg.err != nil {
			m.sbErr[msg.key] = msg.err
			return m, nil
		}
		delete(m.sbErr, msg.key)
		newlyLive := m.newlyLiveFavorites(msg.key, m.scoreboards[msg.key], msg.resp)
		m.scoreboards[msg.key] = msg.resp
		if _, ok := m.cursor[msg.key]; !ok {
			m.cursor[msg.key] = 0
		}
		if len(newlyLive) > 0 {
			m.bellPending = true
			cmds := []tea.Cmd{m.setStatus(newlyLive[0]), func() tea.Msg { return bellRungMsg{} }}
			return m, tea.Batch(cmds...)
		}
		return m, nil

	case summaryMsg:
		m.summaryLoad = false
		m.summaryErr = msg.err
		if msg.err == nil {
			m.summary = msg.resp
		}
		return m, nil

	case standingsMsg:
		m.standingsLoad = false
		m.standingsErr = msg.err
		if msg.err == nil {
			m.standingsGroups = msg.groups
		}
		return m, nil

	case scheduleMsg:
		m.scheduleLoad = false
		m.scheduleErr = msg.err
		if msg.err == nil {
			m.scheduleTeam = msg.team
			m.scheduleEvents = msg.events
		}
		return m, nil

	case pulseMsg:
		m.pulseOn = !m.pulseOn
		return m, pulseCmd()

	case refreshMsg:
		m.lastRefresh = time.Now()
		cmds := fetchAllCmd(m.leagueList)
		for _, l := range m.leagueList {
			m.loading[l.Key] = true
		}
		cmds = append(cmds, refreshCmd())
		return m, tea.Batch(cmds...)

	case statusClearMsg:
		m.status = ""
		return m, nil

	case bellRungMsg:
		m.bellPending = false
		return m, nil

	case critterSpawnMsg:
		if m.critterActive {
			return m, critterSpawnCmd()
		}
		m = m.startCritter()
		return m, critterMoveCmd()

	case critterTickMsg:
		if !m.critterActive {
			return m, nil
		}
		m.critterX += critterStep[m.critterKind]
		m.critterFrame = 1 - m.critterFrame
		if m.critterX > m.width+critterWidth {
			m.critterActive = false
			return m, critterSpawnCmd()
		}
		return m, critterMoveCmd()
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys, available from every view. ctrl+c always quits
	// immediately, matching the universal terminal expectation; q is a
	// dashboard-only shortcut for it since every other view already binds
	// esc/h to "back".
	switch key {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "q":
		if m.view == viewDashboard {
			m.quitting = true
			return m, tea.Quit
		}
	case "t":
		m.th = theme.Next(m.cfg.Theme, m.termDark)
		m.cfg.Theme = m.th.Key
		_ = m.cfg.Save()
		return m, nil
	}

	switch m.view {
	case viewDashboard:
		return m.updateDashboard(key)
	case viewDetail:
		return m.updateDetail(key)
	case viewStandings:
		return m.updateStandings(key)
	case viewSchedule:
		return m.updateSchedule(key)
	case viewSettings:
		return m.updateSettings(key)
	}
	return m, nil
}

func (m *Model) setStatus(s string) tea.Cmd {
	m.status = s
	return clearStatusCmd()
}
