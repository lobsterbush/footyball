# VHS tape for footyball's README demo.
# Regenerate with: XDG_CONFIG_HOME=/tmp/footyball-demo-config vhs demo/demo.tape
# (run from the repo root; a fresh XDG_CONFIG_HOME keeps the recording on
# the default Eucalypt theme with no leftover favorites from local testing)

Output demo/demo.gif

Set Shell "bash"
Set FontSize 15
Set Width 1400
Set Height 1500
Set Padding 20
Set BorderRadius 8
Set TypingSpeed 60ms

Type "/tmp/footyball_demo_bin"
Enter
Sleep 3s

# Browse AFL's live cards
Right
Sleep 400ms
Right
Sleep 400ms
Right
Sleep 700ms

# Open a game: real box score + scoring plays
Enter
Sleep 2500ms
Down
Sleep 500ms
Down
Sleep 900ms
Escape
Sleep 500ms

# AFL standings: real ladder
Type "s"
Sleep 1800ms
Down
Sleep 350ms
Down
Sleep 350ms
Down
Sleep 700ms
Escape
Sleep 600ms

# Tour the other leagues
Tab
Sleep 500ms
Tab
Sleep 500ms
Tab
Sleep 500ms
Tab
Sleep 500ms
Tab
Sleep 900ms

# Cycle themes
Type "t"
Sleep 1300ms
Type "t"
Sleep 1300ms
Type "t"
Sleep 1600ms

Type "q"
Sleep 1s
