package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/crabtree/footyball/internal/leagues"
)

const (
	siteBase = "https://site.api.espn.com/apis/site/v2/sports"
	v2Base   = "https://site.api.espn.com/apis/v2/sports"
)

var httpClient = &http.Client{Timeout: 12 * time.Second}

func get(u string, out any) error {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "footyball/0.1 (+https://github.com/crabtree/footyball)")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: unexpected status %s", u, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// FetchScoreboard returns today's slate for a league (ESPN's scoreboard
// endpoint windows to roughly the current day/round automatically).
func FetchScoreboard(l leagues.League) (*ScoreboardResponse, error) {
	u := fmt.Sprintf("%s/%s/%s/scoreboard", siteBase, l.SportSlug, l.LeagueSlug)
	var out ScoreboardResponse
	if err := get(u, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchSummary returns box score and play-by-play detail for one event.
func FetchSummary(l leagues.League, eventID string) (*Summary, error) {
	u := fmt.Sprintf("%s/%s/%s/summary?event=%s", siteBase, l.SportSlug, l.LeagueSlug, url.QueryEscape(eventID))
	var out Summary
	if err := get(u, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchTeamSchedule returns a team's full-season results and fixtures.
func FetchTeamSchedule(l leagues.League, teamID string) (*ScheduleResponse, error) {
	u := fmt.Sprintf("%s/%s/%s/teams/%s/schedule", siteBase, l.SportSlug, l.LeagueSlug, url.QueryEscape(teamID))
	var out ScheduleResponse
	if err := get(u, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchStandings returns the league table, flattened into one or more
// named groups (ESPN nests conferences/divisions to varying depths across
// sports, so this normalizes whatever shape comes back).
func FetchStandings(l leagues.League) ([]StandingsGroup, error) {
	u := fmt.Sprintf("%s/%s/%s/standings", v2Base, l.SportSlug, l.LeagueSlug)
	var node standingsNode
	if err := get(u, &node); err != nil {
		return nil, err
	}
	groups := node.flatten()
	if len(groups) == 0 {
		return nil, fmt.Errorf("%s: no standings data returned", l.Name)
	}
	for i := range groups {
		sortByRank(groups[i].Entries)
	}
	return groups, nil
}

// sortByRank orders standings entries by their "rank" (or "playoffSeed")
// stat ascending. ESPN does not guarantee entries arrive in ladder order.
func sortByRank(entries []StandingsEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		ri, oki := rankOf(entries[i])
		rj, okj := rankOf(entries[j])
		if !oki || !okj {
			return false
		}
		return ri < rj
	})
}

func rankOf(e StandingsEntry) (int, bool) {
	s, ok := e.Stat("rank", "playoffSeed")
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}
