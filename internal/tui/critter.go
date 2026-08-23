package tui

import (
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Every so often, a kangaroo or koala hops/ambles across the bottom of the
// dashboard. Purely decorative — footyball's answer to the classic
// terminal "sl" train easter egg, just more local wildlife.

const (
	critterWidth  = 14
	critterHeight = 4

	kangaroo critterKind = 0
	koala    critterKind = 1
)

type critterKind int

// Two hop/waddle frames per animal, each critterHeight rows of
// critterWidth runes ('█' filled, ' ' empty).
var critterFrames = map[critterKind][2][critterHeight]string{
	kangaroo: {
		{
			"......██......",
			"....████████..",
			"..██..██..████",
			".█....█....██.",
		},
		{
			"......██......",
			"....████████..",
			".██.........██",
			"..............",
		},
	},
	koala: {
		{
			"....██....██..",
			".████████████.",
			".████████████.",
			"..██......██..",
		},
		{
			"....██....██..",
			".████████████.",
			".████████████.",
			"...██....██...",
		},
	},
}

var critterColor = map[critterKind]lipgloss.Color{
	kangaroo: "#c98a4b",
	koala:    "#9aa0a6",
}

// critterStep is how many columns an animal moves per tick — kangaroos hop
// faster than koalas amble.
var critterStep = map[critterKind]int{
	kangaroo: 2,
	koala:    1,
}

func randomCritterDelay() time.Duration {
	return time.Duration(45+rand.Intn(75)) * time.Second
}

func critterSpawnCmd() tea.Cmd {
	return tea.Tick(randomCritterDelay(), func(t time.Time) tea.Msg {
		return critterSpawnMsg(t)
	})
}

func critterMoveCmd() tea.Cmd {
	return tea.Tick(110*time.Millisecond, func(t time.Time) tea.Msg {
		return critterTickMsg(t)
	})
}

type critterSpawnMsg time.Time
type critterTickMsg time.Time

func (m Model) startCritter() Model {
	if rand.Intn(2) == 0 {
		m.critterKind = kangaroo
	} else {
		m.critterKind = koala
	}
	m.critterActive = true
	m.critterX = -critterWidth
	m.critterFrame = 0
	return m
}

func (m Model) viewCritter() string {
	if !m.critterActive || m.view != viewDashboard || m.width <= 0 {
		return ""
	}
	frame := critterFrames[m.critterKind][m.critterFrame]
	style := lipgloss.NewStyle().Foreground(critterColor[m.critterKind])

	lines := make([]string, critterHeight)
	for row := 0; row < critterHeight; row++ {
		buf := make([]rune, m.width)
		for i := range buf {
			buf[i] = ' '
		}
		sprite := []rune(frame[row])
		for i, r := range sprite {
			col := m.critterX + i
			if col >= 0 && col < m.width && r != ' ' {
				buf[col] = r
			}
		}
		lines[row] = style.Render(string(buf))
	}
	return strings.Join(lines, "\n")
}
