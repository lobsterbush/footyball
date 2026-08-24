package tui

import (
	"testing"

	"github.com/crabtree/footyball/internal/api"
	"github.com/crabtree/footyball/internal/config"
	"github.com/crabtree/footyball/internal/leagues"
)

func TestEventIsFavorite(t *testing.T) {
	cfg := config.Default()
	cfg.Favorites = []string{"afl:home1"}
	m := Model{cfg: cfg}

	e := event("g1", "pre", "away1", "home1")
	if !m.eventIsFavorite("afl", e) {
		t.Error("expected event with favorited home team to be a favorite")
	}

	e2 := event("g2", "pre", "away9", "home9")
	if m.eventIsFavorite("afl", e2) {
		t.Error("expected event with no favorited team to not be a favorite")
	}

	// Same team ID favorited under a different league key must not match:
	// favorites are scoped per league.
	e3 := event("g3", "pre", "away1", "home1")
	if m.eventIsFavorite("nrl", e3) {
		t.Error("expected favorite scoped to afl to not apply under nrl")
	}

	// An event with no competitions at all must not be treated as a favorite.
	empty := api.Event{ID: "g4"}
	if m.eventIsFavorite("afl", empty) {
		t.Error("expected event with no competition data to not be a favorite")
	}
}

func TestLeagueEventsFiltersByState(t *testing.T) {
	cfg := config.Default()
	sb := &api.ScoreboardResponse{Events: []api.Event{
		event("live1", "in", "a1", "h1"),
		event("pre1", "pre", "a2", "h2"),
		event("post1", "post", "a3", "h3"),
	}}
	m := Model{cfg: cfg, scoreboards: map[string]*api.ScoreboardResponse{"afl": sb}, filter: filterLive}

	got := m.leagueEvents(leagues.League{Key: "afl"})
	if len(got) != 1 || got[0].ID != "live1" {
		t.Fatalf("filterLive: got %v, want only live1", eventIDs(got))
	}

	m.filter = filterUpcoming
	got = m.leagueEvents(leagues.League{Key: "afl"})
	if len(got) != 1 || got[0].ID != "pre1" {
		t.Fatalf("filterUpcoming: got %v, want only pre1", eventIDs(got))
	}

	m.filter = filterRecent
	got = m.leagueEvents(leagues.League{Key: "afl"})
	if len(got) != 1 || got[0].ID != "post1" {
		t.Fatalf("filterRecent: got %v, want only post1", eventIDs(got))
	}

	m.filter = filterAll
	got = m.leagueEvents(leagues.League{Key: "afl"})
	if len(got) != 3 {
		t.Fatalf("filterAll: got %v, want all 3 events", eventIDs(got))
	}
}

func TestLeagueEventsNoScoreboardYet(t *testing.T) {
	cfg := config.Default()
	m := Model{cfg: cfg, scoreboards: map[string]*api.ScoreboardResponse{}}
	if got := m.leagueEvents(leagues.League{Key: "afl"}); got != nil {
		t.Fatalf("expected nil for a league with no scoreboard loaded yet, got %v", got)
	}
}

// TestLeagueEventsFavoritesPinnedStably guards against the cursor-drift bug
// where a favorite toggle re-sorted the list and the on-screen selection
// silently landed on the wrong team. Favorited events must sort to the
// front, and non-favorited events must keep their original relative order
// (a stable sort), so the rest of the list doesn't shuffle unexpectedly.
func TestLeagueEventsFavoritesPinnedStably(t *testing.T) {
	cfg := config.Default()
	cfg.Favorites = []string{"afl:fav1"}
	sb := &api.ScoreboardResponse{Events: []api.Event{
		event("g1", "pre", "a1", "h1"),
		event("g2", "pre", "a2", "fav1"), // favorited home team
		event("g3", "pre", "a3", "h3"),
		event("g4", "pre", "a4", "h4"),
	}}
	m := Model{cfg: cfg, scoreboards: map[string]*api.ScoreboardResponse{"afl": sb}, filter: filterAll}

	got := m.leagueEvents(leagues.League{Key: "afl"})
	if len(got) != 4 {
		t.Fatalf("expected 4 events, got %d", len(got))
	}
	if got[0].ID != "g2" {
		t.Fatalf("expected favorited game g2 pinned first, got order %v", eventIDs(got))
	}
	// The remaining, non-favorited events must retain their original
	// relative order: g1, g3, g4.
	wantRest := []string{"g1", "g3", "g4"}
	gotRest := eventIDs(got[1:])
	for i, id := range wantRest {
		if gotRest[i] != id {
			t.Fatalf("non-favorited events not stable: got %v, want %v", gotRest, wantRest)
		}
	}
}

// TestLeagueEventsMultipleFavoritesStable checks that when two events are
// both favorited (or both not), their relative order is preserved rather
// than reshuffled arbitrarily by the sort.
func TestLeagueEventsMultipleFavoritesStable(t *testing.T) {
	cfg := config.Default()
	cfg.Favorites = []string{"afl:fav1", "afl:fav2"}
	sb := &api.ScoreboardResponse{Events: []api.Event{
		event("g1", "pre", "a1", "fav1"),
		event("g2", "pre", "a2", "h2"),
		event("g3", "pre", "a3", "fav2"),
	}}
	m := Model{cfg: cfg, scoreboards: map[string]*api.ScoreboardResponse{"afl": sb}, filter: filterAll}

	got := m.leagueEvents(leagues.League{Key: "afl"})
	want := []string{"g1", "g3", "g2"}
	gotIDs := eventIDs(got)
	for i, id := range want {
		if gotIDs[i] != id {
			t.Fatalf("got order %v, want %v", gotIDs, want)
		}
	}
}

func eventIDs(events []api.Event) []string {
	ids := make([]string, len(events))
	for i, e := range events {
		ids[i] = e.ID
	}
	return ids
}
