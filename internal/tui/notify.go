package tui

import (
	"fmt"

	"github.com/crabtree/footyball/internal/api"
)

// newlyLiveFavorites compares a league's previous and freshly-fetched
// scoreboards and reports any games that just went live and involve a
// favorited team. old is nil on a league's very first fetch, in which case
// nothing is reported — we only want to flag transitions, not every game
// that happens to already be live when footyball starts up.
func (m Model) newlyLiveFavorites(leagueKey string, old, new *api.ScoreboardResponse) []string {
	if old == nil || new == nil {
		return nil
	}
	wasLive := map[string]bool{}
	for _, e := range old.Events {
		if comp, ok := e.Competition(); ok {
			wasLive[e.ID] = comp.Status.Type.State == "in"
		}
	}

	var out []string
	for _, e := range new.Events {
		comp, ok := e.Competition()
		if !ok || comp.Status.Type.State != "in" || wasLive[e.ID] {
			continue
		}
		if !m.eventIsFavorite(leagueKey, e) {
			continue
		}
		away, _ := comp.Away()
		home, _ := comp.Home()
		out = append(out, fmt.Sprintf("● %s vs %s is live!", away.Team.Abbreviation, home.Team.Abbreviation))
	}
	return out
}
