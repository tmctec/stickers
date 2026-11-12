// Package demo2 is the FlexBox Horizontal demo: columns of cells in a HorizontalFlexBox.
package demo2

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tmctec/stickers/example/about"
	"github.com/tmctec/stickers/v2/flexbox"
)

var (
	style1 = lipgloss.NewStyle().Background(lipgloss.Color("#fc5c65"))
	style2 = lipgloss.NewStyle().Background(lipgloss.Color("#fd9644"))
	style3 = lipgloss.NewStyle().Background(lipgloss.Color("#fed330"))
	style4 = lipgloss.NewStyle().Background(lipgloss.Color("#26de81"))
	style5 = lipgloss.NewStyle().Background(lipgloss.Color("#2bcbba"))
	style6 = lipgloss.NewStyle().Background(lipgloss.Color("#eb3b5a"))
)

// Model is the Bubble Tea model for demo 2
type Model struct {
	flexBox   *flexbox.HorizontalFlexBox
	showAbout bool
	width     int
	height    int
}

const aboutText = `Demo 2: FlexBox Horizontal

Column-based layout using HorizontalFlexBox.

Unlike regular FlexBox (rows stacked vertically),
HorizontalFlexBox arranges columns side-by-side,
with cells stacking vertically within each column.

Column 1: 2 cells [1:1] stacked vertically
Column 2: 1 cell [2:1] - takes 2x width ratio
Column 3: 3 cells with varying heights [1:1], [1:2], [1:1]

Press 'a' to close | 'q' to quit`

// New creates a new demo 2 model
func New() *Model {
	m := &Model{
		flexBox: flexbox.NewHorizontal(0, 0),
	}

	columns := []*flexbox.Column{
		m.flexBox.NewColumn().AddCells(
			flexbox.NewCell(1, 1).SetStyle(style1),
			flexbox.NewCell(1, 1).SetStyle(style2),
		),
		m.flexBox.NewColumn().AddCells(
			flexbox.NewCell(2, 1).SetStyle(style3),
		),
		m.flexBox.NewColumn().AddCells(
			flexbox.NewCell(1, 1).SetStyle(style4),
			flexbox.NewCell(1, 2).SetStyle(style5),
			flexbox.NewCell(1, 1).SetStyle(style6),
		),
	}

	m.flexBox.AddColumns(columns)
	return m
}

// Run starts the demo as a standalone program
func Run() {
	m := New()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd { return nil }

// Update implements tea.Model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.flexBox.SetWidth(msg.Width)
		m.flexBox.SetHeight(msg.Height)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			m.showAbout = !m.showAbout
		}
	}
	return m, nil
}

// View implements tea.Model
func (m *Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.WindowTitle = "Stickers: FlexBox Horizontal"
	return view
}

// render returns the demo as text.
func (m *Model) render() string {
	content := m.flexBox.Render()
	if m.showAbout {
		content = about.Over(content, aboutText, m.width, m.height)
	}
	return content
}
