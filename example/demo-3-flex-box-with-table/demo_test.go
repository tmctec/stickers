package demo3

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// TestClickLandsInTheNestedTable: the table sits in a cell in the middle of the grid, and
// a click on one of its values selects that value, because the cell's position gives the
// table's place on screen.
func TestClickLandsInTheNestedTable(t *testing.T) {
	m := New()
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// a click follows a frame, and the frame's layout placed the table's cell
	frame := m.View().Content
	if origin := m.tableOrigin(); origin.X == 0 || origin.Y == 0 {
		t.Fatalf("the table's cell is at %v, want it inside the grid", origin)
	}
	clicked := false
	for y, line := range strings.Split(ansi.Strip(frame), "\n") {
		if before, _, found := strings.Cut(line, "Montgomery"); found {
			m.Update(tea.MouseClickMsg{X: ansi.StringWidth(before) + 2, Y: y, Button: tea.MouseLeft})
			clicked = true
		}
	}
	if !clicked {
		t.Fatal("the table does not show the second row of the sample data")
	}
	if got := m.table.GetCursorValue(); got != "Montgomery" {
		t.Fatalf("the click selected %q, want Montgomery", got)
	}
	if row := m.table.GetCursorRow(); row[1] != "Eddy" {
		t.Fatalf("the click selected the row of %v, want Eddy's", row[1])
	}

	// the wheel and a paste reach the table too
	m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if _, y := m.table.GetCursorLocation(); y != 2 {
		t.Fatalf("the wheel left the cursor on row %d, want 2", y)
	}
	m.Update(tea.PasteMsg{Content: "wat"})
	if column, filter := m.table.GetFilter(); column != 2 || filter != "wat" {
		t.Fatalf("a paste set the filter %q on column %d, want wat on 2", filter, column)
	}
}

// TestEnterCopiesTheValueIntoTheGrid: enter shows the value under the cursor in cells
// around the table and leaves the table's cell alone.
func TestEnterCopiesTheValueIntoTheGrid(t *testing.T) {
	m := New()
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m.View()
	for i := 0; i < 20; i++ {
		m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		if strings.Count(ansi.Strip(m.View().Content), "1") > 1 {
			return
		}
	}
	t.Fatal("twenty presses of enter never copied the value 1 into the grid")
}
