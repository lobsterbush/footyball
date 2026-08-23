// Package theme defines footyball's six Aussie-landscape palettes, each
// available in a dark and light variant: Eucalypt (gum leaf green & wattle
// gold), Ochre (red centre rust & cream), and Reef (Great Barrier Sea teal
// & coral).
package theme

import "github.com/charmbracelet/lipgloss"

// Theme is one named palette.
type Theme struct {
	Key     string
	Name    string
	Dark    bool
	BG      lipgloss.Color // page background
	Surface lipgloss.Color // card background
	Border  lipgloss.Color
	Text    lipgloss.Color // primary text
	Muted   lipgloss.Color // secondary text
	Accent  lipgloss.Color // primary brand color (green/ochre/teal)
	Gold    lipgloss.Color // secondary brand color (wattle/cream/coral)
	Live    lipgloss.Color // live-game red
	Win     lipgloss.Color // success green, used for W/finals
}

var themes = []Theme{
	{
		Key: "eucalypt-dark", Name: "Eucalypt (dark)", Dark: true,
		BG: "#10201a", Surface: "#16281f", Border: "#2c4a3a",
		Text: "#eef3ea", Muted: "#93a89a",
		Accent: "#6fae5a", Gold: "#f2b705", Live: "#e8543a", Win: "#7fd858",
	},
	{
		Key: "eucalypt-light", Name: "Eucalypt (light)", Dark: false,
		BG: "#f4f1e6", Surface: "#ffffff", Border: "#d8d2bd",
		Text: "#1c2b20", Muted: "#5c6b5f",
		Accent: "#3f7d34", Gold: "#9a6b06", Live: "#c23b23", Win: "#2e7d32",
	},
	{
		Key: "ochre-dark", Name: "Ochre (dark)", Dark: true,
		BG: "#1f130a", Surface: "#2b190d", Border: "#5a3a20",
		Text: "#f5e9d8", Muted: "#b89a7c",
		Accent: "#c1622d", Gold: "#e8b64a", Live: "#d43d2a", Win: "#8fae4a",
	},
	{
		Key: "ochre-light", Name: "Ochre (light)", Dark: false,
		BG: "#faf3e6", Surface: "#ffffff", Border: "#e3d2b2",
		Text: "#2e1c0f", Muted: "#7a5f42",
		Accent: "#a8501f", Gold: "#8a6108", Live: "#b8351f", Win: "#5f7d2e",
	},
	{
		Key: "reef-dark", Name: "Reef (dark)", Dark: true,
		BG: "#071a1e", Surface: "#0c262b", Border: "#1f4750",
		Text: "#e7f6f4", Muted: "#85aeae",
		Accent: "#1fa8a0", Gold: "#ff7a59", Live: "#ff5a4e", Win: "#35c2a1",
	},
	{
		Key: "reef-light", Name: "Reef (light)", Dark: false,
		BG: "#eef8f7", Surface: "#ffffff", Border: "#c8e6e2",
		Text: "#0b2b2c", Muted: "#3f6663",
		Accent: "#0f8f86", Gold: "#b84a20", Live: "#c23a29", Win: "#1f8f6f",
	},
}

// Default is used before config loads / if a stored key is unrecognized.
func Default() Theme { return themes[0] }

// ByKey looks up a theme by its stable key.
func ByKey(key string) Theme {
	for _, t := range themes {
		if t.Key == key {
			return t
		}
	}
	return Default()
}

// Next returns the theme after the given key, wrapping around, matching
// the given preferred darkness when possible (mirrors the original app's
// "cycle only variants matching your terminal" behavior).
func Next(key string, preferDark bool) Theme {
	idx := 0
	for i, t := range themes {
		if t.Key == key {
			idx = i
			break
		}
	}
	for i := 1; i <= len(themes); i++ {
		cand := themes[(idx+i)%len(themes)]
		if cand.Dark == preferDark {
			return cand
		}
	}
	return themes[(idx+1)%len(themes)]
}
