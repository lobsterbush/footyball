package tui

import (
	"sort"

	"github.com/crabtree/footyball/internal/api"
	"github.com/crabtree/footyball/internal/leagues"
)

func matchesFilter(e api.Event, f stateFilter) bool {
	comp, ok := e.Competition()
	if !ok {
		return f == filterAll
	}
	switch f {
	case filterLive:
		return comp.Status.Type.State == "in"
	case filterRecent:
		return comp.Status.Type.State == "post"
	case filterUpcoming:
		return comp.Status.Type.State == "pre"
	default:
		return true
	}
}

func (m Model) eventIsFavorite(leagueKey string, e api.Event) bool {
	comp, ok := e.Competition()
	if !ok {
		return false
	}
	for _, c := range comp.Competitors {
		if m.cfg.IsFavorite(leagueKey, c.Team.ID) {
			return true
		}
	}
	return false
}

// leagueEvents returns a league's events narrowed to the active state
// filter, with favorited-team games stably pinned to the front.
func (m Model) leagueEvents(l leagues.League) []api.Event {
	sb, ok := m.scoreboards[l.Key]
	if !ok || sb == nil {
		return nil
	}
	var out []api.Event
	for _, e := range sb.Events {
		if matchesFilter(e, m.filter) {
			out = append(out, e)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		fi, fj := m.eventIsFavorite(l.Key, out[i]), m.eventIsFavorite(l.Key, out[j])
		if fi == fj {
			return false
		}
		return fi
	})
	return out
}
