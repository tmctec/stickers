package flexbox

import (
	"strconv"
	"testing"

	"charm.land/lipgloss/v2"
)

// benchBox builds a full-screen dashboard: a fixed header, three ratio rows of bordered
// cells and a fixed footer.
func benchBox(width, height int) *FlexBox {
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	box := New(width, height)
	rows := []*Row{
		box.NewRow().AddCells(NewCell(1, 1).SetContent("header")).SetFixedHeight(1),
	}
	for r := 0; r < 3; r++ {
		row := box.NewRow()
		for c := 0; c < 5; c++ {
			row.AddCells(NewCell(c+1, r+1).SetStyle(border).
				SetContent("cell " + strconv.Itoa(r) + ":" + strconv.Itoa(c) + " lorem ipsum dolor sit amet"))
		}
		rows = append(rows, row)
	}
	rows = append(rows, box.NewRow().AddCells(NewCell(1, 1).SetContent("footer")).SetFixedHeight(1))
	return box.AddRows(rows)
}

// BenchmarkFlexBoxRenderUnchanged is the steady state of a Bubble Tea app: View runs after
// every message while the grid stays the same.
func BenchmarkFlexBoxRenderUnchanged(b *testing.B) {
	box := benchBox(160, 48)
	box.Render()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		box.Render()
	}
}

// BenchmarkFlexBoxRenderOneCellChanged changes the content of one cell between renders.
func BenchmarkFlexBoxRenderOneCellChanged(b *testing.B) {
	box := benchBox(160, 48)
	n := 0
	box.GetRow(2).GetCell(2).SetContentGenerator(func(_, _ int) string { return "tick " + strconv.Itoa(n) })
	box.Render()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n = i
		box.Render()
	}
}

// BenchmarkFlexBoxRenderResized changes the box size between renders, so every cell is
// laid out and rendered again.
func BenchmarkFlexBoxRenderResized(b *testing.B) {
	box := benchBox(160, 48)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		box.SetWidth(150 + i%10).SetHeight(40 + i%8)
		box.Render()
	}
}

// BenchmarkHorizontalFlexBoxRenderUnchanged renders an unchanged grid of columns.
func BenchmarkHorizontalFlexBoxRenderUnchanged(b *testing.B) {
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	box := NewHorizontal(160, 48)
	var columns []*Column
	for c := 0; c < 4; c++ {
		column := box.NewColumn()
		for r := 0; r < 4; r++ {
			column.AddCells(NewCell(c+1, r+1).SetStyle(border).SetContent("cell " + strconv.Itoa(c) + ":" + strconv.Itoa(r)))
		}
		columns = append(columns, column)
	}
	box.AddColumns(columns)
	box.Render()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		box.Render()
	}
}
