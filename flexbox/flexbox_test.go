package flexbox

import (
	"image"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// lines renders box and returns its lines.
func lines(box interface{ Render() string }) []string {
	return strings.Split(box.Render(), "\n")
}

// assertSize fails unless rendered is exactly height lines of width columns.
func assertSize(t *testing.T, rendered string, width, height int) {
	t.Helper()
	if height == 0 {
		if rendered != "" {
			t.Fatalf("rendered %q, want nothing", rendered)
		}
		return
	}
	got := strings.Split(rendered, "\n")
	if len(got) != height {
		t.Fatalf("rendered %d lines, want %d:\n%s", len(got), height, rendered)
	}
	for i, line := range got {
		if w := lipgloss.Width(line); w != width {
			t.Fatalf("line %d is %d columns wide, want %d:\n%s", i, w, width, rendered)
		}
	}
}

func TestFlexBoxRendersExactlyItsSize(t *testing.T) {
	styles := map[string]lipgloss.Style{
		"plain":   lipgloss.NewStyle(),
		"border":  bordered(),
		"padding": lipgloss.NewStyle().Padding(1, 2),
		"margin":  lipgloss.NewStyle().Margin(1, 3),
		"all":     bordered().Padding(1, 2).Margin(1),
	}
	sizes := [][2]int{{0, 0}, {1, 1}, {3, 2}, {10, 3}, {40, 12}, {7, 30}, {0, 4}}
	for name, style := range styles {
		for _, size := range sizes {
			width, height := size[0], size[1]

			box := New(width, height).SetStyle(style)
			box.AddRows([]*Row{
				box.NewRow().AddCells(NewCell(1, 1).SetContent("header")).SetFixedHeight(1),
				box.NewRow().AddCells(NewCell(1, 2).SetStyle(bordered()).SetContent("left"), NewCell(3, 1).SetContent("some longer text")),
				box.NewRow().SetStyle(bordered()).AddCells(NewCell(1, 1).SetContent("framed row").SetFixedWidth(60)),
			})
			t.Run("rows "+name, func(t *testing.T) { assertSize(t, box.Render(), width, height) })

			horizontal := NewHorizontal(width, height).SetStyle(style)
			horizontal.AddColumns([]*Column{
				horizontal.NewColumn().AddCells(NewCell(1, 1).SetContent("side")).SetFixedWidth(6),
				horizontal.NewColumn().AddCells(NewCell(2, 1).SetStyle(bordered()).SetContent("top"), NewCell(1, 3).SetContent("bottom")),
				horizontal.NewColumn().SetStyle(bordered()).AddCells(NewCell(1, 1).SetContent("framed column").SetMinHeight(40)),
			})
			t.Run("columns "+name, func(t *testing.T) { assertSize(t, horizontal.Render(), width, height) })
		}
	}
}

// TestBoxPaddingShrinksTheRows: padding takes its columns and lines from the rows and
// the box keeps its size.
func TestBoxPaddingShrinksTheRows(t *testing.T) {
	box := New(12, 4).SetStyle(lipgloss.NewStyle().Padding(1, 2))
	box.AddRows([]*Row{box.NewRow().AddCells(NewCell(1, 1).SetContent("aaaa"), NewCell(1, 1).SetContent("bbbb"))})
	assertLines(t, lines(box),
		"            ",
		"  aaaabbbb  ",
		"            ",
		"            ")
}

func TestRowPaddingShrinksTheCells(t *testing.T) {
	box := New(12, 3)
	box.AddRows([]*Row{box.NewRow().SetStyle(lipgloss.NewStyle().Padding(1, 2)).
		AddCells(NewCell(1, 1).SetContent("aaaaaa"), NewCell(1, 1).SetContent("bbbbbb"))})
	assertLines(t, lines(box),
		"            ",
		"  aaaabbbb  ",
		"            ")
}

func TestEmptyGridsRender(t *testing.T) {
	box := New(4, 2)
	assertLines(t, lines(box), "    ", "    ")

	box.AddRows([]*Row{box.NewRow()})
	assertLines(t, lines(box), "    ", "    ")

	box.GetRow(0).AddCells(NewCell(0, 0).SetContent("x"))
	assertLines(t, lines(box), "    ", "    ")

	horizontal := NewHorizontal(4, 2)
	assertLines(t, lines(horizontal), "    ", "    ")

	horizontal.AddColumns([]*Column{horizontal.NewColumn()})
	assertLines(t, lines(horizontal), "    ", "    ")

	horizontal.GetColumn(0).AddCells(NewCell(0, 0).SetContent("x"))
	assertLines(t, lines(horizontal), "    ", "    ")
}

// TestRowsWithoutLinesTakeNoRoom: when the box is too short for every row, the rows left
// without lines must not push the footer out.
func TestRowsWithoutLinesTakeNoRoom(t *testing.T) {
	box := New(6, 3)
	box.AddRows([]*Row{
		box.NewRow().AddCells(NewCell(1, 1).SetContent("header")).SetFixedHeight(1),
		box.NewRow().AddCells(NewCell(1, 1).SetContent("one")),
		box.NewRow().AddCells(NewCell(1, 2).SetContent("two")),
		box.NewRow().AddCells(NewCell(1, 3).SetContent("three")),
		box.NewRow().AddCells(NewCell(1, 1).SetContent("footer")).SetFixedHeight(1),
	})
	assertLines(t, lines(box), "header", "three ", "footer")
}

// TestSettersApplyOnTheNextRender: every setter takes effect without another call, on
// the box, on a row and on a cell kept from before the first render.
func TestSettersApplyOnTheNextRender(t *testing.T) {
	left, right := NewCell(1, 1).SetContent("a"), NewCell(1, 1).SetContent("b")
	box := New(10, 1)
	row := box.NewRow().AddCells(left, right)
	box.AddRows([]*Row{row})
	assertLines(t, lines(box), "a    b    ")

	left.SetFixedWidth(2)
	assertLines(t, lines(box), "a b       ")

	left.SetFixedWidth(0).SetMinWidth(7)
	assertLines(t, lines(box), "a      b  ")

	left.SetMinWidth(0).SetContent("c")
	assertLines(t, lines(box), "c    b    ")

	right.SetStyle(lipgloss.NewStyle().Align(lipgloss.Right))
	assertLines(t, lines(box), "c        b")

	box.SetWidth(6)
	assertLines(t, lines(box), "c    b")

	box.SetHeight(2).SetWidth(4)
	assertLines(t, lines(box), "c  b", "    ")

	row.SetFixedHeight(1)
	box.AddRows([]*Row{box.NewRow().AddCells(NewCell(1, 1).SetContent("d"))})
	assertLines(t, lines(box), "c  b", "d   ")

	box.SetHeight(4).LockRowHeight(1)
	assertLines(t, lines(box), "c  b", "d   ", "    ", "    ")

	box.LockRowHeight(0)
	assertLines(t, lines(box), "c  b", "d   ", "    ", "    ")

	row.SetFixedHeight(0)
	assertLines(t, lines(box), "c  b", "    ", "d   ", "    ")

	box.SetStyle(bordered())
	assertLines(t, lines(box), "┌──┐", "│cb│", "│d │", "└──┘")

	box.SetRows([]*Row{row})
	assertLines(t, lines(box), "┌──┐", "│cb│", "│  │", "└──┘")

	row.UpdateCellWithIndex(0, NewCell(1, 1).SetContent("e"))
	assertLines(t, lines(box), "┌──┐", "│eb│", "│  │", "└──┘")
}

// TestSetWidthAloneKeepsTheFrame: the rows share what the border leaves, also when only
// one of the two sizes changes.
func TestSetWidthAloneKeepsTheFrame(t *testing.T) {
	box := New(6, 3).SetStyle(bordered())
	box.AddRows([]*Row{box.NewRow().AddCells(NewCell(1, 1).SetContent("abcdefghij"))})
	box.Render()
	box.SetWidth(8)
	assertLines(t, lines(box), "┌──────┐", "│abcdef│", "└──────┘")

	horizontal := NewHorizontal(6, 3).SetStyle(bordered())
	horizontal.AddColumns([]*Column{horizontal.NewColumn().AddCells(NewCell(1, 1).SetContent("a\nb\nc"))})
	horizontal.Render()
	horizontal.SetHeight(4)
	assertLines(t, lines(horizontal), "┌────┐", "│a   │", "│b   │", "└────┘")
}

func TestStylePassing(t *testing.T) {
	right := lipgloss.NewStyle().Align(lipgloss.Right)
	center := lipgloss.NewStyle().Align(lipgloss.Center)

	cell := NewCell(1, 1).SetContent("a")
	box := New(5, 1)
	row := box.NewRow().AddCells(cell)
	box.AddRows([]*Row{row})
	assertLines(t, lines(box), "a    ")

	// a style reaches the cells only through a row that passes it
	box.UseStyle(right)
	assertLines(t, lines(box), "a    ")
	row.StylePassing(true)
	assertLines(t, lines(box), "    a")

	// the row's own keys win over the box's
	row.SetStyle(center)
	assertLines(t, lines(box), "  a  ")

	// the cell's own keys win over both
	cell.SetStyle(lipgloss.NewStyle().Align(lipgloss.Left))
	assertLines(t, lines(box), "a    ")
	cell.SetStyle(lipgloss.NewStyle())

	// a change of a passed style reaches the cells
	row.SetStyle(right)
	assertLines(t, lines(box), "    a")
	row.StylePassing(false)
	assertLines(t, lines(box), "a    ")

	// passing leaves the styles set on the children as they are
	row.StylePassing(true)
	box.Render()
	if cell.GetStyle().GetAlign() != lipgloss.Left {
		t.Fatal("GetStyle returned a passed style")
	}
}

func TestHorizontalStylePassing(t *testing.T) {
	cell := NewCell(1, 1).SetContent("a")
	box := NewHorizontal(5, 1)
	column := box.NewColumn().AddCells(cell)
	box.AddColumns([]*Column{column})

	column.UseStyle(lipgloss.NewStyle().Align(lipgloss.Right))
	assertLines(t, lines(box), "    a")
	column.UseStyle(lipgloss.NewStyle().Align(lipgloss.Center))
	assertLines(t, lines(box), "  a  ")
}

// TestCellSizesAcrossARow: a fixed height on a cell sizes its row, a minimum height
// holds against a small share, and each cell takes its ratioY share of the row.
func TestCellSizesAcrossARow(t *testing.T) {
	tall, short := NewCell(1, 2).SetFixedHeight(4), NewCell(1, 1)
	low := NewCell(1, 1).SetMinHeight(3)
	rest := NewCell(1, 9)
	box := New(10, 20)
	box.AddRows([]*Row{
		box.NewRow().AddCells(tall, short),
		box.NewRow().AddCells(low),
		box.NewRow().AddCells(rest),
	})
	box.ForceRecalculate()

	for _, tc := range []struct {
		name string
		cell *Cell
		want int
	}{
		{"cell with the fixed height", tall, 4},
		{"cell of half the ratio next to it", short, 2},
		{"cell with the minimum height", low, 3},
		{"cell with the rest", rest, 13},
	} {
		if got := tc.cell.GetHeight(); got != tc.want {
			t.Errorf("%s is %d lines tall, want %d", tc.name, got, tc.want)
		}
	}
	if got := box.GetRow(0).GetHeight(); got != 4 {
		t.Errorf("row of the fixed-height cell is %d lines tall, want 4", got)
	}
}

func TestContentGeneratorGetsTheContentArea(t *testing.T) {
	var gotWidth, gotHeight, calls int
	cell := NewCell(1, 1).SetStyle(bordered().Padding(0, 1).Margin(1)).
		SetContentGenerator(func(w, h int) string {
			gotWidth, gotHeight = w, h
			calls++
			return ""
		})
	box := New(20, 10)
	box.AddRows([]*Row{box.NewRow().AddCells(cell)})
	box.Render()
	// 20x10 less a margin of 1, a border of 1 and a padding of 1 left and right
	if gotWidth != 14 || gotHeight != 6 {
		t.Fatalf("generator got %dx%d, want 14x6", gotWidth, gotHeight)
	}
	if got := cell.GetContentBounds(); got != image.Rect(3, 2, 17, 8) {
		t.Fatalf("content bounds %v, want (3,2)-(17,8)", got)
	}
	if w, h := cell.GetWidth(), cell.GetHeight(); w != 20 || h != 10 {
		t.Fatalf("cell is %dx%d, want 20x10", w, h)
	}
	box.Render()
	if calls != 2 {
		t.Fatalf("generator ran %d times in 2 renders, want 2", calls)
	}

	// a cell smaller than its frame asks for no content, and never for less than none
	box.SetWidth(3).SetHeight(3)
	box.Render()
	if gotWidth != 0 || gotHeight != 0 {
		t.Fatalf("generator got %dx%d in a cell smaller than its frame, want 0x0", gotWidth, gotHeight)
	}
	box.SetWidth(0)
	box.Render()
	if calls != 3 {
		t.Fatal("generator ran for a cell without columns")
	}
}

// TestRenderFollowsGeneratorContent: content built from outside state shows on the next
// render, and a render without changes returns the same text.
func TestRenderFollowsGeneratorContent(t *testing.T) {
	text := "one"
	box := New(5, 1)
	box.AddRows([]*Row{box.NewRow().AddCells(
		NewCell(1, 1).SetContentGenerator(func(_, _ int) string { return text }),
	)})
	assertLines(t, lines(box), "one  ")
	assertLines(t, lines(box), "one  ")
	text = "two"
	assertLines(t, lines(box), "two  ")
}

func TestCopiesLeaveTheGridAlone(t *testing.T) {
	box := New(6, 1)
	box.AddRows([]*Row{box.NewRow().AddCells(NewCell(1, 1).SetContent("a"), NewCell(1, 1).SetContent("b"))})
	box.Render()

	box.GetRowCopy(0).GetCell(0).SetContent("x")
	box.GetRowCellCopy(0, 1).SetContent("y").SetFixedWidth(1)
	box.GetRow(0).GetCellCopy(1).SetContent("z")
	assertLines(t, lines(box), "a  b  ")

	if c := box.GetRowCellCopy(0, 1); c.GetWidth() != 3 || c.GetContent() != "b" {
		t.Fatalf("copy is %d wide with content %q, want 3 and b", c.GetWidth(), c.GetContent())
	}
	for name, got := range map[string]any{
		"GetRow(9)":            box.GetRow(9),
		"GetRow(-1)":           box.GetRow(-1),
		"GetRowCopy(9)":        box.GetRowCopy(9),
		"GetRowCellCopy(9, 0)": box.GetRowCellCopy(9, 0),
		"GetRowCellCopy(0, 9)": box.GetRowCellCopy(0, 9),
		"GetCell(9)":           box.GetRow(0).GetCell(9),
		"GetCellCopy(-1)":      box.GetRow(0).GetCellCopy(-1),
	} {
		switch v := got.(type) {
		case *Row:
			if v != nil {
				t.Errorf("%s is not nil", name)
			}
		case *Cell:
			if v != nil {
				t.Errorf("%s is not nil", name)
			}
		}
	}

	// a copy put into the grid renders on its own
	replacement := box.GetRow(0).GetCellCopy(0).SetContent("q")
	box.GetRow(0).UpdateCellWithIndex(0, replacement).UpdateCellWithIndex(9, NewCell(1, 1))
	box.UpdateRow(9, box.NewRow())
	assertLines(t, lines(box), "q  b  ")
}

func TestCellIDs(t *testing.T) {
	box := New(6, 1)
	named := NewCell(1, 1).SetID("main")
	row := box.NewRow().AddCells(NewCell(1, 1), named, NewCell(1, 1))
	if got := row.GetCellWithID("main"); got != named {
		t.Fatal("GetCellWithID did not find the named cell")
	}
	if got := row.GetCellWithID("2"); got != row.GetCell(2) || got.GetID() != "2" {
		t.Fatal("an unnamed cell did not get its index as its ID")
	}
	if row.GetCellWithID("nope") != nil {
		t.Fatal("GetCellWithID found a cell that is not there")
	}
	if row.CellsLen() != 3 || box.AddRows([]*Row{row}).RowsLen() != 1 {
		t.Fatal("CellsLen or RowsLen is off")
	}
	if w, h := box.GetWidth(), box.GetHeight(); w != 6 || h != 1 {
		t.Fatalf("box is %dx%d, want 6x1", w, h)
	}
	box.Render()
	if w := row.GetWidth(); w != 6 {
		t.Fatalf("row is %d columns wide, want 6", w)
	}
}

func TestNewCellWithStyle(t *testing.T) {
	cell := NewCell(1, 1, WithStyle(bordered().Width(40).Height(40))).SetContent("ab")
	assertLines(t, renderAt(cell, 4, 3), "┌──┐", "│ab│", "└──┘")
	if cell.GetStyle().GetWidth() != 0 {
		t.Fatal("WithStyle kept the width key")
	}
}

func TestDeprecatedSetMinHeigth(t *testing.T) {
	box := NewHorizontal(1, 10)
	low := NewCell(1, 1).SetMinHeigth(7)
	box.AddColumns([]*Column{box.NewColumn().AddCells(low, NewCell(1, 1))})
	box.ForceRecalculate()
	if low.GetHeight() != 7 {
		t.Fatalf("cell is %d lines tall, want 7", low.GetHeight())
	}
}

// TestFramedRowWithoutRoomKeepsItsWidth: a bordered row too short for its cells closes
// its frame over the width of its cells.
func TestFramedRowWithoutRoomKeepsItsWidth(t *testing.T) {
	box := New(12, 4)
	box.AddRows([]*Row{
		box.NewRow().SetStyle(bordered()).SetFixedHeight(2).AddCells(NewCell(1, 1).SetContent("in row")),
		box.NewRow().AddCells(NewCell(1, 1).SetContent("second row")),
	})
	assertLines(t, lines(box),
		"┌──────────┐",
		"└──────────┘",
		"second row  ",
		"            ")

	box.GetRow(0).SetFixedHeight(1)
	assertLines(t, lines(box)[:1], "┌──────────┐")
}

// TestMinimumsHoldWhenFixedSizesTakeEverything: a minimum width is kept next to fixed
// cells that leave nothing, and the row is cut at its edge.
func TestMinimumsHoldWhenFixedSizesTakeEverything(t *testing.T) {
	low := NewCell(1, 1).SetContent("min").SetMinWidth(5)
	box := New(10, 1)
	box.AddRows([]*Row{box.NewRow().AddCells(low, NewCell(1, 1).SetContent("fixed fixed").SetFixedWidth(10))})
	assertLines(t, lines(box), "min  fixed")
	if low.GetWidth() != 5 {
		t.Fatalf("cell with a minimum width of 5 is %d wide", low.GetWidth())
	}
}

// TestNilElementsAreLeftOut: a nil row, column or cell, such as a getter returns for an
// index that is not there, never gets into a grid.
func TestNilElementsAreLeftOut(t *testing.T) {
	box := New(4, 1)
	row := box.NewRow().AddCells(nil, NewCell(1, 1).SetContent("a"), nil)
	box.AddRows([]*Row{nil, row}).UpdateRow(0, box.GetRowCopy(9))
	row.UpdateCellWithIndex(0, row.GetCellCopy(9)).AddCells(row.GetCellCopy(9))
	if box.RowsLen() != 1 || row.CellsLen() != 1 {
		t.Fatalf("box has %d rows and its row %d cells, want 1 and 1", box.RowsLen(), row.CellsLen())
	}
	assertLines(t, lines(box), "a   ")
	box.SetRows([]*Row{nil})
	assertLines(t, lines(box), "    ")

	horizontal := NewHorizontal(4, 1)
	column := horizontal.NewColumn().AddCells(nil, NewCell(1, 1).SetContent("a"))
	horizontal.AddColumns([]*Column{nil, column}).UpdateColumn(0, horizontal.GetColumnCopy(9))
	column.UpdateCellWithIndex(0, nil)
	if horizontal.ColumnsLen() != 1 || column.CellsLen() != 1 {
		t.Fatalf("box has %d columns and its column %d cells, want 1 and 1", horizontal.ColumnsLen(), column.CellsLen())
	}
	assertLines(t, lines(horizontal), "a   ")
	horizontal.SetColumns([]*Column{nil})
	assertLines(t, lines(horizontal), "    ")
}

// TestSetRowsKeepsItsOwnList: a change to the slice given to SetRows does not reach the
// box.
func TestSetRowsKeepsItsOwnList(t *testing.T) {
	box := New(4, 2)
	rows := []*Row{
		box.NewRow().AddCells(NewCell(1, 1).SetContent("a")),
		box.NewRow().AddCells(NewCell(1, 1).SetContent("b")),
	}
	box.SetRows(rows)
	rows[0], rows[1] = rows[1], rows[0]
	assertLines(t, lines(box), "a   ", "b   ")
	box.SetRows(rows[:1])
	assertLines(t, lines(box), "b   ", "    ")
}

// TestGeneratorReadsThePositionOfItsRender: a content generator that asks for its cell's
// bounds gets the bounds of the render it is called for, from the first one on.
func TestGeneratorReadsThePositionOfItsRender(t *testing.T) {
	var second *Cell
	second = NewCell(1, 1).SetContentGenerator(func(_, _ int) string { return second.GetBounds().String() })
	box := New(40, 1)
	box.AddRows([]*Row{box.NewRow().AddCells(NewCell(1, 1), second)})
	assertLines(t, lines(box), "                    (20,0)-(40,1)       ")
	box.SetWidth(30)
	assertLines(t, lines(box), "               (15,0)-(30,1)  ")

	var cell *Cell
	cell = NewCell(1, 1).SetContentGenerator(func(_, _ int) string { return cell.GetBounds().String() })
	horizontal := NewHorizontal(40, 1)
	horizontal.AddColumns([]*Column{horizontal.NewColumn().AddCells(NewCell(1, 1)), horizontal.NewColumn().AddCells(cell)})
	assertLines(t, lines(horizontal), "                    (20,0)-(40,1)       ")
}

// TestRowFrameWiderThanTheBoxIsCut: a row whose frame alone is wider than the box is cut
// at the box's edge and leaves the rows that fit where they are.
func TestRowFrameWiderThanTheBoxIsCut(t *testing.T) {
	for _, boxStyle := range []lipgloss.Style{lipgloss.NewStyle(), lipgloss.NewStyle().Bold(true)} {
		box := New(8, 6).SetRowAlign(lipgloss.Right).SetStyle(boxStyle)
		wide := box.NewRow().SetStyle(bordered().Padding(0, 4)).AddCells(NewCell(1, 1))
		top := NewCell(1, 1).SetContent("AAAAAAAA\nAAAAAAAA")
		box.AddRows([]*Row{box.NewRow().AddCells(top), wide})
		rendered := box.Render()
		assertSize(t, rendered, 8, 6)
		if got := strings.Split(ansi.Strip(rendered), "\n"); got[0] != "AAAAAAAA" || got[1] != "AAAAAAAA" {
			t.Fatalf("the row that fits renders %q and %q", got[0], got[1])
		}
		if got := wide.GetBounds(); got.Dx() != 8 {
			t.Fatalf("the wide row is %d columns wide, want the 8 of the box", got.Dx())
		}
		if box.CellAt(3, 0) != top {
			t.Fatal("CellAt does not find the cell of the row that fits")
		}
	}
}

// TestCellInTwoRowsKeepsTheBoxItsSize: a cell added to two rows has one size, and the box
// is its size all the same.
func TestCellInTwoRowsKeepsTheBoxItsSize(t *testing.T) {
	shared := NewCell(1, 1).SetStyle(bordered()).SetContent("shared")
	box := New(30, 8)
	box.AddRows([]*Row{
		box.NewRow().AddCells(shared, NewCell(3, 1).SetContent("wide")),
		box.NewRow().AddCells(NewCell(1, 1).SetContent("a"), shared, NewCell(1, 1).SetContent("b")),
	})
	assertSize(t, box.Render(), 30, 8)
	assertSize(t, box.Render(), 30, 8)

	horizontal := NewHorizontal(30, 8)
	horizontal.AddColumns([]*Column{
		horizontal.NewColumn().AddCells(shared, NewCell(1, 3)).SetFixedWidth(6),
		horizontal.NewColumn().AddCells(NewCell(1, 1), shared),
	})
	assertSize(t, horizontal.Render(), 30, 8)
}
