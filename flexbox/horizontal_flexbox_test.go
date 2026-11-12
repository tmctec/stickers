package flexbox

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// TestHorizontalColumnAlign: columns of a locked width are narrower than the box and sit
// where SetColumnAlign puts them.
func TestHorizontalColumnAlign(t *testing.T) {
	for _, tc := range []struct {
		align lipgloss.Position
		want  string
	}{
		{lipgloss.Left, "a b       "},
		{lipgloss.Center, "   a b    "},
		{lipgloss.Right, "      a b "},
	} {
		box := NewHorizontal(10, 1).LockColumnWidth(2).SetColumnAlign(tc.align)
		box.AddColumns([]*Column{
			box.NewColumn().AddCells(NewCell(1, 1).SetContent("a")),
			box.NewColumn().AddCells(NewCell(1, 1).SetContent("b")),
		})
		assertLines(t, strings.Split(box.Render(), "\n"), tc.want)
	}
}

// TestHorizontalClipsColumnsWiderThanBox: columns that together are wider than the box
// are cut at its edge, never wrapped onto extra lines.
func TestHorizontalClipsColumnsWiderThanBox(t *testing.T) {
	box := NewHorizontal(10, 2).LockColumnWidth(6)
	box.AddColumns([]*Column{
		box.NewColumn().AddCells(NewCell(1, 1).SetContent("aaaaaa\naaaaaa")),
		box.NewColumn().AddCells(NewCell(1, 1).SetContent("bbbbbb\nbbbbbb")),
	})
	assertLines(t, lines(box),
		"aaaaaabbbb",
		"aaaaaabbbb")

	box.LockColumnWidth(0)
	assertLines(t, lines(box),
		"aaaaabbbbb",
		"a    b    ")
}

// TestColumnSizes: a fixed width on a column or on one of its cells pins the column, a
// minimum width holds against a small share, and the other columns share the rest.
func TestColumnSizes(t *testing.T) {
	box := NewHorizontal(40, 4)
	sidebar := box.NewColumn().AddCells(NewCell(1, 1).SetContent("side")).SetFixedWidth(8)
	byCell := box.NewColumn().AddCells(NewCell(1, 1), NewCell(1, 1).SetFixedWidth(6))
	narrow := box.NewColumn().AddCells(NewCell(1, 1).SetMinWidth(10))
	wide := box.NewColumn().AddCells(NewCell(9, 1))
	box.AddColumns([]*Column{sidebar, byCell, narrow, wide})
	box.ForceRecalculate()

	for _, tc := range []struct {
		name   string
		column *Column
		want   int
	}{
		{"column with a fixed width", sidebar, 8},
		{"column of a cell with a fixed width", byCell, 6},
		{"column of a cell with a minimum width", narrow, 10},
		{"column with the rest", wide, 16},
	} {
		if got := tc.column.GetWidth(); got != tc.want {
			t.Errorf("%s is %d columns wide, want %d", tc.name, got, tc.want)
		}
		if got := tc.column.GetHeight(); got != 4 {
			t.Errorf("%s is %d lines tall, want 4", tc.name, got)
		}
	}
	if w, h := box.GetWidth(), box.GetHeight(); w != 40 || h != 4 {
		t.Fatalf("box is %dx%d, want 40x4", w, h)
	}

	sidebar.SetFixedWidth(0)
	box.ForceRecalculate()
	if got := sidebar.GetWidth(); got != 2 {
		t.Fatalf("column back on its ratio is %d columns wide, want 2", got)
	}
}

// TestCellSizesAlongAColumn: a fixed height pins a cell, the others share the rest by
// ratio, and each cell takes its ratioX share of the column's width.
func TestCellSizesAlongAColumn(t *testing.T) {
	title := NewCell(2, 1).SetContent("title").SetFixedHeight(1)
	body := NewCell(1, 3).SetContent("b")
	foot := NewCell(2, 1).SetContent("f")
	box := NewHorizontal(6, 9)
	box.AddColumns([]*Column{box.NewColumn().AddCells(title, body, foot)})
	assertLines(t, lines(box),
		"title ",
		"b     ", "      ", "      ", "      ", "      ", "      ",
		"f     ", "      ")
	if w, h := body.GetWidth(), body.GetHeight(); w != 3 || h != 6 {
		t.Fatalf("body cell is %dx%d, want 3x6", w, h)
	}
}

// TestColumnCutsCellsTallerThanItself: cells that together are taller than their column
// are cut at its bottom edge and the column keeps its frame.
func TestColumnCutsCellsTallerThanItself(t *testing.T) {
	box := NewHorizontal(5, 4)
	column := box.NewColumn().SetStyle(bordered()).AddCells(
		NewCell(1, 1).SetContent("a").SetMinHeight(1),
		NewCell(1, 1).SetContent("b\nc\nd").SetMinHeight(3),
	)
	box.AddColumns([]*Column{column})
	assertLines(t, lines(box), "┌───┐", "│a  │", "│b  │", "└───┘")
}

func TestHorizontalGridAccess(t *testing.T) {
	box := NewHorizontal(6, 1)
	first := box.NewColumn().AddCells(NewCell(1, 1).SetContent("a").SetID("first"))
	second := box.NewColumn().AddCells(NewCell(1, 1).SetContent("b"))
	box.SetColumns([]*Column{first, second})
	box.Render()

	if box.ColumnsLen() != 2 || first.CellsLen() != 1 {
		t.Fatal("ColumnsLen or CellsLen is off")
	}
	if box.GetColumn(1) != second || box.GetColumn(2) != nil || box.GetColumn(-1) != nil {
		t.Fatal("GetColumn returned the wrong column")
	}
	if first.GetCellWithID("first") != first.GetCell(0) || first.GetCell(1) != nil || first.GetCellWithID("x") != nil {
		t.Fatal("GetCell or GetCellWithID returned the wrong cell")
	}

	// copies leave the grid alone
	box.GetColumnCopy(0).GetCell(0).SetContent("x")
	box.GetColumnCellCopy(1, 0).SetContent("y")
	first.GetCellCopy(0).SetContent("z")
	assertLines(t, lines(box), "a  b  ")
	if box.GetColumnCopy(5) != nil || box.GetColumnCellCopy(5, 0) != nil || box.GetColumnCellCopy(0, 5) != nil || first.GetCellCopy(5) != nil {
		t.Fatal("a copy of something that is not there is not nil")
	}
	if c := box.GetColumnCellCopy(1, 0); c.GetWidth() != 3 || c.GetContent() != "b" {
		t.Fatalf("copy is %d wide with content %q, want 3 and b", c.GetWidth(), c.GetContent())
	}

	first.UpdateCellWithIndex(0, NewCell(1, 1).SetContent("c")).UpdateCellWithIndex(7, NewCell(1, 1))
	box.UpdateColumn(1, box.NewColumn().AddCells(NewCell(1, 1).SetContent("d"))).UpdateColumn(7, box.NewColumn())
	assertLines(t, lines(box), "c  d  ")

	box.UseStyle(lipgloss.NewStyle().Align(lipgloss.Right))
	box.GetColumn(0).StylePassing(true)
	assertLines(t, lines(box), "  cd  ")
}
