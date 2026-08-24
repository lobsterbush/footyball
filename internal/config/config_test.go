package config

import (
	"testing"

	"github.com/crabtree/footyball/internal/leagues"
)

func TestToggleFavorite(t *testing.T) {
	c := Default()
	if c.IsFavorite("afl", "t1") {
		t.Fatal("expected t1 to not be a favorite initially")
	}
	c.ToggleFavorite("afl", "t1")
	if !c.IsFavorite("afl", "t1") {
		t.Fatal("expected t1 to be a favorite after toggling on")
	}
	// Same team ID under a different league must not be affected.
	if c.IsFavorite("nrl", "t1") {
		t.Fatal("expected favorite to be scoped per league")
	}
	c.ToggleFavorite("afl", "t1")
	if c.IsFavorite("afl", "t1") {
		t.Fatal("expected t1 to no longer be a favorite after toggling off")
	}
}

func TestSetHiddenIdempotent(t *testing.T) {
	c := Default()
	c.SetHidden("afl", true)
	c.SetHidden("afl", true) // setting the same state twice must not duplicate the entry
	if len(c.HiddenLeagues) != 1 {
		t.Fatalf("expected exactly one hidden entry, got %v", c.HiddenLeagues)
	}
	c.SetHidden("afl", false)
	if c.IsHidden("afl") {
		t.Fatal("expected afl to be visible again")
	}
	if len(c.HiddenLeagues) != 0 {
		t.Fatalf("expected no hidden leagues left, got %v", c.HiddenLeagues)
	}
}

func TestOrderedLeaguesAppendsNewlyAddedLeagues(t *testing.T) {
	c := Default()
	// Simulate a config saved before "nbl" existed in the registry: order
	// only lists the other keys.
	c.LeagueOrder = []string{"afl", "nrl"}
	order := c.OrderedLeagues()
	if len(order) != len(leagues.All) {
		t.Fatalf("expected every known league present, got %d of %d", len(order), len(leagues.All))
	}
	if order[0].Key != "afl" || order[1].Key != "nrl" {
		t.Fatalf("expected configured order preserved first, got %v", orderKeys(order))
	}
}

// TestVisibleLeaguesExcludesHidden confirms hidden leagues are dropped from
// the dashboard's list, and reappear once unhidden.
func TestVisibleLeaguesExcludesHidden(t *testing.T) {
	c := Default()
	c.SetHidden("nrl", true)
	visible := c.VisibleLeagues()
	for _, l := range visible {
		if l.Key == "nrl" {
			t.Fatal("expected nrl to be excluded while hidden")
		}
	}
	c.SetHidden("nrl", false)
	visible = c.VisibleLeagues()
	found := false
	for _, l := range visible {
		if l.Key == "nrl" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected nrl to reappear once unhidden")
	}
}

func TestResetLeagueOrderClearsHiddenAndOrder(t *testing.T) {
	c := Default()
	c.SetHidden("afl", true)
	c.LeagueOrder = []string{"nrl", "afl"}
	c.ResetLeagueOrder()
	if len(c.HiddenLeagues) != 0 {
		t.Fatalf("expected no hidden leagues after reset, got %v", c.HiddenLeagues)
	}
	if c.LeagueOrder[0] != leagues.DefaultOrder()[0] {
		t.Fatalf("expected default order restored, got %v", c.LeagueOrder)
	}
}

func TestMoveLeagueSwapsNeighbours(t *testing.T) {
	c := Default()
	order := c.OrderedLeagues()
	first, second := order[0].Key, order[1].Key

	newIdx := c.MoveLeague(0, 1)
	if newIdx != 1 {
		t.Fatalf("expected moved index 1, got %d", newIdx)
	}
	after := c.OrderedLeagues()
	if after[0].Key != second || after[1].Key != first {
		t.Fatalf("expected %s/%s swapped, got %v", first, second, orderKeys(after))
	}
}

// TestMoveLeagueOutOfBoundsIsNoOp confirms moving the first entry further
// left (or the last entry further right) leaves the order untouched instead
// of panicking on an out-of-range index.
func TestMoveLeagueOutOfBoundsIsNoOp(t *testing.T) {
	c := Default()
	before := orderKeys(c.OrderedLeagues())

	if idx := c.MoveLeague(0, -1); idx != 0 {
		t.Fatalf("expected index unchanged at 0, got %d", idx)
	}
	after := orderKeys(c.OrderedLeagues())
	if !equalKeys(before, after) {
		t.Fatalf("expected order unchanged, got %v want %v", after, before)
	}

	last := len(before) - 1
	if idx := c.MoveLeague(last, 1); idx != last {
		t.Fatalf("expected index unchanged at %d, got %d", last, idx)
	}
}

func orderKeys(order []leagues.League) []string {
	out := make([]string, len(order))
	for i, l := range order {
		out[i] = l.Key
	}
	return out
}

func equalKeys(a, b []string) bool {
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
