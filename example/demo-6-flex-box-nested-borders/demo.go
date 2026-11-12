// Package demo6 is the Nested Borders demo: boxes nested in each other through a content generator.
package demo6

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tmctec/stickers/example/about"
	"github.com/tmctec/stickers/v2/flexbox"
)

var (
	colors = []string{
		"#fc5c65", "#fd9644", "#fed330", "#26de81", "#2bcbba",
		"#eb3b5a", "#fa8231", "#f7b731", "#20bf6b", "#0fb9b1",
		"#45aaf2", "#4b7bec", "#a55eea", "#d1d8e0", "#778ca3",
		"#fc5c65", "#fd9644", "#fed330", "#26de81", "#2bcbba",
		"#eb3b5a", "#fa8231", "#f7b731", "#20bf6b", "#0fb9b1",
		"#45aaf2", "#4b7bec", "#a55eea", "#d1d8e0", "#778ca3",
	}

	borderTypes = []lipgloss.Border{
		lipgloss.NormalBorder(),
		lipgloss.RoundedBorder(),
		lipgloss.ThickBorder(),
		lipgloss.DoubleBorder(),
	}

	borderNames = []string{
		"Normal Border",
		"Rounded Border",
		"Thick Border",
		"Double Border",
	}
)

// Model is the Bubble Tea model for demo 6
type Model struct {
	flexBox    *flexbox.FlexBox
	borderType int // 0=normal, 1=rounded, 2=thick, 3=double
	showAbout  bool
	width      int
	height     int
}

const aboutText = `Demo 6: Nested Borders

Recursive FlexBox with nested borders.

Uses ContentGenerator to create infinitely nested
FlexBoxes, each with a colored border. Nesting
continues until the space becomes too small.

The depth counter shows how many levels deep
the recursion reached.

Press 'b' to cycle border types:
Normal → Rounded → Thick → Double

Press 'a' to close | 'q' to quit`

// New creates a new demo 6 model
func New() *Model {
	m := &Model{
		flexBox:    flexbox.New(0, 0),
		borderType: 1, // Start with rounded border
	}

	// Create a single cell that will contain nested FlexBoxes
	row := m.flexBox.NewRow()
	cell := flexbox.NewCell(1, 1)

	// Use ContentGenerator to create nested borders
	cell.SetContentGenerator(func(maxX, maxY int) string {
		return createNestedFlexBox(maxX, maxY, 0, m.borderType)
	})

	row.AddCells(cell)
	m.flexBox.AddRows([]*flexbox.Row{row})

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

// createNestedFlexBox recursively creates FlexBoxes with borders
// depth tracks which color to use and when to stop
// borderType selects which border style to use
func createNestedFlexBox(width, height, depth, borderType int) string {
	// Stop when we run out of space or colors
	// Minimum needed: 2 chars width (left+right border), 2 chars height (top+bottom border)
	if width < 4 || height < 4 || depth >= len(colors) {
		// Return centered text at the deepest level
		style := lipgloss.NewStyle().
			Width(width).
			Height(height).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#ffffff"))
		return style.Render(fmt.Sprintf("Depth: %d\n%s (b)", depth, borderNames[borderType]))
	}

	// Create a FlexBox for this level
	box := flexbox.New(width, height)
	row := box.NewRow()
	cell := flexbox.NewCell(1, 1)

	// Create border style with current color and selected border type
	borderStyle := lipgloss.NewStyle().
		Border(borderTypes[borderType]).
		BorderForeground(lipgloss.Color(colors[depth]))

	cell.SetStyle(borderStyle)

	// Set content generator to recursively create inner FlexBox
	cell.SetContentGenerator(func(innerWidth, innerHeight int) string {
		return createNestedFlexBox(innerWidth, innerHeight, depth+1, borderType)
	})

	row.AddCells(cell)
	box.AddRows([]*flexbox.Row{row})

	return box.Render()
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
		case "b":
			m.borderType = (m.borderType + 1) % len(borderTypes)
		}
	}
	return m, nil
}

// View implements tea.Model
func (m *Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.WindowTitle = "Stickers: Nested Borders"
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
