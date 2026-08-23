// Package api is a thin client for ESPN's public (unofficial) site API,
// covering the scoreboard, summary, standings, and team-schedule endpoints
// used by footyball. No API key is required.
package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Time wraps time.Time with lenient parsing for ESPN's timestamps, which
// are RFC3339 except they drop the ":00" seconds field whenever it's zero
// (e.g. "2026-08-23T09:20Z" instead of "2026-08-23T09:20:00Z").
type Time struct {
	time.Time
}

var timeLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04Z",
}

// UnmarshalJSON implements json.Unmarshaler with the lenient layouts above.
func (t *Time) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}
	var lastErr error
	for _, layout := range timeLayouts {
		if parsed, err := time.Parse(layout, s); err == nil {
			t.Time = parsed
			return nil
		} else {
			lastErr = err
		}
	}
	return fmt.Errorf("api: cannot parse time %q: %w", s, lastErr)
}

// Score is a competitor's score. ESPN encodes it as a plain string on the
// scoreboard/summary endpoints but as an object ({"displayValue": "..."})
// on the team-schedule endpoint; this accepts either.
type Score string

// UnmarshalJSON accepts either a JSON string or a {"displayValue": "..."} object.
func (s *Score) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err == nil {
		*s = Score(str)
		return nil
	}
	var obj struct {
		DisplayValue string `json:"displayValue"`
	}
	if err := json.Unmarshal(b, &obj); err != nil {
		return fmt.Errorf("api: cannot parse score %s: %w", string(b), err)
	}
	*s = Score(obj.DisplayValue)
	return nil
}

// Team is the subset of ESPN's team object footyball renders.
type Team struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Abbreviation     string `json:"abbreviation"`
	DisplayName      string `json:"displayName"`
	ShortDisplayName string `json:"shortDisplayName"`
	Color            string `json:"color"`
	AlternateColor   string `json:"alternateColor"`
	Logo             string `json:"logo"`
}

// LineScore is one period's score for a competitor.
type LineScore struct {
	DisplayValue string `json:"displayValue"`
	Period       int    `json:"period"`
}

// Competitor is one side of a Competition.
type Competitor struct {
	ID         string      `json:"id"`
	HomeAway   string      `json:"homeAway"`
	Winner     bool        `json:"winner"`
	Team       Team        `json:"team"`
	Score      Score       `json:"score"`
	LineScores []LineScore `json:"linescores"`
}

// StatusType describes a competition's current lifecycle state.
type StatusType struct {
	Name        string `json:"name"`
	State       string `json:"state"` // "pre" | "in" | "post"
	Completed   bool   `json:"completed"`
	Description string `json:"description"`
	Detail      string `json:"detail"`
	ShortDetail string `json:"shortDetail"`
}

// Status wraps clock/period info plus the StatusType.
type Status struct {
	Period       int        `json:"period"`
	DisplayClock string     `json:"displayClock"`
	Type         StatusType `json:"type"`
}

// Venue is where a competition is played.
type Venue struct {
	FullName string `json:"fullName"`
}

// Competition is one match within an Event (almost always exactly one).
type Competition struct {
	ID          string       `json:"id"`
	Date        Time         `json:"date"`
	Venue       Venue        `json:"venue"`
	Status      Status       `json:"status"`
	Competitors []Competitor `json:"competitors"`
}

// Event is one fixture: past, live, or upcoming.
type Event struct {
	ID           string        `json:"id"`
	Date         Time          `json:"date"`
	Name         string        `json:"name"`
	ShortName    string        `json:"shortName"`
	Competitions []Competition `json:"competitions"`
}

// Competition returns the event's primary competition, if any.
func (e Event) Competition() (Competition, bool) {
	if len(e.Competitions) == 0 {
		return Competition{}, false
	}
	return e.Competitions[0], true
}

// Home returns the home competitor, if present.
func (c Competition) Home() (Competitor, bool) {
	for _, comp := range c.Competitors {
		if comp.HomeAway == "home" {
			return comp, true
		}
	}
	return Competitor{}, false
}

// Away returns the away competitor, if present.
func (c Competition) Away() (Competitor, bool) {
	for _, comp := range c.Competitors {
		if comp.HomeAway == "away" {
			return comp, true
		}
	}
	return Competitor{}, false
}

// ScoreboardResponse is the top-level shape of the /scoreboard endpoint.
type ScoreboardResponse struct {
	Events []Event `json:"events"`
}

// ScheduleResponse is the top-level shape of a team's /schedule endpoint.
type ScheduleResponse struct {
	Team   Team    `json:"team"`
	Events []Event `json:"events"`
}

// BoxTeamStat is one labelled statistic for a team in a box score. ESPN
// uses two different shapes here depending on sport: AFL returns a flat
// list of stats directly; most other sports (rugby, soccer, basketball)
// group them into named categories, each carrying its own nested "stats"
// list and using "displayName" instead of "label" for the readable name.
// Flatten (see FlattenBoxStats) before rendering to handle either shape.
type BoxTeamStat struct {
	Name         string        `json:"name"`
	Label        string        `json:"label"`
	DisplayName  string        `json:"displayName"`
	Abbreviation string        `json:"abbreviation"`
	DisplayValue string        `json:"displayValue"`
	Stats        []BoxTeamStat `json:"stats"`
}

// FlattenBoxStats resolves either box-score shape into one flat,
// consistently-labelled list of leaf stats.
func FlattenBoxStats(stats []BoxTeamStat) []BoxTeamStat {
	var out []BoxTeamStat
	for _, s := range stats {
		if len(s.Stats) > 0 {
			out = append(out, FlattenBoxStats(s.Stats)...)
			continue
		}
		if s.Label == "" {
			s.Label = s.DisplayName
		}
		out = append(out, s)
	}
	return out
}

// BoxTeam is one team's row in a box score.
type BoxTeam struct {
	Team       Team          `json:"team"`
	Statistics []BoxTeamStat `json:"statistics"`
}

// Boxscore holds both teams' statistic rows.
type Boxscore struct {
	Teams []BoxTeam `json:"teams"`
}

// PlayTeamRef identifies the team responsible for a play.
type PlayTeamRef struct {
	ID string `json:"id"`
}

// Play is one entry in a match's play-by-play feed.
type Play struct {
	ID        string      `json:"id"`
	Text      string      `json:"text"`
	AwayScore int         `json:"awayScore"`
	HomeScore int         `json:"homeScore"`
	Period    PlayPeriod  `json:"period"`
	Clock     PlayClock   `json:"clock"`
	Team      PlayTeamRef `json:"team"`
}

// PlayPeriod is the quarter/half/session a play occurred in.
type PlayPeriod struct {
	Number int `json:"number"`
}

// PlayClock is the display clock at the time of a play.
type PlayClock struct {
	DisplayValue string `json:"displayValue"`
}

// Summary is the top-level shape of the /summary endpoint.
type Summary struct {
	Boxscore Boxscore      `json:"boxscore"`
	Plays    []Play        `json:"plays"`
	Leaders  []TeamLeaders `json:"leaders"`
}

// LeaderAthlete is the player behind one leaderboard entry.
type LeaderAthlete struct {
	ShortName string `json:"shortName"`
}

// LeaderEntry is one player's value within a leader category.
type LeaderEntry struct {
	DisplayValue string        `json:"displayValue"`
	Athlete      LeaderAthlete `json:"athlete"`
}

// LeaderCategory is one statistical category (e.g. "Goals", "Tries"),
// ranked by player.
type LeaderCategory struct {
	Name        string        `json:"name"`
	DisplayName string        `json:"displayName"`
	Leaders     []LeaderEntry `json:"leaders"`
}

// TeamLeaders holds one team's top performers, not available for every
// sport/game — ESPN populates it for AFL and the NBL but not consistently
// for rugby or soccer.
type TeamLeaders struct {
	Team    Team             `json:"team"`
	Leaders []LeaderCategory `json:"leaders"`
}

// TopPerformer returns the single headline stat leader for this team, if any.
func (t TeamLeaders) TopPerformer() (category string, player string, value string, ok bool) {
	if len(t.Leaders) == 0 || len(t.Leaders[0].Leaders) == 0 {
		return "", "", "", false
	}
	cat := t.Leaders[0]
	top := cat.Leaders[0]
	return cat.DisplayName, top.Athlete.ShortName, top.DisplayValue, true
}

// StandingsStat is one labelled statistic for a standings entry.
type StandingsStat struct {
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	DisplayValue string `json:"displayValue"`
}

// StandingsEntry is one team's row in a standings table.
type StandingsEntry struct {
	Team  Team            `json:"team"`
	Stats []StandingsStat `json:"stats"`
}

// Stat looks up a standings stat by any of the given candidate field names,
// returning the first match. ESPN uses different field names for
// conceptually identical stats across sports (e.g. "gamesWon" vs "wins").
func (e StandingsEntry) Stat(candidates ...string) (string, bool) {
	for _, want := range candidates {
		for _, s := range e.Stats {
			if s.Name == want {
				return s.DisplayValue, true
			}
		}
	}
	return "", false
}

// StandingsGroup is a named table of entries (a whole league, or one
// conference/division within it).
type StandingsGroup struct {
	Name    string
	Entries []StandingsEntry
}

// standingsNode mirrors ESPN's recursive standings shape: a node either
// carries its own "standings" table or nests further "children" nodes.
type standingsNode struct {
	Name      string `json:"name"`
	Standings *struct {
		Entries []StandingsEntry `json:"entries"`
	} `json:"standings"`
	Children []standingsNode `json:"children"`
}

func (n standingsNode) flatten() []StandingsGroup {
	var groups []StandingsGroup
	if n.Standings != nil && len(n.Standings.Entries) > 0 {
		groups = append(groups, StandingsGroup{Name: n.Name, Entries: n.Standings.Entries})
	}
	for _, child := range n.Children {
		groups = append(groups, child.flatten()...)
	}
	return groups
}
