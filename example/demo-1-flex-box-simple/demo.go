// Package demo1 is the FlexBox Simple demo: a grid of colored cells sized by ratios.
package demo1

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tmctec/stickers/example/about"
	"github.com/tmctec/stickers/v2/flexbox"
)

var (
	style1  = lipgloss.NewStyle().Background(lipgloss.Color("#fc5c65"))
	style2  = lipgloss.NewStyle().Background(lipgloss.Color("#fd9644"))
	style3  = lipgloss.NewStyle().Background(lipgloss.Color("#fed330"))
	style4  = lipgloss.NewStyle().Background(lipgloss.Color("#26de81"))
	style5  = lipgloss.NewStyle().Background(lipgloss.Color("#2bcbba"))
	style6  = lipgloss.NewStyle().Background(lipgloss.Color("#eb3b5a"))
	style7  = lipgloss.NewStyle().Background(lipgloss.Color("#fa8231"))
	style8  = lipgloss.NewStyle().Background(lipgloss.Color("#f7b731"))
	style9  = lipgloss.NewStyle().Background(lipgloss.Color("#20bf6b"))
	style10 = lipgloss.NewStyle().Background(lipgloss.Color("#0fb9b1"))
	style11 = lipgloss.NewStyle().Background(lipgloss.Color("#45aaf2"))
	style12 = lipgloss.NewStyle().Background(lipgloss.Color("#4b7bec"))
	style13 = lipgloss.NewStyle().Background(lipgloss.Color("#a55eea"))
	style14 = lipgloss.NewStyle().Background(lipgloss.Color("#d1d8e0"))
	style15 = lipgloss.NewStyle().Background(lipgloss.Color("#778ca3"))
	style16 = lipgloss.NewStyle().Background(lipgloss.Color("#2d98da"))
	style17 = lipgloss.NewStyle().Background(lipgloss.Color("#3867d6"))
	style18 = lipgloss.NewStyle().Background(lipgloss.Color("#8854d0"))
	style19 = lipgloss.NewStyle().Background(lipgloss.Color("#a5b1c2"))
	style20 = lipgloss.NewStyle().Background(lipgloss.Color("#4b6584"))
)

// Model is the Bubble Tea model for demo 1
type Model struct {
	flexBox   *flexbox.FlexBox
	showAbout bool
	width     int
	height    int
}

const aboutText = `Demo 1: FlexBox Simple

Basic grid of colored cells using ratio-based sizing.

Each cell is created with NewCell(ratioX, ratioY):
- ratioX: width ratio relative to siblings in the row
- ratioY: height ratio relative to other rows

Row 1: 3 cells [1:6] - equal width, tall
Row 2: 7 cells with varying width ratios [2:4] to [4:4]
Row 3: 5 cells [2:5] to [10:5] - center cell widest
Row 4: 5 cells with varying height ratios [1:2] to [1:4]

Press 'a' to close | 'q' to quit`

// New creates a new demo 1 model
func New() *Model {
	m := &Model{
		flexBox: flexbox.New(0, 0),
	}

	rows := []*flexbox.Row{
		m.flexBox.NewRow().AddCells(
			flexbox.NewCell(1, 6).SetStyle(style1),
			flexbox.NewCell(1, 6).SetStyle(style2),
			flexbox.NewCell(1, 6).SetStyle(style3),
		),
		m.flexBox.NewRow().AddCells(
			flexbox.NewCell(2, 4).SetStyle(style4),
			flexbox.NewCell(2, 4).SetStyle(style5),
			flexbox.NewCell(3, 4).SetStyle(style6),
			flexbox.NewCell(3, 4).SetStyle(style7),
			flexbox.NewCell(3, 4).SetStyle(style8),
			flexbox.NewCell(4, 4).SetStyle(style9),
			flexbox.NewCell(4, 4).SetStyle(style10),
		),
		m.flexBox.NewRow().AddCells(
			flexbox.NewCell(2, 5).SetStyle(style11),
			flexbox.NewCell(3, 5).SetStyle(style12),
			flexbox.NewCell(10, 5).SetStyle(style13),
			flexbox.NewCell(3, 5).SetStyle(style14),
			flexbox.NewCell(2, 5).SetStyle(style15),
		),
		m.flexBox.NewRow().AddCells(
			flexbox.NewCell(1, 4).SetStyle(style16),
			flexbox.NewCell(1, 3).SetStyle(style17),
			flexbox.NewCell(1, 2).SetStyle(style18),
			flexbox.NewCell(1, 3).SetStyle(style19),
			flexbox.NewCell(1, 4).SetStyle(style20),
		),
	}

	m.flexBox.AddRows(rows)
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
	view.WindowTitle = "Stickers: FlexBox Simple"
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
