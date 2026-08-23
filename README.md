# footyball

A terminal dashboard for Australian sport — AFL, NRL, A-League (men's and
women's), the NBL, and Super Rugby Pacific, live in your terminal. Built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea), in the spirit of
[sportsball](https://codex.humdrum.me/r/sportsball).

## Overview

footyball pulls live scores, box scores, scoring plays, standings, and team
schedules from ESPN's public scoreboard API — no API key required. It ships
with three Australian-landscape color palettes (Eucalypt, Ochre, Reef), each
in a light and dark variant, and follows your terminal's own background so
`t` only cycles through variants that actually fit.

Leagues covered:

- **AFL** — Australian Football League
- **NRL** — National Rugby League
- **A-League Men**
- **A-League Women**
- **NBL** — National Basketball League
- **Super Rugby Pacific**

Every so often, on the dashboard, a kangaroo or koala hops or ambles across
the bottom of the screen — purely decorative, no key required.

Cricket (Big Bash, internationals) isn't covered — ESPN's cricket API is
structured around innings and overs rather than the quarter/half scoreboard
shape the other four leagues share, and folding it in cleanly would need its
own detail view.

## Requirements

- Go 1.21+
- A terminal with 256-color or true-color support for the best look

## Install

```bash
go build -o footyball .
```

or run directly during development:

```bash
go run .
```

## Usage

```bash
footyball            # launch the dashboard
footyball --version  # print version
```

### Keys

**Dashboard**

| Key | Action |
| --- | --- |
| ↑/↓ (k/j) | move focus between leagues |
| ← (h) / → | move selection within a league's games |
| enter / l | open the selected game |
| tab / shift+tab | next / previous league |
| v / V | cycle state filter forward / back (all · live · recent · upcoming) |
| f / F | favorite the away / home team |
| g / G | open the away / home team's schedule |
| s | standings for the focused league |
| L | league settings (show / hide / reorder) |
| r | refresh now |
| t | cycle theme |
| q | quit |

`ctrl+c` quits immediately from any view, not just the dashboard.

**Detail view** — ↑/↓ scroll · f/F favorite · g/G schedule · esc/h back

**Standings** (`s`) — ↑/↓ move · tab/shift+tab switch league · enter team schedule · f favorite · esc/s back

**Team schedule** — ↑/↓ scroll · f favorite · esc/h back

**League settings** (`L`) — ↑/↓ move · space show/hide · K/J reorder · 0 reset · esc done

## Data & config

Scores come from ESPN's public (unofficial) site API. Preferences (favorite
teams, active theme, league order/visibility) persist to
`~/.config/footyball/config.json` (`$XDG_CONFIG_HOME/footyball/config.json`).

## Replication

Everything here is original Go source; there's nothing to regenerate beyond
`go build`. To confirm a working copy:

```bash
go vet ./...
go build ./...
```

## Authors

Charles Crabtree
