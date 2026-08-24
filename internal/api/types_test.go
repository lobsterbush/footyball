package api

import "testing"

// TestFlattenBoxStatsFlatShape covers AFL's box score shape: a flat list of
// stats with no nested "stats" and a "label" field already populated.
func TestFlattenBoxStatsFlatShape(t *testing.T) {
	in := []BoxTeamStat{
		{Name: "kicks", Label: "Kicks", DisplayValue: "180"},
		{Name: "marks", Label: "Marks", DisplayValue: "90"},
	}
	got := FlattenBoxStats(in)
	if len(got) != 2 {
		t.Fatalf("flat shape: got %d stats, want 2", len(got))
	}
	if got[0].Label != "Kicks" || got[0].DisplayValue != "180" {
		t.Errorf("flat shape: got %+v, want Kicks/180", got[0])
	}
}

// TestFlattenBoxStatsNestedShape covers the nested-category shape ESPN uses
// for NRL/soccer/basketball: top-level entries carry no leaf DisplayValue of
// their own, only a "stats" sublist, and leaves use "displayName" instead of
// "label". This is exactly the bug that shipped once already (box scores
// only rendering for AFL); confirm the nested categories are flattened into
// leaf stats rather than dropped or rendered as blank rows.
func TestFlattenBoxStatsNestedShape(t *testing.T) {
	in := []BoxTeamStat{
		{
			Name: "passing",
			Stats: []BoxTeamStat{
				{Name: "completions", DisplayName: "Completions", DisplayValue: "24"},
				{Name: "tackles", DisplayName: "Tackles", DisplayValue: "310"},
			},
		},
		{
			Name: "scoring",
			Stats: []BoxTeamStat{
				{Name: "tries", DisplayName: "Tries", DisplayValue: "4"},
			},
		},
	}
	got := FlattenBoxStats(in)
	if len(got) != 3 {
		t.Fatalf("nested shape: got %d stats, want 3 (flattened leaves), got %+v", len(got), got)
	}
	for _, s := range got {
		if s.Label == "" {
			t.Errorf("nested shape: leaf stat %+v has empty Label, want it filled from DisplayName", s)
		}
	}
	if got[0].Label != "Completions" || got[0].DisplayValue != "24" {
		t.Errorf("nested shape: got %+v, want Completions/24 first", got[0])
	}
	if got[2].Label != "Tries" || got[2].DisplayValue != "4" {
		t.Errorf("nested shape: got %+v, want Tries/4 third", got[2])
	}
}

func TestFlattenBoxStatsEmpty(t *testing.T) {
	if got := FlattenBoxStats(nil); len(got) != 0 {
		t.Fatalf("expected no stats for empty input, got %v", got)
	}
}

// TestFlattenBoxStatsMixedShape guards a category that mixes a populated
// Label with no nested stats alongside one with a DisplayName only, in case
// a future sport returns a partially-flat, partially-nested payload.
func TestFlattenBoxStatsMixedShape(t *testing.T) {
	in := []BoxTeamStat{
		{Name: "kicks", Label: "Kicks", DisplayValue: "180"},
		{
			Name: "scoring",
			Stats: []BoxTeamStat{
				{Name: "tries", DisplayName: "Tries", DisplayValue: "4"},
			},
		},
	}
	got := FlattenBoxStats(in)
	if len(got) != 2 {
		t.Fatalf("mixed shape: got %d stats, want 2", len(got))
	}
	if got[0].Label != "Kicks" {
		t.Errorf("mixed shape: got %+v, want Kicks first", got[0])
	}
	if got[1].Label != "Tries" {
		t.Errorf("mixed shape: got %+v, want Tries second", got[1])
	}
}
