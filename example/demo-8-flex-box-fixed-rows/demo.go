// Package demo8 is the Fixed Row Heights demo: rows of a fixed height next to rows sized by ratio.
package demo8

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tmctec/stickers/example/about"
	"github.com/tmctec/stickers/v2/flexbox"
)

// Model is the Bubble Tea model for demo 8
type Model struct {
	flexBox    *flexbox.FlexBox
	useFixed   bool
	termWidth  int
	termHeight int
	showAbout  bool
}

const aboutText = `Demo 8: Fixed Row Heights

Demonstrates SetFixedHeight for rows.

In Fixed mode (default):
- Header and Footer rows have fixed 3-row height
- Content area expands to fill remaining space

In Dynamic mode:
- All rows use ratio-based heights
- Header/Footer shrink proportionally

Press 'm' to toggle Fixed/Dynamic mode.

Press 'a' to close | 'q' to quit`

// New creates a new demo 8 model
func New() *Model {
	return &Model{
		flexBox:  flexbox.New(0, 0),
		useFixed: true, // Start with fixed heights enabled
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
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.rebuildFlexBox()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			m.showAbout = !m.showAbout
		case "m":
			m.useFixed = !m.useFixed
			m.rebuildFlexBox()
		}
	}
	return m, nil
}

func (m *Model) rebuildFlexBox() {
	// Create new FlexBox with current terminal size
	m.flexBox = flexbox.New(m.termWidth, m.termHeight)

	// Create header row
	headerRow := m.flexBox.NewRow()
	headerCell := flexbox.NewCell(1, 1)
	headerCell.SetStyle(lipgloss.NewStyle().
		Background(lipgloss.Color("#fc5c65")).
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true))
	headerCell.SetContentGenerator(func(w, h int) string {
		setting := "[1:1 = 1/1w 1/4h]"
		if m.useFixed {
			setting = "[fixed H:3]"
		}
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("HEADER %s\n%dx%d", setting, w, h))
	})
	headerRow.AddCells(headerCell)

	// Set fixed height if enabled
	if m.useFixed {
		headerRow.SetFixedHeight(3)
	}

	// Create content row with two columns
	contentRow := m.flexBox.NewRow()

	// Sidebar cell
	sidebarCell := flexbox.NewCell(1, 2) // Width ratio 1, Height ratio 2
	sidebarCell.SetStyle(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#26de81")))
	sidebarCell.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Sidebar\n[1:2 = 1/4w 2/4h]\n%dx%d", w+2, h+2))
	})

	// Main content cell
	mainCell := flexbox.NewCell(3, 2) // Width ratio 3, Height ratio 2
	mainCell.SetStyle(lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#45aaf2")))

	mainCell.SetContentGenerator(func(w, h int) string {
		mode := "Dynamic (m)"
		setting := "[3:2 = 3/4w 2/4h]"
		desc := "Header/Footer use ratio heights"
		if m.useFixed {
			mode = "Fixed (m)"
			setting = "[3:2 = 3/4w 2/4h]"
			desc = "Header/Footer fixed at 3 rows"
		}
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Fixed Row Heights Demo\n%s %dx%d\n\nMode: %s\n%s", setting, w+2, h+2, mode, desc))
	})

	contentRow.AddCells(sidebarCell, mainCell)
	// Content row remains dynamic (no fixed height)

	// Create footer row
	footerRow := m.flexBox.NewRow()
	footerCell := flexbox.NewCell(1, 1)
	footerCell.SetStyle(lipgloss.NewStyle().
		Background(lipgloss.Color("#fd9644")).
		Foreground(lipgloss.Color("#000000")).
		Bold(true))
	footerCell.SetContentGenerator(func(w, h int) string {
		setting := "[1:1 = 1/1w 1/4h]"
		if m.useFixed {
			setting = "[fixed H:3]"
		}
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("FOOTER %s\n%dx%d", setting, w, h))
	})
	footerRow.AddCells(footerCell)

	// Set fixed height if enabled
	if m.useFixed {
		footerRow.SetFixedHeight(3)
	}

	// Add all rows to FlexBox
	m.flexBox.AddRows([]*flexbox.Row{headerRow, contentRow, footerRow})
}

// View implements tea.Model
func (m *Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.WindowTitle = "Stickers: Fixed Row Heights"
	return view
}

// render returns the demo as text.
func (m *Model) render() string {
	if m.termWidth == 0 || m.termHeight == 0 {
		return "Initializing..."
	}
	m.flexBox.SetWidth(m.termWidth)
	m.flexBox.SetHeight(m.termHeight)
	content := m.flexBox.Render()
	if m.showAbout {
		content = about.Over(content, aboutText, m.termWidth, m.termHeight)
	}
	return content
}
