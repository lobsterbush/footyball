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

// keyEvent builds a fixture api.KeyEvent, mirroring the shape ESPN returns
// for A-League soccer's "keyEvents" field.
func keyEvent(typeText, teamID string, scoringPlay, shootout bool) api.KeyEvent {
	return api.KeyEvent{
		Text:        typeText + " by " + teamID,
		Type:        api.KeyEventType{Text: typeText},
		ScoringPlay: scoringPlay,
		Shootout:    shootout,
		Team:        api.PlayTeamRef{ID: teamID},
	}
}

// TestScoringKeyEventsFiltersToGoals covers a fixture keyEvents feed (the
// A-League Men/Women shape) mixing goals with substitutions and cards. Only
// the scoringPlay:true "Goal" entries should render, each carrying a
// plausible running score tallied by incrementing the scoring team's count,
// since soccer's keyEvents feed has no score field of its own.
func TestScoringKeyEventsFiltersToGoals(t *testing.T) {
	const awayID, homeID = "away1", "home1"
	events := []api.KeyEvent{
		keyEvent("Substitution", homeID, false, false),
		keyEvent("Goal", awayID, true, false),
		keyEvent("Yellow Card", homeID, false, false),
		keyEvent("Goal", homeID, true, false),
		keyEvent("Substitution", awayID, false, false),
		keyEvent("Goal", homeID, true, false),
	}
	got := scoringKeyEvents(events, awayID, homeID)
	if len(got) != 3 {
		t.Fatalf("expected 3 goals to render, got %d: %v", len(got), got)
	}
	wantScores := [][2]int{{0, 1}, {1, 1}, {2, 1}} // {homeScore, awayScore} after each goal
	for i, w := range wantScores {
		if got[i].HomeScore != w[0] || got[i].AwayScore != w[1] {
			t.Errorf("goal %d = (home=%d,away=%d), want (home=%d,away=%d)",
				i, got[i].HomeScore, got[i].AwayScore, w[0], w[1])
		}
	}
}

// TestScoringKeyEventsIncludesPenaltyGoals covers ESPN's distinct type text
// for a penalty-kick goal ("Penalty - Scored" rather than "Goal"), confirmed
// live against completed A-League matches. It carries scoringPlay:true like
// a regular goal and must count toward the running score.
func TestScoringKeyEventsIncludesPenaltyGoals(t *testing.T) {
	const awayID, homeID = "away1", "home1"
	events := []api.KeyEvent{
		keyEvent("Penalty - Scored", homeID, true, false),
		keyEvent("Goal", homeID, true, false),
		keyEvent("Goal", homeID, true, false),
	}
	got := scoringKeyEvents(events, awayID, homeID)
	if len(got) != 3 {
		t.Fatalf("expected 3 goals including the penalty, got %d: %v", len(got), got)
	}
	if got[len(got)-1].HomeScore != 3 {
		t.Errorf("final home score = %d, want 3 (penalty goal must count)", got[len(got)-1].HomeScore)
	}
}

// TestScoringKeyEventsExcludesShootout covers a penalty shootout, whose
// goals shouldn't count toward the normal-time running score even though
// they carry scoringPlay:true and type.text == "Goal".
func TestScoringKeyEventsExcludesShootout(t *testing.T) {
	const awayID, homeID = "away1", "home1"
	events := []api.KeyEvent{
		keyEvent("Goal", homeID, true, false),
		keyEvent("Goal", awayID, true, true), // shootout goal: excluded
		keyEvent("Goal", homeID, true, true), // shootout goal: excluded
	}
	got := scoringKeyEvents(events, awayID, homeID)
	if len(got) != 1 {
		t.Fatalf("expected 1 non-shootout goal, got %d: %v", len(got), got)
	}
	if got[0].HomeScore != 1 || got[0].AwayScore != 0 {
		t.Errorf("got (home=%d,away=%d), want (home=1,away=0)", got[0].HomeScore, got[0].AwayScore)
	}
}

// TestScoringKeyEventsEmpty covers no keyEvents at all, the AFL/NRL/rugby
// case, where the feed lives in "plays" instead and KeyEvents is nil.
func TestScoringKeyEventsEmpty(t *testing.T) {
	if got := scoringKeyEvents(nil, "away1", "home1"); len(got) != 0 {
		t.Fatalf("expected no scoring key events for empty input, got %v", got)
	}
}

// TestScoringPlaysPreferredOverKeyEvents guards the fallback ordering in
// viewDetailBody: when a summary has both a populated "plays" feed (AFL,
// NRL, rugby) and, hypothetically, keyEvents, the existing plays-based path
// must be used untouched rather than falling back to keyEvents tallying.
func TestScoringPlaysPreferredOverKeyEvents(t *testing.T) {
	plays := []api.Play{play(0, 0), play(6, 0)}
	got := scoringPlays(plays)
	if len(got) != 1 || got[0].HomeScore != 6 {
		t.Fatalf("AFL-style plays path affected by keyEvents fallback: got %v", got)
	}
}
