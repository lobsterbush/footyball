package tui

import (
	"testing"

	"github.com/crabtree/footyball/internal/api"
	"github.com/crabtree/footyball/internal/config"
)

func event(id, state, awayID, homeID string) api.Event {
	return api.Event{
		ID: id,
		Competitions: []api.Competition{{
			Status: api.Status{Type: api.StatusType{State: state}},
			Competitors: []api.Competitor{
				{HomeAway: "away", Team: api.Team{ID: awayID, Abbreviation: "AWY"}},
				{HomeAway: "home", Team: api.Team{ID: homeID, Abbreviation: "HOM"}},
			},
		}},
	}
}

func TestNewlyLiveFavorites(t *testing.T) {
	cfg := config.Default()
	cfg.Favorites = []string{"afl:home1"}
	m := Model{cfg: cfg}

	old := &api.ScoreboardResponse{Events: []api.Event{event("g1", "pre", "away1", "home1")}}
	newLive := &api.ScoreboardResponse{Events: []api.Event{event("g1", "in", "away1", "home1")}}

	got := m.newlyLiveFavorites("afl", old, newLive)
	if len(got) != 1 {
		t.Fatalf("expected 1 newly-live favorite, got %d: %v", len(got), got)
	}

	// Same live game again on the next poll: must not re-fire.
	got = m.newlyLiveFavorites("afl", newLive, newLive)
	if len(got) != 0 {
		t.Fatalf("expected no repeat notification, got %v", got)
	}

	// A live game with no favorited team: must not fire.
	noFav := &api.ScoreboardResponse{Events: []api.Event{event("g2", "in", "away9", "home9")}}
	got = m.newlyLiveFavorites("afl", old, noFav)
	if len(got) != 0 {
		t.Fatalf("expected no notification for non-favorited game, got %v", got)
	}

	// First-ever fetch (old == nil): must not fire, even if already live.
	got = m.newlyLiveFavorites("afl", nil, newLive)
	if len(got) != 0 {
		t.Fatalf("expected no notification on initial load, got %v", got)
	}
}
