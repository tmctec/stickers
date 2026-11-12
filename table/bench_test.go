package table

import (
	"strconv"
	"testing"

	"github.com/tmctec/stickers/v2/flexbox"
)

// benchTable builds a table of n rows with an int, a string and a float column.
func benchTable(n int) *Table {
	tb := NewTable(160, 48, []string{"ID", "Name", "Score", "City", "Notes"})
	if _, err := tb.SetTypes(0, "", 0.0, "", ""); err != nil {
		panic(err)
	}
	tb.SetRatio([]int{1, 3, 1, 2, 4}).SetMinWidth([]int{6, 12, 8, 10, 16})
	rows := make([][]any, 0, n)
	for i := 0; i < n; i++ {
		// a multiplicative hash spreads the keys so the sort has work to do
		k := (i * 2654435761) % 1000003
		rows = append(rows, []any{k, "name-" + strconv.Itoa(k), float64(k) / 7, "city-" + strconv.Itoa(k%97), "some longer note text"})
	}
	return tb.MustAddRows(rows)
}

// BenchmarkTableRenderUnchanged is the steady state: View runs while nothing changes.
func BenchmarkTableRenderUnchanged(b *testing.B) {
	tb := benchTable(1000)
	tb.Render()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.Render()
	}
}

// BenchmarkTableInCellRenderUnchanged renders an unchanged table through the content
// generator of a flexbox cell, which sizes the table on every render.
func BenchmarkTableInCellRenderUnchanged(b *testing.B) {
	tb := benchTable(1000)
	box := flexbox.New(160, 48)
	box.AddRows([]*flexbox.Row{box.NewRow().AddCells(
		flexbox.NewCell(1, 1).SetContent("sidebar").SetFixedWidth(20),
		flexbox.NewCell(1, 1).SetContentGenerator(func(width, height int) string {
			return tb.SetWidth(width).SetHeight(height).Render()
		}),
	)})
	box.Render()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		box.Render()
	}
}

// BenchmarkTableCursorMove moves the cursor one row and renders, without scrolling.
func BenchmarkTableCursorMove(b *testing.B) {
	tb := benchTable(1000)
	tb.Render()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			tb.CursorDown()
		} else {
			tb.CursorUp()
		}
		tb.Render()
	}
}

// BenchmarkTableScroll moves the cursor down through the rows, scrolling on every move
// once the cursor reaches the bottom of the viewport.
func BenchmarkTableScroll(b *testing.B) {
	tb := benchTable(100000)
	tb.Render()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.CursorDown()
		tb.Render()
	}
}

// BenchmarkTableCursorMoveFiltered moves the cursor in a filtered table of 100000 rows.
func BenchmarkTableCursorMoveFiltered(b *testing.B) {
	tb := benchTable(100000)
	tb.SetFilter(1, "name-1")
	tb.Render()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			tb.CursorDown()
		} else {
			tb.CursorUp()
		}
		tb.Render()
	}
}

// BenchmarkTableSort sorts 10000 rows, alternating direction and column.
func BenchmarkTableSort(b *testing.B) {
	tb := benchTable(10000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			tb.OrderByAsc(i % 3)
		} else {
			tb.OrderByDesc(i % 3)
		}
	}
}

// BenchmarkTableFilter applies a filter to 100000 rows and renders.
func BenchmarkTableFilter(b *testing.B) {
	tb := benchTable(100000)
	filters := []string{"name-1", "name-12", "name-123"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.SetFilter(1, filters[i%len(filters)])
		tb.Render()
	}
}
