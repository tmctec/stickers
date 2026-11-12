package demo13

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// TestClickSelectsTheCellUnderThePointer: a click anywhere on a cell selects it, the
// cursor goes to the start of its content, and a click on nothing changes nothing.
func TestClickSelectsTheCellUnderThePointer(t *testing.T) {
	m := New()
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.View()

	for index, cell := range m.cells {
		bounds := cell.GetBounds()
		m.Update(tea.MouseClickMsg{X: bounds.Max.X - 1, Y: bounds.Max.Y - 1, Button: tea.MouseLeft})
		if m.selected != index {
			t.Fatalf("a click at the corner of cell %d selected cell %d", index+1, m.selected+1)
		}
		view := m.View()
		want := cell.GetContentBounds().Min
		if view.Cursor == nil || view.Cursor.X != want.X || view.Cursor.Y != want.Y {
			t.Fatalf("the cursor is at %+v, want the content of cell %d at %v", view.Cursor, index+1, want)
		}
		// the cell's label shows at the cursor, inside its border
		line := strings.Split(ansi.Strip(view.Content), "\n")[want.Y]
		if got := ansi.Cut(line, want.X, want.X+4); got != "cell" {
			t.Fatalf("the screen shows %q at the cursor of cell %d, want its label", got, index+1)
		}
	}

	m.Update(tea.MouseClickMsg{X: 200, Y: 5, Button: tea.MouseLeft})
	m.Update(tea.MouseClickMsg{X: 5, Y: 5, Button: tea.MouseRight})
	if m.selected != len(m.cells)-1 {
		t.Fatalf("a click outside the box or with the right button selected cell %d", m.selected+1)
	}
}

func TestArrowsCycleTheSelection(t *testing.T) {
	m := New()
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if m.selected != len(m.cells)-1 {
		t.Fatalf("left from the first cell selected cell %d, want the last", m.selected+1)
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.selected != 1 {
		t.Fatalf("right and down selected cell %d, want cell 2", m.selected+1)
	}

	// the about box hides the cursor
	m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	if view := m.View(); view.Cursor != nil || !strings.Contains(view.Content, "Mouse and Cursor") {
		t.Fatal("the about box does not show, or the cursor shows on top of it")
	}
}

// TestNoCursorWithoutRoom: a window too small for a cell's content shows no cursor, and
// the demo renders at any size.
func TestNoCursorWithoutRoom(t *testing.T) {
	for _, size := range [][2]int{{8, 1}, {0, 0}, {3, 3}, {1, 20}} {
		m := New()
		m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		if view := m.View(); view.Cursor != nil && m.cells[m.selected].GetContentBounds().Empty() {
			t.Fatalf("a %dx%d window shows a cursor at %+v in a cell without room", size[0], size[1], view.Cursor)
		}
		m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
		m.View()
	}
}
