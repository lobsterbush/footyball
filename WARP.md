# footyball

**Status:** Active

A terminal (Bubble Tea) dashboard for Australian sport — AFL, NRL, A-League
Men, and Super Rugby Pacific — with live scores, box scores, scoring plays,
standings, and team schedules pulled from ESPN's public API, styled with
three Australian-landscape color palettes (Eucalypt, Ochre, Reef).

**Authors:** Charles Crabtree

## Architecture

- `main.go` — entry point, flag handling, program bootstrap.
- `internal/leagues` — the fixed AFL/NRL/A-League/Super Rugby registry.
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
  the shared box-score/scoring-play rendering the other four leagues use.
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
