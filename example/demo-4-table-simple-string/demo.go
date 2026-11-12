// Package demo4 is the Table Simple String demo: a Table of string columns with sorting and filtering.
package demo4

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"image"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tmctec/stickers/example/about"
	"github.com/tmctec/stickers/example/sample"
	"github.com/tmctec/stickers/example/tableinput"
	"github.com/tmctec/stickers/v2/flexbox"
	"github.com/tmctec/stickers/v2/table"
)

var selectedValue string = "\nselect something with spacebar or enter"

// Model is the Bubble Tea model for demo 4
type Model struct {
	table     *table.Table
	infoBox   *flexbox.FlexBox
	headers   []string
	showAbout bool
	width     int
	height    int
}

const aboutText = `Demo 4: Table Simple String

Navigable table with sorting and filtering.

Demonstrates the Table component with string data
loaded from a CSV file.

Navigation:
- Arrow keys, click, wheel: Move cursor
- Ctrl+S, click a header: Sort by that column (toggle asc/desc)
- Enter/Space: Copy cell value to the clipboard
- Type letters/numbers, paste: Filter current column
- Backspace: Clear filter character

Press 'a' to close | 'q' to quit`

// New creates a new demo 4 model
func New() *Model {
	data, err := csv.NewReader(bytes.NewReader(sample.CSV)).ReadAll()
	if err != nil {
		log.Fatal(err)
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
		table:   table.NewTable(0, 0, headers),
		infoBox: flexbox.New(0, 0).SetHeight(7),
		headers: headers,
	}
	m.table.SetStylePassing(true)
	// setup
	m.table.SetRatio(ratio).SetMinWidth(minSize)
	// add rows
	if _, err := m.table.AddRows(rows); err != nil {
		panic(err)
	}

	// setup info box
	infoText := `
arrows, click, wheel: move the cursor
ctrl+s, click a header: sort
letters, digits, paste: filter
enter, space: copy the cell
q: quit
`
	r1 := m.infoBox.NewRow()
	r1.AddCells(
		flexbox.NewCell(1, 1).
			SetID("info").
			SetContent(infoText),
		flexbox.NewCell(1, 1).
			SetID("info").
			SetContent(selectedValue).
			SetStyle(lipgloss.NewStyle().Bold(true)),
	)
	m.infoBox.AddRows([]*flexbox.Row{r1})

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
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(msg.Height - m.infoBox.GetHeight())
		m.infoBox.SetWidth(msg.Width)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			m.showAbout = !m.showAbout
		case "enter", "space":
			// the value goes to the info box and to the system clipboard
			selectedValue = m.table.GetCursorValue()
			m.infoBox.GetRow(0).GetCell(1).SetContent("\nselected cell: " + selectedValue)
			return m, tea.SetClipboard(selectedValue)
		default:
			tableinput.Update(m.table, msg, image.Point{})
		}
	case tea.MouseClickMsg, tea.MouseWheelMsg, tea.PasteMsg:
		// the table is at the top left corner of the screen
		tableinput.Update(m.table, msg, image.Point{})
	}
	return m, nil
}

// View implements tea.Model
func (m *Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	view.WindowTitle = "Stickers: Table Simple String"
	return view
}

// render returns the demo as text.
func (m *Model) render() string {
	content := lipgloss.JoinVertical(lipgloss.Left, m.table.Render(), m.infoBox.Render())
	if m.showAbout {
		content = about.Over(content, aboutText, m.width, m.height)
	}
	return content
}
