package tui

import (
	"testing"

	"github.com/crabtree/footyball/internal/api"
)

func standingsStat(name, value string) api.StandingsStat {
	return api.StandingsStat{Name: name, DisplayValue: value}
}

// TestActiveStandingsColumnsAFLShape mirrors a sport that carries every
// stat, including drawn games (e.g. AFL, NRL).
func TestActiveStandingsColumnsAFLShape(t *testing.T) {
	entries := []api.StandingsEntry{
		{Stats: []api.StandingsStat{
			standingsStat("rank", "1"),
			standingsStat("gamesPlayed", "22"),
			standingsStat("gamesWon", "18"),
			standingsStat("gamesDrawn", "0"),
			standingsStat("gamesLost", "4"),
			standingsStat("winPercent", "0.818"),
			standingsStat("points", "72"),
			standingsStat("pointsDifference", "350"),
		}},
	}
	cols := activeStandingsColumns(entries)
	headers := columnHeaders(cols)
	want := []string{"GP", "W", "D", "L", "PCT", "PTS", "DIFF"}
	if !equalStrings(headers, want) {
		t.Fatalf("AFL-shaped entry: got columns %v, want %v", headers, want)
	}
}

// TestActiveStandingsColumnsBasketballShape mirrors basketball's standings,
// which have no gamesDrawn/ties field and no ladder "points" stat — the
// exact case activeStandingsColumns was added to handle (see the "make
// standings columns sport-aware" commit). Confirm the column set actually
// adapts instead of showing a blank/dashed D or PTS column.
func TestActiveStandingsColumnsBasketballShape(t *testing.T) {
	entries := []api.StandingsEntry{
		{Stats: []api.StandingsStat{
			standingsStat("rank", "1"),
			standingsStat("gamesPlayed", "28"),
			standingsStat("wins", "20"),
			standingsStat("losses", "8"),
			standingsStat("winPercent", "0.714"),
		}},
	}
	cols := activeStandingsColumns(entries)
	headers := columnHeaders(cols)

	for _, absent := range []string{"D", "PTS", "DIFF"} {
		for _, h := range headers {
			if h == absent {
				t.Errorf("basketball-shaped entry should not show a %s column, got columns %v", absent, headers)
			}
		}
	}
	want := []string{"GP", "W", "L", "PCT"}
	if !equalStrings(headers, want) {
		t.Fatalf("basketball-shaped entry: got columns %v, want %v", headers, want)
	}
}

func TestActiveStandingsColumnsEmpty(t *testing.T) {
	if cols := activeStandingsColumns(nil); cols != nil {
		t.Fatalf("expected nil columns for no entries, got %v", cols)
	}
}

func columnHeaders(cols []standingsColumn) []string {
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = c.header
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
