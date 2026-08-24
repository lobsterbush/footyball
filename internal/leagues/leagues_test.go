package leagues

import "testing"

func TestByKeyFound(t *testing.T) {
	l, ok := ByKey("afl")
	if !ok {
		t.Fatal("expected afl to be found")
	}
	if l.Name != "AFL" {
		t.Errorf("got Name %q, want AFL", l.Name)
	}
}

func TestByKeyNotFound(t *testing.T) {
	if _, ok := ByKey("nope"); ok {
		t.Error("expected unknown key to report not-found")
	}
}

func TestDefaultOrderMatchesAll(t *testing.T) {
	order := DefaultOrder()
	if len(order) != len(All) {
		t.Fatalf("got %d keys, want %d", len(order), len(All))
	}
	for i, l := range All {
		if order[i] != l.Key {
			t.Errorf("index %d: got key %q, want %q", i, order[i], l.Key)
		}
	}
}

// TestAllKeysUnique guards against a copy/paste duplicate key breaking
// config lookups and favorites scoping silently.
func TestAllKeysUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, l := range All {
		if seen[l.Key] {
			t.Errorf("duplicate league key %q", l.Key)
		}
		seen[l.Key] = true
	}
}
