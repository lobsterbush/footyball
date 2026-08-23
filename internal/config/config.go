// Package config persists user preferences (theme, favorites, league
// visibility/order) to $XDG_CONFIG_HOME/footyball/config.json.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/crabtree/footyball/internal/leagues"
)

// Config is the on-disk preferences file.
type Config struct {
	Theme         string   `json:"theme"`
	Favorites     []string `json:"favorites"`      // "leagueKey:teamID"
	LeagueOrder   []string `json:"league_order"`   // league keys, display order
	HiddenLeagues []string `json:"hidden_leagues"` // league keys, currently hidden
}

// Dir returns the directory config.json lives in.
func Dir() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "footyball"), nil
}

func path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads config.json, returning sensible defaults if it doesn't exist.
func Load() (*Config, error) {
	p, err := path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), nil
	}
	if cfg.Theme == "" {
		cfg.Theme = "eucalypt-dark"
	}
	if len(cfg.LeagueOrder) == 0 {
		cfg.LeagueOrder = leagues.DefaultOrder()
	}
	return &cfg, nil
}

// Default returns fresh, empty preferences.
func Default() *Config {
	return &Config{
		Theme:         "eucalypt-dark",
		Favorites:     []string{},
		LeagueOrder:   leagues.DefaultOrder(),
		HiddenLeagues: []string{},
	}
}

// Save writes config.json, creating its directory if needed.
func (c *Config) Save() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	p := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// IsFavorite reports whether a team is favorited.
func (c *Config) IsFavorite(leagueKey, teamID string) bool {
	key := leagueKey + ":" + teamID
	for _, f := range c.Favorites {
		if f == key {
			return true
		}
	}
	return false
}

// ToggleFavorite adds or removes a team from favorites.
func (c *Config) ToggleFavorite(leagueKey, teamID string) {
	key := leagueKey + ":" + teamID
	for i, f := range c.Favorites {
		if f == key {
			c.Favorites = append(c.Favorites[:i], c.Favorites[i+1:]...)
			return
		}
	}
	c.Favorites = append(c.Favorites, key)
}

// IsHidden reports whether a league is currently hidden.
func (c *Config) IsHidden(leagueKey string) bool {
	for _, h := range c.HiddenLeagues {
		if h == leagueKey {
			return true
		}
	}
	return false
}

// SetHidden sets a league's hidden state.
func (c *Config) SetHidden(leagueKey string, hidden bool) {
	already := c.IsHidden(leagueKey)
	if hidden == already {
		return
	}
	if hidden {
		c.HiddenLeagues = append(c.HiddenLeagues, leagueKey)
		return
	}
	for i, h := range c.HiddenLeagues {
		if h == leagueKey {
			c.HiddenLeagues = append(c.HiddenLeagues[:i], c.HiddenLeagues[i+1:]...)
			return
		}
	}
}

// OrderedLeagues resolves the configured order/visibility against the
// known league registry, appending any newly-added leagues at the end.
func (c *Config) OrderedLeagues() []leagues.League {
	seen := map[string]bool{}
	var out []leagues.League
	for _, key := range c.LeagueOrder {
		if l, ok := leagues.ByKey(key); ok {
			out = append(out, l)
			seen[key] = true
		}
	}
	for _, l := range leagues.All {
		if !seen[l.Key] {
			out = append(out, l)
		}
	}
	return out
}

// VisibleLeagues is OrderedLeagues filtered to non-hidden entries.
func (c *Config) VisibleLeagues() []leagues.League {
	var out []leagues.League
	for _, l := range c.OrderedLeagues() {
		if !c.IsHidden(l.Key) {
			out = append(out, l)
		}
	}
	return out
}

// ResetLeagueOrder restores default order and clears all hidden leagues.
func (c *Config) ResetLeagueOrder() {
	c.LeagueOrder = leagues.DefaultOrder()
	c.HiddenLeagues = []string{}
}

// MoveLeague swaps the league at index i with its neighbour at i+delta,
// within the configured order. Returns the new index of the moved league.
func (c *Config) MoveLeague(i, delta int) int {
	order := c.OrderedLeagues()
	j := i + delta
	if i < 0 || i >= len(order) || j < 0 || j >= len(order) {
		return i
	}
	keys := make([]string, len(order))
	for k, l := range order {
		keys[k] = l.Key
	}
	keys[i], keys[j] = keys[j], keys[i]
	c.LeagueOrder = keys
	return j
}
