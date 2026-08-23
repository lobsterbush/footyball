# footyball

I found [sportsball](https://codex.humdrum.me/r/sportsball), a lovely little
terminal dashboard for American sports, and Australia didn't have anything
like it. So now it does.

footyball is a live terminal dashboard for AFL, NRL, both A-Leagues, the NBL,
and Super Rugby Pacific. It runs right in your terminal, no browser or
account needed, and pulls scores straight from ESPN's public API. There are
three color palettes drawn from the Australian landscape, Eucalypt, Ochre,
and Reef, each in a light and dark variant, and it follows whatever your
terminal is already doing, so cycling themes only ever shows you variants
that actually fit.

Star a team and footyball rings the terminal bell the moment their game goes
live, so you can leave it running in a background pane instead of checking
it yourself. Open any game for the full detail view: box score, scoring
plays, and, when ESPN has it, who's leading the game for each side.

Every so often a kangaroo or koala wanders across the bottom of the
dashboard. It doesn't do anything. I like it anyway.

Cricket isn't in here yet. ESPN structures cricket around innings and overs
rather than quarters, so it needs its own detail view, not just another
league entry. That's something for later, and I'd rather say so than
pretend the gap isn't there.

## Requirements

- A terminal with 256-color or true-color support for the best look
- Go 1.21+, only if you're building from source rather than using Homebrew

## Install

```bash
brew tap lobsterbush/footyball
brew install footyball
```

Homebrew gates third-party taps, so it'll ask you to run
`brew trust lobsterbush/footyball` the first time.

To upgrade later:

```bash
brew update && brew upgrade footyball
```

Or build it yourself:

```bash
go build -o footyball .
```

Or run it straight from source during development:

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

**Detail view:** ↑/↓ scroll · f/F favorite · g/G schedule · esc/h back

**Standings** (`s`): ↑/↓ move · tab/shift+tab switch league · enter team schedule · f favorite · esc/s back

**Team schedule:** ↑/↓ scroll · f favorite · esc/h back

**League settings** (`L`): ↑/↓ move · space show/hide · K/J reorder · 0 reset · esc done

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
