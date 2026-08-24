package api

import "testing"

func entryWithRank(teamID, rank string) StandingsEntry {
	return StandingsEntry{
		Team:  Team{ID: teamID},
		Stats: []StandingsStat{{Name: "rank", DisplayValue: rank}},
	}
}

// TestSortByRankOrdersAscending covers the core reason sortByRank exists:
// ESPN does not guarantee standings entries arrive in ladder order, so they
// must be sorted client-side by rank.
func TestSortByRankOrdersAscending(t *testing.T) {
	entries := []StandingsEntry{
		entryWithRank("t3", "3"),
		entryWithRank("t1", "1"),
		entryWithRank("t2", "2"),
	}
	sortByRank(entries)
	want := []string{"t1", "t2", "t3"}
	for i, w := range want {
		if entries[i].Team.ID != w {
			t.Fatalf("got order %v, want %v", teamIDs(entries), want)
		}
	}
}

// TestSortByRankUsesPlayoffSeedFallback covers leagues (e.g. finals ladders)
// that use "playoffSeed" instead of "rank".
func TestSortByRankUsesPlayoffSeedFallback(t *testing.T) {
	entries := []StandingsEntry{
		{Team: Team{ID: "t2"}, Stats: []StandingsStat{{Name: "playoffSeed", DisplayValue: "2"}}},
		{Team: Team{ID: "t1"}, Stats: []StandingsStat{{Name: "playoffSeed", DisplayValue: "1"}}},
	}
	sortByRank(entries)
	if entries[0].Team.ID != "t1" || entries[1].Team.ID != "t2" {
		t.Fatalf("got order %v, want [t1 t2]", teamIDs(entries))
	}
}

// TestSortByRankMissingRankIsStable covers entries with no rank/playoffSeed
// stat at all: sortByRank must leave the input order untouched rather than
// panic or produce an arbitrary ordering.
func TestSortByRankMissingRankIsStable(t *testing.T) {
	entries := []StandingsEntry{
		{Team: Team{ID: "t1"}},
		{Team: Team{ID: "t2"}},
	}
	sortByRank(entries)
	if entries[0].Team.ID != "t1" || entries[1].Team.ID != "t2" {
		t.Fatalf("expected original order preserved when rank is missing, got %v", teamIDs(entries))
	}
}

func TestRankOfMissingStat(t *testing.T) {
	if _, ok := rankOf(StandingsEntry{}); ok {
		t.Error("expected rankOf to report not-ok for an entry with no rank stat")
	}
}

func TestRankOfNonNumeric(t *testing.T) {
	e := StandingsEntry{Stats: []StandingsStat{{Name: "rank", DisplayValue: "N/A"}}}
	if _, ok := rankOf(e); ok {
		t.Error("expected rankOf to report not-ok for a non-numeric rank value")
	}
}

func teamIDs(entries []StandingsEntry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Team.ID
	}
	return out
}
