package flexbox

import (
	"image"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// crop returns the text that rendered shows inside area, without escape sequences.
func crop(rendered string, area image.Rectangle) []string {
	lines := strings.Split(rendered, "\n")
	var cropped []string
	for y := area.Min.Y; y < area.Max.Y; y++ {
		cropped = append(cropped, ansi.Strip(ansi.Cut(lines[y], area.Min.X, area.Max.X)))
	}
	return cropped
}

// stripped returns the lines of a block without escape sequences.
func stripped(lines []string) []string {
	plain := make([]string, len(lines))
	for i, line := range lines {
		plain[i] = ansi.Strip(line)
	}
	return plain
}

func assertBounds(t *testing.T, name string, got, want image.Rectangle) {
	t.Helper()
	if got != want {
		t.Errorf("%s is at %v, want %v", name, got, want)
	}
}

func TestBoundsInAFlexBox(t *testing.T) {
	side, main := NewCell(1, 1).SetStyle(bordered()).SetFixedWidth(8), NewCell(1, 1).SetStyle(bordered().Padding(0, 1))
	tall, short := NewCell(1, 2), NewCell(1, 1)
	box := New(30, 10).SetStyle(bordered().Padding(0, 1))
	top := box.NewRow().AddCells(side, main).SetFixedHeight(4)
	bottom := box.NewRow().AddCells(tall, short)
	box.AddRows([]*Row{top, bottom})
	box.Render()

	// the box's border and padding leave the rows 26 columns from (2, 1)
	assertBounds(t, "top row", top.GetBounds(), image.Rect(2, 1, 28, 5))
	assertBounds(t, "bottom row", bottom.GetBounds(), image.Rect(2, 5, 28, 9))
	assertBounds(t, "side cell", side.GetBounds(), image.Rect(2, 1, 10, 5))
	assertBounds(t, "main cell", main.GetBounds(), image.Rect(10, 1, 28, 5))
	assertBounds(t, "tall cell", tall.GetBounds(), image.Rect(2, 5, 15, 9))
	assertBounds(t, "short cell", short.GetBounds(), image.Rect(15, 5, 28, 7))

	// content starts inside the border and the padding
	assertBounds(t, "side content", side.GetContentBounds(), image.Rect(3, 2, 9, 4))
	assertBounds(t, "main content", main.GetContentBounds(), image.Rect(12, 2, 26, 4))
	assertBounds(t, "tall content", tall.GetContentBounds(), tall.GetBounds())

	for _, tc := range []struct {
		name string
		x, y int
		want *Cell
	}{
		{"the box's border", 0, 0, nil},
		{"the box's padding", 1, 3, nil},
		{"the side cell's corner", 2, 1, side},
		{"the side cell's last column", 9, 4, side},
		{"the main cell", 10, 1, main},
		{"the main cell's last column", 27, 4, main},
		{"the box's right padding", 28, 3, nil},
		{"the tall cell", 14, 8, tall},
		{"the short cell", 15, 6, short},
		{"the blank under the short cell", 15, 7, nil},
		{"the box's bottom border", 5, 9, nil},
		{"left of the box", -1, 3, nil},
		{"under the box", 5, 10, nil},
	} {
		if got := box.CellAt(tc.x, tc.y); got != tc.want {
			t.Errorf("CellAt(%d, %d), %s, found the wrong cell", tc.x, tc.y, tc.name)
		}
	}
}

// TestBoundsOfAlignedAndCutRows: a row narrower than the box sits where the alignment
// puts it, and a cell cut at the row's edge is found only where it shows.
func TestBoundsOfAlignedAndCutRows(t *testing.T) {
	full, narrow, wide := NewCell(1, 1), NewCell(1, 1).SetFixedWidth(4), NewCell(1, 1).SetFixedWidth(30)
	box := New(10, 3).SetRowAlign(lipgloss.Right)
	box.AddRows([]*Row{
		box.NewRow().AddCells(full),
		box.NewRow().AddCells(narrow),
		box.NewRow().AddCells(wide),
	})
	box.ForceRecalculate()

	assertBounds(t, "full cell", full.GetBounds(), image.Rect(0, 0, 10, 1))
	assertBounds(t, "narrow cell", narrow.GetBounds(), image.Rect(6, 1, 10, 2))
	assertBounds(t, "narrow row", box.GetRow(1).GetBounds(), image.Rect(6, 1, 10, 2))
	assertBounds(t, "cut cell", wide.GetBounds(), image.Rect(0, 2, 30, 3))
	assertBounds(t, "cut row", box.GetRow(2).GetBounds(), image.Rect(0, 2, 10, 3))
	if box.CellAt(5, 1) != nil || box.CellAt(6, 1) != narrow {
		t.Error("the narrow row is not found at its place on the right")
	}
	if box.CellAt(9, 2) != wide || box.CellAt(10, 2) != nil {
		t.Error("the cut cell is not found exactly where it shows")
	}

	box.SetRowAlign(lipgloss.Center)
	box.ForceRecalculate()
	assertBounds(t, "centered cell", narrow.GetBounds(), image.Rect(3, 1, 7, 2))

	// a box with a style moves rows that leave room as a group
	box.SetRows([]*Row{box.GetRow(1)}).SetStyle(lipgloss.NewStyle().Align(lipgloss.Right, lipgloss.Bottom))
	box.ForceRecalculate()
	assertBounds(t, "cell in a box aligned right and bottom", narrow.GetBounds(), image.Rect(6, 0, 10, 3))
}

func TestBoundsInAHorizontalFlexBox(t *testing.T) {
	menu := NewCell(1, 1)
	title, body := NewCell(2, 1).SetFixedHeight(2), NewCell(1, 1)
	box := NewHorizontal(30, 8).SetStyle(bordered())
	left := box.NewColumn().AddCells(menu).SetFixedWidth(8)
	right := box.NewColumn().AddCells(title, body).SetStyle(lipgloss.NewStyle().Padding(1, 2))
	box.AddColumns([]*Column{left, right})
	box.Render()

	assertBounds(t, "left column", left.GetBounds(), image.Rect(1, 1, 9, 7))
	assertBounds(t, "right column", right.GetBounds(), image.Rect(9, 1, 29, 7))
	assertBounds(t, "menu cell", menu.GetBounds(), image.Rect(1, 1, 9, 7))
	assertBounds(t, "title cell", title.GetBounds(), image.Rect(11, 2, 27, 4))
	assertBounds(t, "body cell", body.GetBounds(), image.Rect(11, 4, 19, 6))

	for _, tc := range []struct {
		name string
		x, y int
		want *Cell
	}{
		{"the box's border", 0, 3, nil},
		{"the menu", 8, 6, menu},
		{"the right column's padding", 9, 3, nil},
		{"the title", 26, 3, title},
		{"the body", 18, 5, body},
		{"beside the body, which is half the column wide", 19, 5, nil},
		{"the right column's bottom padding", 12, 6, nil},
	} {
		if got := box.CellAt(tc.x, tc.y); got != tc.want {
			t.Errorf("CellAt(%d, %d), %s, found the wrong cell", tc.x, tc.y, tc.name)
		}
	}

	// columns of a fixed width sit where the column alignment puts them
	narrow := NewHorizontal(20, 2).LockColumnWidth(4).SetColumnAlign(lipgloss.Right)
	a, b := NewCell(1, 1), NewCell(1, 1)
	narrow.AddColumns([]*Column{narrow.NewColumn().AddCells(a), narrow.NewColumn().AddCells(b)})
	narrow.ForceRecalculate()
	assertBounds(t, "first column on the right", a.GetBounds(), image.Rect(12, 0, 16, 2))
	assertBounds(t, "second column on the right", b.GetBounds(), image.Rect(16, 0, 20, 2))
}

// TestBoundsMatchTheRender: in random grids, with every style and alignment, the text a
// box renders inside a cell's bounds is the cell's own block. That ties the positions
// to what is on screen.
func TestBoundsMatchTheRender(t *testing.T) {
	label := 0
	labeled := func(rng *rand.Rand) *Cell {
		label++
		text := "#" + strconv.Itoa(label)
		return randomCell(rng).SetContent(text + " " + text + "\n" + text)
	}
	check := func(t *testing.T, seed int64, rendered string, visible image.Rectangle, cell *Cell) {
		t.Helper()
		bounds := cell.GetBounds()
		if bounds.Empty() || !bounds.In(visible) {
			return
		}
		got, want := crop(rendered, bounds), stripped(cell.out.lines)
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Fatalf("seed %d: the cell at %v shows\n%s\nand its block is\n%s\nin\n%s",
				seed, bounds, strings.Join(got, "\n"), strings.Join(want, "\n"), ansi.Strip(rendered))
		}
	}
	aligned := append([]lipgloss.Style{
		lipgloss.NewStyle().Align(lipgloss.Right, lipgloss.Bottom),
		lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Border(lipgloss.NormalBorder()),
		lipgloss.NewStyle().Margin(1, 2).Padding(1),
	}, referenceContainerStyles...)

	for seed := int64(0); seed < 200; seed++ {
		rng := rand.New(rand.NewSource(seed))

		box := New(40+rng.Intn(60), 20+rng.Intn(20)).SetRowAlign(pick(rng, referenceAligns))
		if rng.Intn(2) == 0 {
			box.SetStyle(pick(rng, aligned))
		}
		for r, n := 0, 1+rng.Intn(4); r < n; r++ {
			row := box.NewRow()
			for c, m := 0, 1+rng.Intn(4); c < m; c++ {
				cell := labeled(rng)
				if rng.Intn(3) == 0 {
					cell.SetFixedWidth(3 + rng.Intn(12))
				}
				row.AddCells(cell)
			}
			if rng.Intn(3) == 0 {
				row.SetStyle(pick(rng, aligned))
			}
			if rng.Intn(4) == 0 {
				row.SetFixedHeight(1 + rng.Intn(4))
			}
			box.AddRows([]*Row{row})
		}
		rendered := box.Render()
		for _, row := range box.rows {
			for _, cell := range row.cells {
				check(t, seed, rendered, row.contentBounds().Intersect(box.contentBounds()), cell)
			}
		}

		horizontal := NewHorizontal(40+rng.Intn(60), 20+rng.Intn(20)).SetColumnAlign(pick(rng, referenceAligns))
		if rng.Intn(2) == 0 {
			horizontal.SetStyle(pick(rng, aligned))
		}
		if rng.Intn(3) == 0 {
			horizontal.LockColumnWidth(6 + rng.Intn(10))
		}
		for c, n := 0, 1+rng.Intn(4); c < n; c++ {
			column := horizontal.NewColumn()
			for r, m := 0, 1+rng.Intn(3); r < m; r++ {
				cell := labeled(rng)
				if rng.Intn(3) == 0 {
					cell.SetFixedHeight(1 + rng.Intn(4))
				}
				column.AddCells(cell)
			}
			if rng.Intn(3) == 0 {
				column.SetStyle(pick(rng, aligned))
			}
			horizontal.AddColumns([]*Column{column})
		}
		rendered = horizontal.Render()
		for _, column := range horizontal.columns {
			for _, cell := range column.cells {
				check(t, seed, rendered, column.contentBounds().Intersect(horizontal.contentBounds()), cell)
			}
		}
	}
}

// TestCellAtFindsEveryCellWhereItShows: every point of a random grid belongs to the one
// cell whose text shows there, or to none.
func TestCellAtFindsEveryCellWhereItShows(t *testing.T) {
	for seed := int64(0); seed < 40; seed++ {
		rng := rand.New(rand.NewSource(seed))
		box := New(20+rng.Intn(30), 8+rng.Intn(10)).SetRowAlign(pick(rng, referenceAligns))
		if rng.Intn(2) == 0 {
			box.SetStyle(pick(rng, referenceContainerStyles))
		}
		for r, n := 0, 1+rng.Intn(3); r < n; r++ {
			row := box.NewRow()
			for c, m := 0, 1+rng.Intn(4); c < m; c++ {
				row.AddCells(randomCell(rng))
			}
			box.AddRows([]*Row{row})
		}
		box.Render()

		for y := -1; y <= box.GetHeight(); y++ {
			for x := -1; x <= box.GetWidth(); x++ {
				var want *Cell
				for _, row := range box.rows {
					for _, cell := range row.cells {
						shown := cell.GetBounds().Intersect(row.contentBounds()).Intersect(box.contentBounds())
						if image.Pt(x, y).In(shown) {
							if want != nil {
								t.Fatalf("seed %d: two cells show at (%d, %d)", seed, x, y)
							}
							want = cell
						}
					}
				}
				if got := box.CellAt(x, y); got != want {
					t.Fatalf("seed %d: CellAt(%d, %d) found the wrong cell", seed, x, y)
				}
			}
		}
	}
}
