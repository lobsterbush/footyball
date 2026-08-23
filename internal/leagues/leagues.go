// Package leagues defines the fixed set of Australian competitions footyball tracks.
package leagues

// League identifies one ESPN competition endpoint and how to present it.
type League struct {
	Key        string // stable identifier used in config and lookups, e.g. "afl"
	Name       string // short display name, e.g. "AFL"
	FullName   string // e.g. "Australian Football League"
	SportSlug  string // ESPN sport path segment, e.g. "australian-football"
	LeagueSlug string // ESPN league path segment, e.g. "afl"
	Mark       string // single/double-glyph mark shown on cards and headers
}

// All is the full known set, in default display order.
var All = []League{
	{
		Key:        "afl",
		Name:       "AFL",
		FullName:   "Australian Football League",
		SportSlug:  "australian-football",
		LeagueSlug: "afl",
		Mark:       "AF",
	},
	{
		Key:        "nrl",
		Name:       "NRL",
		FullName:   "National Rugby League",
		SportSlug:  "rugby-league",
		LeagueSlug: "3",
		Mark:       "NR",
	},
	{
		Key:        "aleague",
		Name:       "A-LEAGUE M",
		FullName:   "A-League Men",
		SportSlug:  "soccer",
		LeagueSlug: "aus.1",
		Mark:       "AL",
	},
	{
		Key:        "aleaguew",
		Name:       "A-LEAGUE W",
		FullName:   "A-League Women",
		SportSlug:  "soccer",
		LeagueSlug: "aus.w.1",
		Mark:       "AW",
	},
	{
		Key:        "nbl",
		Name:       "NBL",
		FullName:   "National Basketball League",
		SportSlug:  "basketball",
		LeagueSlug: "nbl",
		Mark:       "NB",
	},
	{
		Key:        "srp",
		Name:       "SUPER RUGBY",
		FullName:   "Super Rugby Pacific",
		SportSlug:  "rugby",
		LeagueSlug: "242041",
		Mark:       "SR",
	},
}

// ByKey looks up a league by its stable key.
func ByKey(key string) (League, bool) {
	for _, l := range All {
		if l.Key == key {
			return l, true
		}
	}
	return League{}, false
}

// DefaultOrder returns the default key ordering.
func DefaultOrder() []string {
	keys := make([]string, len(All))
	for i, l := range All {
		keys[i] = l.Key
	}
	return keys
}
