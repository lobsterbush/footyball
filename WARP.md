# footyball

**Status:** Active

A terminal (Bubble Tea) dashboard for Australian sport — AFL, NRL, A-League
Men, A-League Women, NBL, and Super Rugby Pacific — with live scores, box
scores, scoring plays, standings, and team schedules pulled from ESPN's
public API, styled with three Australian-landscape color palettes
(Eucalypt, Ochre, Reef).

**Authors:** Charles Crabtree

## Architecture

- `main.go` — entry point, flag handling, program bootstrap.
- `internal/leagues` — the fixed AFL/NRL/A-League Men/A-League Women/NBL/
  Super Rugby registry.
- `internal/api` — ESPN site-API client (scoreboard, summary, standings,
  team schedule) and response types. Handles ESPN's inconsistent timestamp
  and score encodings across endpoints.
- `internal/theme` — the six palettes (three families × light/dark).
- `internal/config` — preferences persisted to `~/.config/footyball/config.json`.
- `internal/tui` — the Bubble Tea model: dashboard, detail, standings,
  schedule, and league-settings views, one file per view plus shared
  `model.go` / `update.go` / `view.go` / `styles.go`.

## Notes for future work

- Cricket (Big Bash, internationals) is deliberately out of scope — ESPN's
  cricket API models innings/overs rather than quarters, which doesn't fit
  the shared box-score/scoring-play rendering the other six leagues use.
  Adding it would mean a parallel detail view, not just another league entry.
- ESPN's `/teams/{id}/schedule` endpoint returns `score` as an object
  (`{"displayValue": ...}`) while `/scoreboard` and `/summary` return it as a
  plain string — handled by the custom `api.Score` unmarshaler.
- ESPN timestamps drop the seconds field when it's `:00` (e.g.
  `2026-08-23T09:20Z`), which breaks Go's default RFC3339 parsing — handled
  by the custom `api.Time` unmarshaler.
- Standings entries are not guaranteed to arrive in ladder order; they're
  sorted by the `rank` (or `playoffSeed`) stat client-side.
- Test coverage: every package now has unit tests for its pure logic
  (`internal/api`, `internal/config`, `internal/leagues`, `internal/theme`,
  `internal/tui`). The `tui` package tests favor Model literals built
  directly with the fields under test (see `filter_test.go`,
  `standings_test.go`, `detail_test.go`) rather than driving the full Bubble
  Tea update loop, matching `notify_test.go`'s existing style. Nothing in
  `internal/api`'s HTTP fetch functions is tested (they're thin wrappers
  around `get`); the pure parsing/sorting logic they call
  (`FlattenBoxStats`, `sortByRank`/`rankOf`) is covered instead.
- To test the critter easter egg (`internal/tui/critter.go`) quickly,
  temporarily lower `randomCritterDelay`'s range (e.g. to 2 to 4 seconds),
  rebuild, and watch a tmux session: it spawns a kangaroo or koala roughly
  every cycle and it's easy to tell them apart (kangaroo hops in 2-column
  steps, koala ambles in 1-column steps). Restore the original 45 to 120
  second range before committing.
- Manually re-verified (Aug 2026) against a live run: ctrl+c quits from all
  five views, 70-column terminals (both a fresh launch and a live resize)
  render every view without overflow, box scores render correctly for
  nested-category sports (NRL) as well as AFL's flat shape, and league
  settings hide/show/reset all take effect on the dashboard immediately,
  including a fresh fetch for a newly re-shown league.
- Manually re-verified (Aug 2026), extending the pass above to the leagues
  it hadn't specifically exercised: dashboard cards, detail view, standings,
  and team schedule all checked live for A-League Women and the NBL, not
  just AFL/NRL. NBL's ESPN summary endpoint has no team-level box score,
  play-by-play, or leaders data at all for any sampled game (preseason,
  in-progress, and a completed grand final all showed
  `boxscoreSource: "none"`); the app already degrades gracefully by simply
  omitting those sections rather than showing anything garbled, so no
  code change was needed there. NBL's team schedule and standings, which do
  have real data, render correctly, including a team name that overflows
  the standings/schedule truncation width ("South East Melbourne Phoenix")
  and one that lands exactly on it ("Western Sydney Wanderers", 25 chars).
  Separately, confirmed that ESPN's soccer summary endpoint (both A-Leagues)
  never populates `plays`, the field `Summary.Plays` reads; its scoring
  events live under a differently-named `keyEvents` field the app doesn't
  parse, so the SCORING PLAYS section never appears for A-League Men/Women
  even when ESPN has goal-by-goal data for the match. This wasn't fixed in
  this pass (would need new parsing plus a fallback rendering path, since
  `keyEvents` entries carry no running score, only a `scoringPlay` flag) but
  is flagged for a follow-up.
