// Package demo12 is the Row Alignment demo: rows narrower than the box placed left, center or right.
package demo12

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
	}
)

// Model is the Bubble Tea model for demo 12
type Model struct {
	flexBox   *flexbox.FlexBox
	alignType int // 0=left, 1=center, 2=right
	showAbout bool
	width     int
	height    int
}

const aboutText = `Demo 12: Row Alignment

SetRowAlign places whole rows inside the box. A row
narrower than the box sits left, center or right.
Text inside the cells does not move.

Press 'r' to cycle: Left → Center → Right

Rows 6 & 7 have a fixed width of 140. On a wider
terminal they are narrower than the box and show the
effect. On a narrower one they run off the right edge
and are cut. Other rows take full width and never move.

Press 'a' to close | 'q' to quit`

var alignTypes = []lipgloss.Position{
	lipgloss.Left,
	lipgloss.Center,
	lipgloss.Right,
}

var alignNames = []string{
	"Left",
	"Center",
	"Right",
}

// New creates a new demo 12 model
func New() *Model {
	return &Model{
		flexBox:   flexbox.New(0, 0),
		alignType: 0, // Start with left alignment
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
		m.rebuildFlexBox()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			m.showAbout = !m.showAbout
		case "r":
			m.alignType = (m.alignType + 1) % len(alignTypes)
			m.rebuildFlexBox()
		}
	}
	return m, nil
}

func (m *Model) rebuildFlexBox() {
	// Clear and rebuild
	m.flexBox.SetRows([]*flexbox.Row{})
	m.flexBox.SetRowAlign(alignTypes[m.alignType])

	// Row 1: Full width with thick border (takes up more space due to border)
	row1 := m.flexBox.NewRow()
	cell1 := flexbox.NewCell(1, 1).SetStyle(lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color(colors[0])))
	cell1.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Row 1: Thick Border\nRow align: %s (r)\n%dx%d", alignNames[m.alignType], w+2, h+2))
	})
	row1.AddCells(cell1)

	// Row 2: No border (narrower effective width)
	row2 := m.flexBox.NewRow()
	cell2a := flexbox.NewCell(1, 1).SetStyle(lipgloss.NewStyle().
		Background(lipgloss.Color(colors[1])))
	cell2a.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Row 2A\n%dx%d", w, h))
	})
	cell2b := flexbox.NewCell(1, 1).SetStyle(lipgloss.NewStyle().
		Background(lipgloss.Color(colors[2])))
	cell2b.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Row 2B\n%dx%d", w, h))
	})
	row2.AddCells(cell2a, cell2b)

	// Row 3: Normal border
	row3 := m.flexBox.NewRow()
	cell3 := flexbox.NewCell(1, 1).SetStyle(lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(colors[3])))
	cell3.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Row 3: Normal Border\n%dx%d", w+2, h+2))
	})
	row3.AddCells(cell3)

	// Row 4: Double border (same as thick, different style)
	row4 := m.flexBox.NewRow()
	cell4a := flexbox.NewCell(1, 1).SetStyle(lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(colors[4])))
	cell4a.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Row 4A\n%dx%d", w+2, h+2))
	})
	cell4b := flexbox.NewCell(1, 1).SetStyle(lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(colors[5])))
	cell4b.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Row 4B\n%dx%d", w+2, h+2))
	})
	row4.AddCells(cell4a, cell4b)

	// Row 5: No border again
	row5 := m.flexBox.NewRow()
	for i := 0; i < 3; i++ {
		idx := i
		cell := flexbox.NewCell(1, 1).SetStyle(lipgloss.NewStyle().
			Background(lipgloss.Color(colors[6+i])))
		cell.SetContentGenerator(func(w, h int) string {
			return lipgloss.NewStyle().
				Width(w).Height(h).
				Align(lipgloss.Center, lipgloss.Center).
				Render(fmt.Sprintf("R5-%d\n%dx%d", idx+1, w, h))
		})
		row5.AddCells(cell)
	}

	// Row 6: Rounded border - forced width test
	row6 := m.flexBox.NewRow()
	cell6 := flexbox.NewCell(1, 1).SetStyle(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colors[9]))).
		SetFixedWidth(140) // Force 140 width to test edge alignment
	cell6.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Row 6: Rounded Border\nPress (r) to cycle alignment\n%dx%d", w+2, h+2))
	})
	row6.AddCells(cell6)

	// Row 7: Solid fill at fixed 140 width for comparison
	row7 := m.flexBox.NewRow()
	cell7 := flexbox.NewCell(1, 1).SetStyle(lipgloss.NewStyle().
		Background(lipgloss.Color(colors[0]))).
		SetFixedWidth(140)
	cell7.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Row 7: Solid Fill\n%dx%d", w, h))
	})
	row7.AddCells(cell7)

	m.flexBox.AddRows([]*flexbox.Row{row1, row2, row3, row4, row5, row6, row7})
	m.flexBox.ForceRecalculate()
}

// View implements tea.Model
func (m *Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.WindowTitle = "Stickers: Row Alignment"
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
