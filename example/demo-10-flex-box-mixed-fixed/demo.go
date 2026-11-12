// Package demo10 is the Mixed Fixed/Dynamic demo: fixed and ratio sizing in one grid.
package demo10

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

// Model is the Bubble Tea model for demo 10
type Model struct {
	flexBox    *flexbox.FlexBox
	borderType int  // 0=none, 1=normal, 2=rounded, 3=thick, 4=double
	useFixed   bool // Toggle fixed vs dynamic
	hideText   bool // Toggle text visibility
	showAbout  bool
}

const aboutText = `Demo 10: Mixed Fixed/Dynamic

Combines fixed and dynamic sizing in one layout.

In Fixed mode:
- Row 1: fixed height, sidebar/info fixed width
- Row 2: dynamic height, nav/details fixed width
- Row 3: dynamic height, left/right fixed width
- Row 4: fixed height, first/last cell fixed width

In Dynamic mode:
- All cells use ratio-based sizing

Press 'b' border | 'm' mode | 'h' hide text

Press 'a' to close | 'q' to quit`

var borderTypes = []lipgloss.Border{
	{}, // no border
	lipgloss.NormalBorder(),
	lipgloss.RoundedBorder(),
	lipgloss.ThickBorder(),
	lipgloss.DoubleBorder(),
}

var borderNames = []string{
	"Filled Background",
	"Normal Border",
	"Rounded Border",
	"Thick Border",
	"Double Border",
}

// New creates a new demo 10 model
func New() *Model {
	return &Model{
		flexBox:    flexbox.New(0, 0),
		borderType: 1,    // Start with normal border
		useFixed:   true, // Start with fixed sizes enabled
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
		m.flexBox.SetWidth(msg.Width)
		m.flexBox.SetHeight(msg.Height)
		m.rebuildFlexBox()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			m.showAbout = !m.showAbout
		case "b":
			m.borderType = (m.borderType + 1) % len(borderTypes)
			m.rebuildFlexBox()
		case "m":
			m.useFixed = !m.useFixed
			m.rebuildFlexBox()
		case "h":
			m.hideText = !m.hideText
			m.rebuildFlexBox()
		}
	}
	return m, nil
}

func (m *Model) rebuildFlexBox() {
	// Clear and rebuild
	m.flexBox.SetRows([]*flexbox.Row{})

	// Row 1: Header-like row (fixed height when enabled)
	row1 := m.flexBox.NewRow()
	if m.useFixed {
		row1.SetFixedHeight(5) // Set to 5 to fully account for borders
	}

	// Three cells: Fixed sidebar | Dynamic content | Fixed info.
	// Height ratio 4 matches the other rows, so in dynamic mode the row keeps room for
	// its border and three lines of text instead of collapsing to a single line.
	cell1 := flexbox.NewCell(1, 4).SetStyle(m.getCellStyle(0))
	if m.useFixed {
		cell1.SetFixedWidth(25) // Fixed sidebar width
	}
	cell1.SetContentGenerator(func(w, h int) string {
		if m.hideText {
			return lipgloss.NewStyle().Width(w).Height(h).Render("")
		}
		hStatus := "H dynamic"
		wStatus := "W dynamic"
		if m.useFixed {
			hStatus = "H fixed"
			wStatus = "W fixed"
		}
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Sidebar\n%s, %s\n(%dx%d)", hStatus, wStatus, w+m.borderOffset(), h+m.borderOffset()))
	})

	cell2 := flexbox.NewCell(3, 4).SetStyle(m.getCellStyle(1))
	cell2.SetContentGenerator(func(w, h int) string {
		if m.hideText {
			return lipgloss.NewStyle().Width(w).Height(h).Render("")
		}
		hStatus := "H dynamic"
		wStatus := "W dynamic" // Always dynamic width
		if m.useFixed {
			hStatus = "H fixed"
		}
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Header Area\n%s, %s\n(%dx%d)", hStatus, wStatus, w+m.borderOffset(), h+m.borderOffset()))
	})

	cell3 := flexbox.NewCell(1, 4).SetStyle(m.getCellStyle(2))
	if m.useFixed {
		cell3.SetFixedWidth(20) // Fixed info panel width
	}
	cell3.SetContentGenerator(func(w, h int) string {
		if m.hideText {
			return lipgloss.NewStyle().Width(w).Height(h).Render("")
		}
		hStatus := "H dynamic"
		wStatus := "W dynamic"
		if m.useFixed {
			hStatus = "H fixed"
			wStatus = "W fixed"
		}
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Info\n%s, %s\n(%dx%d)", hStatus, wStatus, w+m.borderOffset(), h+m.borderOffset()))
	})

	row1.AddCells(cell1, cell2, cell3)

	// Row 2: Main content row (dynamic height)
	row2 := m.flexBox.NewRow()

	// Seven cells with mixed fixed/dynamic widths
	cells := []struct {
		ratio      int
		fixedWidth int
		label      string
	}{
		{2, 15, "Nav"},     // Fixed narrow nav
		{2, 0, "Col1"},     // Dynamic
		{3, 0, "Main"},     // Dynamic main area
		{3, 0, "Col2"},     // Dynamic
		{3, 0, "Col3"},     // Dynamic
		{4, 0, "Content"},  // Dynamic
		{4, 30, "Details"}, // Fixed details panel
	}

	for i, c := range cells {
		cell := flexbox.NewCell(c.ratio, 4).SetStyle(m.getCellStyle(i + 3))
		if m.useFixed && c.fixedWidth > 0 {
			cell.SetFixedWidth(c.fixedWidth)
		}
		idx := i
		label := c.label
		cell.SetContentGenerator(func(w, h int) string {
			if m.hideText {
				return ""
			}
			hStatus := "H dynamic" // Row 2 is always dynamic height
			wStatus := "W dynamic"
			if m.useFixed && cells[idx].fixedWidth > 0 {
				wStatus = "W fixed"
			}
			return lipgloss.NewStyle().
				Width(w).Height(h).
				Align(lipgloss.Center, lipgloss.Center).
				Render(fmt.Sprintf("%s\n%s, %s\n(%dx%d)", label, hStatus, wStatus, w+m.borderOffset(), h+m.borderOffset()))
		})
		row2.AddCells(cell)
	}

	// Row 3: Mixed row with central focus
	row3 := m.flexBox.NewRow()

	cell31 := flexbox.NewCell(2, 5).SetStyle(m.getCellStyle(10))
	if m.useFixed {
		cell31.SetFixedWidth(20)
	}
	cell31.SetContentGenerator(func(w, h int) string {
		if m.hideText {
			return lipgloss.NewStyle().Width(w).Height(h).Render("")
		}
		hStatus := "H dynamic" // Row 3 is always dynamic height
		wStatus := "W dynamic"
		if m.useFixed {
			wStatus = "W fixed"
		}
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Left\n%s, %s\n(%dx%d)", hStatus, wStatus, w+m.borderOffset(), h+m.borderOffset()))
	})

	cell32 := flexbox.NewCell(3, 5).SetStyle(m.getCellStyle(11))
	cell32.SetContentGenerator(func(w, h int) string {
		if m.hideText {
			return lipgloss.NewStyle().Width(w).Height(h).Render("")
		}
		hStatus := "H dynamic" // Row 3 is always dynamic height
		wStatus := "W dynamic" // Always dynamic width
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Center\n%s, %s\n(%dx%d)", hStatus, wStatus, w+m.borderOffset(), h+m.borderOffset()))
	})

	cell33 := flexbox.NewCell(10, 5).SetStyle(m.getCellStyle(12))
	cell33.SetContentGenerator(func(w, h int) string {
		if m.hideText {
			return lipgloss.NewStyle().Width(w).Height(h).Render("")
		}
		modeText := "Dynamic (m)"
		modeDesc := "All cells use ratios [X:Y = X/Tw Y/Th]"
		if m.useFixed {
			modeText = "Fixed (m)"
			modeDesc = "Some cells have fixed W or H"
		}
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Mixed Fixed/Dynamic Demo\n%s (b) | Text (h)\n%s\n%s\n%dx%d",
				borderNames[m.borderType], modeText, modeDesc, w+m.borderOffset(), h+m.borderOffset()))
	})

	cell34 := flexbox.NewCell(3, 5).SetStyle(m.getCellStyle(13))
	cell34.SetContentGenerator(func(w, h int) string {
		if m.hideText {
			return lipgloss.NewStyle().Width(w).Height(h).Render("")
		}
		hStatus := "H dynamic" // Row 3 is always dynamic height
		wStatus := "W dynamic" // Always dynamic width
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Dynamic\n%s, %s\n(%dx%d)", hStatus, wStatus, w+m.borderOffset(), h+m.borderOffset()))
	})

	cell35 := flexbox.NewCell(2, 5).SetStyle(m.getCellStyle(14))
	if m.useFixed {
		cell35.SetFixedWidth(20)
	}
	cell35.SetContentGenerator(func(w, h int) string {
		if m.hideText {
			return lipgloss.NewStyle().Width(w).Height(h).Render("")
		}
		hStatus := "H dynamic" // Row 3 is always dynamic height
		wStatus := "W dynamic"
		if m.useFixed {
			wStatus = "W fixed"
		}
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Right\n%s, %s\n(%dx%d)", hStatus, wStatus, w+m.borderOffset(), h+m.borderOffset()))
	})

	row3.AddCells(cell31, cell32, cell33, cell34, cell35)

	// Row 4: Footer-like row (fixed height when enabled)
	row4 := m.flexBox.NewRow()
	if m.useFixed {
		row4.SetFixedHeight(6) // Set to 6 - narrow cells F1/F5 need extra height for text wrapping
	}

	// Five cells with varying ratios
	for i := 0; i < 5; i++ {
		ratios := []int{1, 1, 1, 1, 1}
		fixedWidths := []int{12, 0, 0, 0, 12} // First and last fixed

		cell := flexbox.NewCell(ratios[i], 4).SetStyle(m.getCellStyle(15 + i))
		if m.useFixed && fixedWidths[i] > 0 {
			cell.SetFixedWidth(fixedWidths[i])
		}

		idx := i
		cell.SetContentGenerator(func(w, h int) string {
			if m.hideText {
				return ""
			}
			label := fmt.Sprintf("F%d", idx+1)
			hStatus := "H dynamic"
			wStatus := "W dynamic"
			if m.useFixed {
				hStatus = "H fixed" // Row 4 has fixed height when useFixed
				if fixedWidths[idx] > 0 {
					wStatus = "W fixed"
					label = fmt.Sprintf("Fix%d", idx+1)
				}
			}
			return lipgloss.NewStyle().
				Width(w).Height(h).
				Align(lipgloss.Center, lipgloss.Center).
				Render(fmt.Sprintf("%s\n%s, %s\n(%dx%d)", label, hStatus, wStatus, w+m.borderOffset(), h+m.borderOffset()))
		})
		row4.AddCells(cell)
	}

	m.flexBox.AddRows([]*flexbox.Row{row1, row2, row3, row4})
}

func (m *Model) getCellStyle(colorIndex int) lipgloss.Style {
	if m.borderType == 0 {
		// Filled background
		return lipgloss.NewStyle().Background(lipgloss.Color(colors[colorIndex]))
	}
	// Border with type
	return lipgloss.NewStyle().
		Border(borderTypes[m.borderType]).
		BorderForeground(lipgloss.Color(colors[colorIndex]))
}

func (m *Model) borderOffset() int {
	if m.borderType == 0 {
		return 0
	}
	return 2
}

// View implements tea.Model
func (m *Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.WindowTitle = "Stickers: Mixed Fixed/Dynamic"
	return view
}

// render returns the demo as text.
func (m *Model) render() string {
	content := m.flexBox.Render()
	if m.showAbout {
		width := m.flexBox.GetWidth()
		height := m.flexBox.GetHeight()
		content = about.Over(content, aboutText, width, height)
	}
	return content
}
