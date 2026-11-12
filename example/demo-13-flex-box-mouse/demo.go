// Package demo13 is the Mouse and Cursor demo: cells found by their position.
package demo13

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tmctec/stickers/example/about"
	"github.com/tmctec/stickers/v2/flexbox"
)

var (
	styleCell     = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#778ca3"))
	styleSelected = lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("#f7b731")).Bold(true)
)

// Model is the Bubble Tea model for demo 13
type Model struct {
	flexBox   *flexbox.FlexBox
	cells     []*flexbox.Cell
	selected  int // index into cells
	showAbout bool
	width     int
	height    int
}

const aboutText = `Demo 13: Mouse and Cursor

Cells know where they are in the box.

A click selects the cell under the pointer: CellAt turns
the pointer's position into a cell.

The terminal's cursor sits at the start of the selected
cell's content, the position GetContentBounds returns.

Every cell shows its bounds: (left,top)-(right,bottom).

Keys:
- Click: Select the cell under the pointer
- Arrow keys: Select the next or the previous cell

Press 'a' to close | 'q' to quit`

// New creates a new demo 13 model
func New() *Model {
	m := &Model{flexBox: flexbox.New(0, 0)}

	// one list of width ratios per row; a later row is taller
	shape := [][]int{{1, 2, 1}, {3, 1}, {1, 1, 1, 1}}
	var rows []*flexbox.Row
	for r, ratios := range shape {
		row := m.flexBox.NewRow()
		for _, ratioX := range ratios {
			index := len(m.cells)
			cell := flexbox.NewCell(ratioX, r+1).SetContentGenerator(func(_, _ int) string {
				return m.label(index)
			})
			m.cells = append(m.cells, cell)
			row.AddCells(cell)
		}
		rows = append(rows, row)
	}
	m.flexBox.AddRows(rows)
	m.selectCell(0)
	return m
}

// label names the cell at index and gives its bounds in the box.
func (m *Model) label(index int) string {
	bounds := m.cells[index].GetBounds()
	return fmt.Sprintf("cell %d\n(%d,%d)-(%d,%d)", index+1, bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
}

// selectCell moves the selection to the cell at index, wrapping around at both ends.
func (m *Model) selectCell(index int) {
	m.selected = (index + len(m.cells)) % len(m.cells)
	for i, cell := range m.cells {
		if i == m.selected {
			cell.SetStyle(styleSelected)
		} else {
			cell.SetStyle(styleCell)
		}
	}
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
		case "right", "down":
			m.selectCell(m.selected + 1)
		case "left", "up":
			m.selectCell(m.selected - 1)
		}
	case tea.MouseClickMsg:
		// the box fills the screen, so the pointer's position is a position in the box
		if clicked := m.flexBox.CellAt(msg.X, msg.Y); clicked != nil && msg.Button == tea.MouseLeft {
			for i, cell := range m.cells {
				if cell == clicked {
					m.selectCell(i)
				}
			}
		}
	}
	return m, nil
}

// View implements tea.Model
func (m *Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	view.WindowTitle = "Stickers: Mouse and Cursor"
	// the cursor goes where the selected cell's content starts, when the cell has room
	// for content
	if content := m.cells[m.selected].GetContentBounds(); !m.showAbout && !content.Empty() {
		view.Cursor = tea.NewCursor(content.Min.X, content.Min.Y)
	}
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
