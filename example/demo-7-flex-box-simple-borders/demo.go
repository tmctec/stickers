// Package demo7 is the Simple Borders demo: labeled cells in every border style.
package demo7

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
		"#2d98da", "#3867d6", "#8854d0", "#a5b1c2", "#4b6584",
	}
)

// Model is the Bubble Tea model for demo 7
type Model struct {
	flexBox     *flexbox.FlexBox
	borderType  int     // index into borderTypes; 0 is the filled background
	cellIndexes [][]int // Store cell indexes for each row
	showAbout   bool
	width       int
	height      int
}

const aboutText = `Demo 7: Simple Borders

Grid of labeled cells with different border styles.

Each cell shows its label (A-T) and ratio [X:Y].
The ratio determines relative cell size:
- First number: width ratio within the row
- Second number: height ratio for the row
A cell with room for one line shows only its label.

Press 'b' to cycle styles:
Fill → Hidden → Normal → Rounded → Thick → Double

Press 'a' to close | 'q' to quit`

var borderTypes = []lipgloss.Border{
	{}, // no border
	lipgloss.HiddenBorder(),
	lipgloss.NormalBorder(),
	lipgloss.RoundedBorder(),
	lipgloss.ThickBorder(),
	lipgloss.DoubleBorder(),
}

var borderNames = []string{
	"Filled Background",
	"Hidden Border",
	"Normal Border",
	"Rounded Border",
	"Thick Border",
	"Double Border",
}

// Cell labels
var cellLabels = []string{
	"A", "B", "C",
	"D", "E", "F", "G", "H", "I", "J",
	"K", "L", "M", "N", "O",
	"P", "Q", "R", "S", "T",
}

// New creates a new demo 7 model
func New() *Model {
	m := &Model{
		flexBox:    flexbox.New(0, 0),
		borderType: 2, // Start with normal border
		cellIndexes: [][]int{
			{0, 1, 2},
			{3, 4, 5, 6, 7, 8, 9},
			{10, 11, 12, 13, 14},
			{15, 16, 17, 18, 19},
		},
	}

	// Row 1: 3 cells [1:6]
	row1 := m.flexBox.NewRow()
	for i, r := range [][]int{{1, 6}, {1, 6}, {1, 6}} {
		row1.AddCells(labelCell(i, r[0], r[1]))
	}

	// Row 2: 7 cells with varying width ratios
	row2 := m.flexBox.NewRow()
	for i, r := range [][]int{{2, 4}, {2, 4}, {3, 4}, {3, 4}, {3, 4}, {4, 4}, {4, 4}} {
		row2.AddCells(labelCell(3+i, r[0], r[1]))
	}

	// Row 3: 5 cells, the center one names the current border style
	row3 := m.flexBox.NewRow()
	for i, r := range [][]int{{2, 5}, {3, 5}, {10, 5}, {3, 5}, {2, 5}} {
		cell := labelCell(10+i, r[0], r[1])
		if i == 2 {
			ratioX, ratioY := r[0], r[1]
			cell.SetContentGenerator(func(w, h int) string {
				return cellContent(borderNames[m.borderType]+" (b)", ratioX, ratioY, w, h)
			})
		}
		row3.AddCells(cell)
	}

	// Row 4: 5 cells with varying height ratios; the shorter cells get fewer lines
	row4 := m.flexBox.NewRow()
	for i, r := range [][]int{{1, 4}, {1, 3}, {1, 2}, {1, 3}, {1, 4}} {
		row4.AddCells(labelCell(15+i, r[0], r[1]))
	}

	m.flexBox.AddRows([]*flexbox.Row{row1, row2, row3, row4})
	m.updateStyles()

	return m
}

// labelCell makes a cell whose content is its label and ratio.
func labelCell(idx, ratioX, ratioY int) *flexbox.Cell {
	return flexbox.NewCell(ratioX, ratioY).SetContentGenerator(func(w, h int) string {
		return cellContent(cellLabels[idx], ratioX, ratioY, w, h)
	})
}

// cellContent centers the label and ratio in a w by h area. With one line it shows only
// the label and with none it shows nothing, so a short cell keeps its frame.
func cellContent(label string, ratioX, ratioY, w, h int) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	text := fmt.Sprintf("%s\n[%d:%d]", label, ratioX, ratioY)
	if h == 1 {
		text = label
	}
	return lipgloss.NewStyle().
		Width(w).Height(h).
		Align(lipgloss.Center, lipgloss.Center).
		Render(text)
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
		case "b":
			m.borderType = (m.borderType + 1) % len(borderTypes)
			m.updateStyles()
		}
	}
	return m, nil
}

func (m *Model) updateStyles() {
	for rowIdx := 0; rowIdx < m.flexBox.RowsLen(); rowIdx++ {
		row := m.flexBox.GetRow(rowIdx)
		if row == nil {
			continue
		}
		for cellIdx := 0; cellIdx < row.CellsLen(); cellIdx++ {
			cell := row.GetCell(cellIdx)
			if cell == nil {
				continue
			}
			colorIdx := m.cellIndexes[rowIdx][cellIdx]
			var style lipgloss.Style
			if m.borderType == 0 {
				// Filled background
				style = lipgloss.NewStyle().Background(lipgloss.Color(colors[colorIdx]))
			} else {
				// Border with type
				style = lipgloss.NewStyle().
					Border(borderTypes[m.borderType]).
					BorderForeground(lipgloss.Color(colors[colorIdx]))
			}
			cell.SetStyle(style)
		}
	}
}

// View implements tea.Model
func (m *Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.WindowTitle = "Stickers: Simple Borders"
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
