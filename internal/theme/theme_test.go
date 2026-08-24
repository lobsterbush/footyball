package theme

import "testing"

func TestByKeyFound(t *testing.T) {
	th := ByKey("ochre-dark")
	if th.Key != "ochre-dark" {
		t.Fatalf("got key %q, want ochre-dark", th.Key)
	}
}

func TestByKeyUnknownFallsBackToDefault(t *testing.T) {
	th := ByKey("does-not-exist")
	if th.Key != Default().Key {
		t.Fatalf("got key %q, want default %q", th.Key, Default().Key)
	}
}

// TestNextStaysWithinPreferredDarkness confirms 't' only cycles through
// variants matching the terminal's actual background, e.g. a dark terminal
// should never land on a light theme.
func TestNextStaysWithinPreferredDarkness(t *testing.T) {
	for _, preferDark := range []bool{true, false} {
		next := ByKey("eucalypt-dark").Key
		seen := map[string]bool{}
		for i := 0; i < len(themes); i++ {
			th := Next(next, preferDark)
			if th.Dark != preferDark {
				t.Fatalf("Next(%q, preferDark=%v) = %q (Dark=%v), want Dark=%v",
					next, preferDark, th.Key, th.Dark, preferDark)
			}
			seen[th.Key] = true
			next = th.Key
		}
		// Three theme families each have one variant per darkness, so a
		// full cycle should visit exactly 3 distinct themes.
		if len(seen) != 3 {
			t.Errorf("preferDark=%v: cycled through %d distinct themes, want 3: %v", preferDark, len(seen), seen)
		}
	}
}

func TestNextWraps(t *testing.T) {
	// Starting from the last dark theme, Next should wrap back to the first.
	first := Next("reef-dark", true)
	if first.Key != "eucalypt-dark" {
		t.Fatalf("expected wrap to eucalypt-dark, got %q", first.Key)
	}
}

func TestNextUnknownKeyStillReturnsAValidTheme(t *testing.T) {
	th := Next("bogus-key", true)
	if th.Key == "" {
		t.Fatal("expected Next to return a valid theme even for an unrecognized starting key")
	}
}
