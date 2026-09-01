package tui

import (
	"strings"
	"testing"

	"github.com/crabtree/footyball/internal/api"
)

// TestPeriodLabel covers the sport-specific quarter/half prefix used by
// both the live-status fallback clock and the scoring-plays period column.
// AFL and the NBL play quarters; NRL, Super Rugby, and soccer play halves.
func TestPeriodLabel(t *testing.T) {
	cases := map[string]string{
		"australian-football": "Q",
		"basketball":          "Q",
		"rugby-league":        "H",
		"rugby":               "H",
		"soccer":              "H",
		"unknown-sport":       "Q",
	}
	for slug, want := range cases {
		if got := periodLabel(slug); got != want {
			t.Errorf("periodLabel(%q) = %q, want %q", slug, got, want)
		}
	}
}

// TestStatusLabelLiveFallbackClock covers the case where ESPN's
// displayClock is empty for a live game: the fallback must use the
// sport-appropriate period prefix (a soccer match at half 2 reads "H2", not
// the AFL-style "Q2" it rendered before periodLabel existed).
func TestStatusLabelLiveFallbackClock(t *testing.T) {
	status := api.Status{
		Period: 2,
		Type:   api.StatusType{State: "in"},
	}
	got := statusLabel(status, true, "soccer", styleSet{})
	if !strings.Contains(got, "H2") {
		t.Errorf("statusLabel soccer fallback = %q, want it to contain %q", got, "H2")
	}

	got = statusLabel(status, true, "australian-football", styleSet{})
	if !strings.Contains(got, "Q2") {
		t.Errorf("statusLabel AFL fallback = %q, want it to contain %q", got, "Q2")
	}
}

// TestStatusLabelUsesDisplayClockWhenPresent covers the common case where
// ESPN does populate displayClock: the sport-specific fallback must not
// override it.
func TestStatusLabelUsesDisplayClockWhenPresent(t *testing.T) {
	status := api.Status{
		Period:       2,
		DisplayClock: "45'",
		Type:         api.StatusType{State: "in"},
	}
	got := statusLabel(status, true, "soccer", styleSet{})
	if !strings.Contains(got, "45'") {
		t.Errorf("statusLabel = %q, want it to contain the real displayClock %q", got, "45'")
	}
	if strings.Contains(got, "H2") {
		t.Errorf("statusLabel = %q, should not show the period fallback when displayClock is set", got)
	}
}

// TestIndexOfEvent covers the cursor-following lookup used after a
// favorite toggle re-sorts a league's event list: it must find the event by
// ID regardless of position, and fall back to 0 if the event vanished.
func TestIndexOfEvent(t *testing.T) {
	events := []api.Event{
		event("g1", "post", "a1", "h1"),
		event("g2", "in", "a2", "h2"),
		event("g3", "pre", "a3", "h3"),
	}
	if got := indexOfEvent(events, "g3"); got != 2 {
		t.Errorf("indexOfEvent(g3) = %d, want 2", got)
	}
	if got := indexOfEvent(events, "g1"); got != 0 {
		t.Errorf("indexOfEvent(g1) = %d, want 0", got)
	}
	if got := indexOfEvent(events, "missing"); got != 0 {
		t.Errorf("indexOfEvent(missing) = %d, want 0 (fallback)", got)
	}
}
