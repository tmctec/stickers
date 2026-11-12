// Package demo3 is the FlexBox with Table demo: a Table inside a cell of a FlexBox, filled from CSV data.
package demo3

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"image"
	"math/rand"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tmctec/stickers/example/about"
	"github.com/tmctec/stickers/example/sample"
	"github.com/tmctec/stickers/example/tableinput"
	"github.com/tmctec/stickers/v2/flexbox"
	"github.com/tmctec/stickers/v2/table"
)

var (
	styleRow        = lipgloss.NewStyle().Align(lipgloss.Center).Foreground(lipgloss.Color("#000000")).Bold(true)
	styleBlank      = lipgloss.NewStyle()
	styleBackground = lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("#ffffff"))
	style1          = lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("#f368e0"))
	style2          = lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("#ff9f43"))
	style3          = lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("#ee5253"))
	style4          = lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("#0abde3"))
	style5          = lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("#10ac84"))
	style6          = lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("#222f3e"))
)

// Model is the Bubble Tea model for demo 3
type Model struct {
	flexBox   *flexbox.FlexBox
	table     *table.Table
	tableCell *flexbox.Cell // the cell of flexBox that holds the table
	headers   []string
	showAbout bool
	width     int
	height    int
}

const aboutText = `Demo 3: FlexBox with Table

Embedded table inside a flexbox layout with CSV data.

Shows how to nest a Table component inside a FlexBox cell.
The table is placed in the center cell of row 2: the cell's
content generator sizes the table to the cell and renders it,
and the cell's position tells where a mouse click landed in it.

Navigation:
- Arrow keys, click, wheel: Move cursor in table
- Ctrl+S, click a header: Sort by that column
- Enter/Space: Copy cell value to random flexbox cells
- Type letters, paste: Filter current column
- Backspace: Clear filter character

Press 'a' to close | 'q' to quit`

// New creates a new demo 3 model
func New() *Model {
	data, err := csv.NewReader(bytes.NewReader(sample.CSV)).ReadAll()
	if err != nil {
		panic(err)
	}

	headers := data[0]
	rows := make([][]any, len(data[1:]))
	for i, row := range data[1:] {
		rows[i] = make([]any, len(headers))
		for j, cell := range row {
			rows[i][j] = cell
		}
	}
	ratio := []int{1, 10, 10, 5, 10}
	minSize := []int{4, 5, 5, 2, 5}

	m := &Model{
		flexBox: flexbox.New(0, 0).SetStyle(styleBackground),
		table:   table.NewTable(0, 0, headers),
		headers: headers,
	}

	m.table.SetRatio(ratio).SetMinWidth(minSize)
	if _, err := m.table.AddRows(rows); err != nil {
		panic(err)
	}
	m.table.SetStylePassing(true)

	// the table takes the size of its cell on every render
	m.tableCell = flexbox.NewCell(10, 5).SetStyle(styleBlank).SetContentGenerator(func(width, height int) string {
		return m.table.SetWidth(width).SetHeight(height).Render()
	})

	r1 := m.flexBox.NewRow().AddCells(
		flexbox.NewCell(5, 5).SetStyle(style2),
		flexbox.NewCell(2, 5).SetStyle(style3),
		flexbox.NewCell(5, 5).SetStyle(style5),
	).SetStyle(styleRow)
	r2 := m.flexBox.NewRow().AddCells(
		flexbox.NewCell(1, 5).SetStyle(style6),
		m.tableCell,
		flexbox.NewCell(1, 5).SetStyle(style6),
	).SetStyle(styleRow)
	r3 := m.flexBox.NewRow().AddCells(
		flexbox.NewCell(1, 5).SetStyle(style5),
		flexbox.NewCell(1, 4).SetStyle(style4),
		flexbox.NewCell(1, 3).SetStyle(style3),
		flexbox.NewCell(1, 4).SetStyle(style2),
		flexbox.NewCell(1, 5).SetStyle(style1),
	).SetStyle(styleRow)

	_rows := []*flexbox.Row{r1, r2, r3}
	m.flexBox.AddRows(_rows)

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
		windowHeight := msg.Height
		windowWidth := msg.Width
		m.width = windowWidth
		m.height = windowHeight
		m.flexBox.SetWidth(windowWidth)
		m.flexBox.SetHeight(windowHeight)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			m.showAbout = !m.showAbout
		case "enter", "space":
			m.copyToGrid(m.table.GetCursorValue())
		default:
			tableinput.Update(m.table, msg, m.tableOrigin())
		}
	case tea.MouseClickMsg, tea.MouseWheelMsg, tea.PasteMsg:
		tableinput.Update(m.table, msg, m.tableOrigin())
	}
	return m, nil
}

// tableOrigin returns where the table's top left corner is on screen: the content area
// of the cell that holds it. Mouse positions are counted from there.
func (m *Model) tableOrigin() image.Point {
	return m.tableCell.GetContentBounds().Min
}

// copyToGrid shows the value in about half of the cells around the table, picked at
// random, and empties the others.
func (m *Model) copyToGrid(value string) {
	for ir := 0; ir < m.flexBox.RowsLen(); ir++ {
		// the middle row holds the table
		if ir == 1 {
			continue
		}
		row := m.flexBox.GetRow(ir)
		for ic := 0; ic < row.CellsLen(); ic++ {
			cell := row.GetCell(ic)
			if rand.Int()%2 == 0 { // #nosec G404 -- picks a demo layout, not a secret
				cell.SetContent(strings.Repeat("\n", cell.GetHeight()/2) + value)
			} else {
				cell.SetContent("")
			}
		}
	}
}

// View implements tea.Model
func (m *Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	view.WindowTitle = "Stickers: FlexBox with Table"
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
