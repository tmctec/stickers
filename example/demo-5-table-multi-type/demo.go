// Package demo5 is the Table Multi-Type demo: a Table of typed columns that sort numerically.
package demo5

import (
	"fmt"
	"image"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/gocarina/gocsv"
	"github.com/tmctec/stickers/example/about"
	"github.com/tmctec/stickers/example/sample"
	"github.com/tmctec/stickers/example/tableinput"
	"github.com/tmctec/stickers/v2/flexbox"
	"github.com/tmctec/stickers/v2/table"
)

var selectedValue string = "\nselect something with spacebar or enter"

// Model is the Bubble Tea model for demo 5
type Model struct {
	table     *table.Table
	infoBox   *flexbox.FlexBox
	headers   []string
	showAbout bool
	width     int
	height    int
}

const aboutText = `Demo 5: Table Multi-Type

Table with typed columns for proper sorting.

Unlike demo-4 (string-only), this demo uses gocsv
to parse CSV data with typed columns (int, string).

This enables proper numeric sorting: integers sort
numerically (1, 2, 10) instead of lexically (1, 10, 2).

Navigation:
- Arrow keys, click, wheel: Move cursor
- Ctrl+S, click a header: Sort by that column (numeric or alpha)
- Enter/Space: Copy cell value to the clipboard
- Type or paste to filter

Press 'a' to close | 'q' to quit`

// SampleData represents a row from the CSV file
type SampleData struct {
	ID         int    `csv:"id"`
	FirstName  string `csv:"First Name"`
	LastName   string `csv:"Last Name"`
	Age        int    `csv:"Age"`
	Occupation string `csv:"Occupation"`
}

// New creates a new demo 5 model
func New() *Model {
	var sampleData []*SampleData
	if err := gocsv.UnmarshalBytes(sample.CSV, &sampleData); err != nil {
		panic(err)
	}

	headers := []string{"id", "First Name", "Last Name", "Age", "Occupation"}
	ratio := []int{1, 10, 10, 5, 10}
	minSize := []int{4, 5, 5, 2, 5}

	var s string
	var i int
	types := []any{i, s, s, i, s}

	m := &Model{
		table:   table.NewTable(0, 0, headers),
		infoBox: flexbox.New(0, 0).SetHeight(7),
		headers: headers,
	}
	// set types
	if _, err := m.table.SetTypes(types...); err != nil {
		panic(err)
	}
	// setup dimensions
	m.table.SetRatio(ratio).SetMinWidth(minSize)
	// set style passing
	m.table.SetStylePassing(true)
	// add rows
	// with multi type table we have to convert our rows to []any first which is a bit of a pain
	var orderedRows [][]any
	for _, row := range sampleData {
		orderedRows = append(orderedRows, []any{
			row.ID, row.FirstName, row.LastName, row.Age, row.Occupation,
		})
	}
	m.table.MustAddRows(orderedRows)

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
	view.WindowTitle = "Stickers: Table Multi-Type"
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
