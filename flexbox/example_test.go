package flexbox_test

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/tmctec/stickers/v2/flexbox"
)

var border = lipgloss.NewStyle().Border(lipgloss.NormalBorder())

// A sidebar of a fixed width next to a main cell that takes the rest, over a footer of a
// fixed height.
func Example() {
	box := flexbox.New(30, 7)
	box.AddRows([]*flexbox.Row{
		box.NewRow().AddCells(
			flexbox.NewCell(1, 1).SetStyle(border).SetContent("sidebar").SetFixedWidth(11),
			flexbox.NewCell(1, 1).SetStyle(border).SetContent("main"),
		),
		box.NewRow().AddCells(
			flexbox.NewCell(1, 1).SetStyle(border).SetContent("footer"),
		).SetFixedHeight(3),
	})
	fmt.Println(box.Render())
	// Output:
	// ┌─────────┐┌─────────────────┐
	// │sidebar  ││main             │
	// │         ││                 │
	// └─────────┘└─────────────────┘
	// ┌────────────────────────────┐
	// │footer                      │
	// └────────────────────────────┘
}

// Ratios split a row's width between its cells and the box's height between its rows.
func ExampleNewCell() {
	box := flexbox.New(24, 6)
	box.AddRows([]*flexbox.Row{
		box.NewRow().AddCells(
			flexbox.NewCell(1, 1, flexbox.WithStyle(border)).SetContent("1:1"),
			flexbox.NewCell(2, 1, flexbox.WithStyle(border)).SetContent("2:1"),
		),
		box.NewRow().AddCells(
			flexbox.NewCell(1, 1, flexbox.WithStyle(border)).SetContent("1:1"),
		),
	})
	fmt.Println(box.Render())
	// Output:
	// ┌──────┐┌──────────────┐
	// │1:1   ││2:1           │
	// └──────┘└──────────────┘
	// ┌──────────────────────┐
	// │1:1                   │
	// └──────────────────────┘
}

// A content generator receives the area inside the cell's border, so the content can
// follow the cell's size.
func ExampleCell_SetContentGenerator() {
	box := flexbox.New(16, 4)
	box.AddRows([]*flexbox.Row{
		box.NewRow().AddCells(
			flexbox.NewCell(1, 1).SetStyle(border).SetContentGenerator(func(width, height int) string {
				return fmt.Sprintf("%d by %d", width, height)
			}),
		),
	})
	fmt.Println(box.Render())

	box.SetWidth(12).SetHeight(3)
	fmt.Println(box.Render())
	// Output:
	// ┌──────────────┐
	// │14 by 2       │
	// │              │
	// └──────────────┘
	// ┌──────────┐
	// │10 by 1   │
	// └──────────┘
}

// A box nests in a cell through a content generator that sizes the inner box to the
// cell and renders it.
func ExampleCell_SetContentGenerator_nested() {
	inner := flexbox.NewHorizontal(0, 0)
	inner.AddColumns([]*flexbox.Column{
		inner.NewColumn().AddCells(flexbox.NewCell(1, 1).SetStyle(border).SetContent("a")),
		inner.NewColumn().AddCells(flexbox.NewCell(1, 1).SetStyle(border).SetContent("b")),
	})

	outer := flexbox.New(20, 5)
	outer.AddRows([]*flexbox.Row{
		outer.NewRow().AddCells(
			flexbox.NewCell(1, 1).SetStyle(border).SetContentGenerator(func(width, height int) string {
				return inner.SetWidth(width).SetHeight(height).Render()
			}),
		),
	})
	fmt.Println(outer.Render())
	// Output:
	// ┌──────────────────┐
	// │┌───────┐┌───────┐│
	// ││a      ││b      ││
	// │└───────┘└───────┘│
	// └──────────────────┘
}

// Rows narrower than the widest row sit where SetRowAlign puts them.
func ExampleFlexBox_SetRowAlign() {
	box := flexbox.New(20, 6).SetRowAlign(lipgloss.Right)
	box.AddRows([]*flexbox.Row{
		box.NewRow().AddCells(flexbox.NewCell(1, 1).SetStyle(border).SetContent("full width")),
		box.NewRow().AddCells(flexbox.NewCell(1, 1).SetStyle(border).SetContent("fixed").SetFixedWidth(10)),
	})
	fmt.Println(box.Render())
	// Output:
	// ┌──────────────────┐
	// │full width        │
	// └──────────────────┘
	//           ┌────────┐
	//           │fixed   │
	//           └────────┘
}

// A style set with UseStyle is passed down: from the box to its rows, and from a row
// that passes styles to its cells.
func ExampleRow_UseStyle() {
	box := flexbox.New(18, 3)
	box.AddRows([]*flexbox.Row{
		box.NewRow().UseStyle(lipgloss.NewStyle().Align(lipgloss.Center)).AddCells(
			flexbox.NewCell(1, 1).SetStyle(border).SetContent("a"),
			flexbox.NewCell(1, 1).SetStyle(border).SetContent("b"),
		),
	})
	fmt.Println(box.Render())
	// Output:
	// ┌───────┐┌───────┐
	// │   a   ││   b   │
	// └───────┘└───────┘
}

// A HorizontalFlexBox lines up columns. Here a column of a fixed width holds a menu and
// the other column splits its height between two cells.
func ExampleHorizontalFlexBox() {
	box := flexbox.NewHorizontal(26, 7)
	box.AddColumns([]*flexbox.Column{
		box.NewColumn().SetFixedWidth(8).AddCells(
			flexbox.NewCell(1, 1).SetStyle(border).SetContent("menu"),
		),
		box.NewColumn().AddCells(
			flexbox.NewCell(1, 1).SetStyle(border).SetContent("title").SetFixedHeight(3),
			flexbox.NewCell(1, 1).SetStyle(border).SetContent("body"),
		),
	})
	fmt.Println(box.Render())
	// Output:
	// ┌──────┐┌────────────────┐
	// │menu  ││title           │
	// │      │└────────────────┘
	// │      │┌────────────────┐
	// │      ││body            │
	// │      ││                │
	// └──────┘└────────────────┘
}

// The sizes of rows and cells can be read after ForceRecalculate or a render.
func ExampleFlexBox_ForceRecalculate() {
	box := flexbox.New(30, 10)
	box.AddRows([]*flexbox.Row{
		box.NewRow().AddCells(flexbox.NewCell(1, 1), flexbox.NewCell(2, 1)),
		box.NewRow().AddCells(flexbox.NewCell(1, 4)),
	})
	box.ForceRecalculate()

	for i := 0; i < box.RowsLen(); i++ {
		row := box.GetRow(i)
		var sizes []string
		for j := 0; j < row.CellsLen(); j++ {
			cell := row.GetCell(j)
			sizes = append(sizes, fmt.Sprintf("%dx%d", cell.GetWidth(), cell.GetHeight()))
		}
		fmt.Printf("row %d: %s\n", i, strings.Join(sizes, " "))
	}
	// Output:
	// row 0: 10x2 20x2
	// row 1: 30x8
}

// CellAt turns a point, such as the X and Y of a mouse click, into the cell that shows
// there. The bounds of the cell tell where it is and where its content starts.
func ExampleFlexBox_CellAt() {
	box := flexbox.New(20, 4)
	box.AddRows([]*flexbox.Row{
		box.NewRow().AddCells(
			flexbox.NewCell(1, 1).SetID("menu").SetStyle(border),
			flexbox.NewCell(1, 1).SetID("main").SetStyle(border),
		),
	})
	box.Render()

	if cell := box.CellAt(14, 2); cell != nil {
		fmt.Println("clicked:", cell.GetID())
		fmt.Println("cell:", cell.GetBounds())
		fmt.Println("content:", cell.GetContentBounds())
	}
	fmt.Println("outside the box:", box.CellAt(40, 2) == nil)
	// Output:
	// clicked: main
	// cell: (10,0)-(20,4)
	// content: (11,1)-(19,3)
	// outside the box: true
}
