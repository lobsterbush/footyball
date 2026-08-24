package tui

import (
	"testing"

	"github.com/crabtree/footyball/internal/api"
)

func play(homeScore, awayScore int) api.Play {
	return api.Play{HomeScore: homeScore, AwayScore: awayScore}
}

// TestScoringPlaysDetectsDelta checks that scoringPlays keeps only plays
// where either team's score actually changed from the previous play, which
// is how it identifies "scoring plays" without a per-sport list of play
// types.
func TestScoringPlaysDetectsDelta(t *testing.T) {
	plays := []api.Play{
		play(0, 0), // kickoff, no score yet: not a scoring play
		play(6, 0), // home scores
		play(6, 0), // no change: not a scoring play
		play(6, 3), // away scores
		play(9, 3), // home scores again
	}
	got := scoringPlays(plays)
	if len(got) != 3 {
		t.Fatalf("expected 3 scoring plays, got %d: %v", len(got), got)
	}
	wantScores := [][2]int{{6, 0}, {6, 3}, {9, 3}}
	for i, w := range wantScores {
		if got[i].HomeScore != w[0] || got[i].AwayScore != w[1] {
			t.Errorf("scoring play %d = (home=%d,away=%d), want (home=%d,away=%d)",
				i, got[i].HomeScore, got[i].AwayScore, w[0], w[1])
		}
	}
}

func TestScoringPlaysNoScoring(t *testing.T) {
	plays := []api.Play{play(0, 0), play(0, 0)}
	if got := scoringPlays(plays); len(got) != 0 {
		t.Fatalf("expected no scoring plays when nothing changes, got %v", got)
	}
}

func TestScoringPlaysEmpty(t *testing.T) {
	if got := scoringPlays(nil); len(got) != 0 {
		t.Fatalf("expected no scoring plays for empty input, got %v", got)
	}
}

// TestScoringPlaysFirstPlayScoring covers a play-by-play feed that opens
// already on the board (e.g. a resumed/partial feed) — the very first play
// should count as scoring since it differs from the implicit 0-0 start.
func TestScoringPlaysFirstPlayScoring(t *testing.T) {
	plays := []api.Play{play(6, 0)}
	got := scoringPlays(plays)
	if len(got) != 1 {
		t.Fatalf("expected the opening play to count as scoring, got %d", len(got))
	}
}
