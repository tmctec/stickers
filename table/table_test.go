package table

import (
	"errors"
	"math"
	"math/rand"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

// plain returns the lines of a render without escape sequences.
func plain(rendered string) []string {
	return strings.Split(ansiRe.ReplaceAllString(rendered, ""), "\n")
}

// people returns a table of three typed columns with five rows.
func people(width, height int) *Table {
	tb := NewTable(width, height, []string{"ID", "Name", "Score"})
	if _, err := tb.SetTypes(0, "", 0.0); err != nil {
		panic(err)
	}
	return tb.MustAddRows([][]any{
		{20, "Linus", 8.0},
		{3, "ada", 9.5},
		{100, "Grace", 9.5},
		{7, "Édith", 7.25},
		{51, "Barbara", 1234567.0},
	})
}

// column returns the values of the column at index over every row, top to bottom, read
// with the cursor.
func column(tb *Table, index int) []string {
	x, y := tb.GetCursorLocation()
	for ; x > index; x-- {
		tb.CursorLeft()
	}
	for ; x < index; x++ {
		tb.CursorRight()
	}
	for ; y > 0; y-- {
		tb.CursorUp()
	}
	var values []string
	for {
		if value := tb.GetCursorValue(); value != "" || tb.GetCursorRow() != nil {
			values = append(values, value)
		}
		_, before := tb.GetCursorLocation()
		tb.CursorDown()
		if _, after := tb.GetCursorLocation(); after == before {
			return values
		}
	}
}

func assertColumn(t *testing.T, tb *Table, index int, want ...string) {
	t.Helper()
	if got := column(tb, index); !slices.Equal(got, want) {
		t.Fatalf("column %d reads %q, want %q", index, got, want)
	}
}

func assertRendered(t *testing.T, tb *Table, want ...string) {
	t.Helper()
	if got := plain(tb.Render()); !slices.Equal(got, want) {
		t.Fatalf("rendered\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestRenderShowsHeaderRowsAndFooter(t *testing.T) {
	tb := people(30, 6).SetRatio([]int{1, 2, 2})
	assertRendered(t, tb,
		"ID    Name        Score       ",
		"20    Linus       8           ",
		"3     ada         9.5         ",
		"100   Grace       9.5         ",
		"7     Édith       7.25        ",
		"                    0:0 / 30:4")
	tb.CursorDown().CursorDown().CursorDown().CursorDown().CursorRight()
	assertRendered(t, tb,
		"ID    Name        Score       ",
		"3     ada         9.5         ",
		"100   Grace       9.5         ",
		"7     Édith       7.25        ",
		"51    Barbara     1234567     ",
		"                    1:4 / 30:4")
	if got := tb.GetCursorValue(); got != "Barbara" {
		t.Fatalf("value under the cursor is %q, want Barbara", got)
	}
}

// TestRenderIsExactlyTheTableSize: the header, then the footer, then the rows give way
// as the table gets shorter, and the footer never wraps.
func TestRenderIsExactlyTheTableSize(t *testing.T) {
	for height := 0; height <= 9; height++ {
		for _, width := range []int{1, 5, 12, 40} {
			tb := people(width, height).SetFilter(0, "0").CursorDown()
			rendered := tb.Render()
			if height == 0 {
				if rendered != "" {
					t.Fatalf("%dx0 rendered %q, want nothing", width, rendered)
				}
				continue
			}
			got := strings.Split(rendered, "\n")
			if len(got) != height {
				t.Fatalf("%dx%d rendered %d lines:\n%s", width, height, len(got), rendered)
			}
			for i, line := range got {
				if w := lipgloss.Width(line); w != width {
					t.Fatalf("%dx%d line %d is %d columns wide:\n%s", width, height, i, w, rendered)
				}
			}
		}
	}
}

func TestRenderWithoutColumnsOrRows(t *testing.T) {
	assertRendered(t, NewTable(10, 3, nil), "          ", "          ", "0:0 / 10:1")
	assertRendered(t, NewTable(10, 3, []string{"A"}), "A         ", "          ", "0:0 / 10:1")
}

func TestSortComparesByColumnType(t *testing.T) {
	tb := people(40, 8)
	if column, _ := tb.GetOrder(); column != -1 {
		t.Fatalf("a new table is sorted by column %d", column)
	}

	tb.OrderByAsc(0)
	assertColumn(t, tb, 0, "3", "7", "20", "51", "100")
	tb.OrderByDesc(0)
	assertColumn(t, tb, 0, "100", "51", "20", "7", "3")
	if column, order := tb.GetOrder(); column != 0 || order != SortingOrderDescending {
		t.Fatalf("GetOrder = %d, %d", column, order)
	}

	tb.OrderByAsc(1)
	assertColumn(t, tb, 1, "Barbara", "Grace", "Linus", "ada", "Édith")

	// rows that compare equal keep their order, in both directions
	tb.OrderByAsc(0).OrderByAsc(2)
	assertColumn(t, tb, 0, "7", "20", "3", "100", "51")
	tb.OrderByAsc(0).OrderByDesc(2)
	assertColumn(t, tb, 0, "51", "3", "100", "20", "7")

	// an index out of range changes nothing
	tb.OrderByAsc(-1).OrderByDesc(3)
	if column, order := tb.GetOrder(); column != 2 || order != SortingOrderDescending {
		t.Fatalf("GetOrder = %d, %d after sorting by a column that is not there", column, order)
	}
}

func TestSortHandlesEveryColumnType(t *testing.T) {
	tb := NewTable(80, 6, []string{"a", "b", "c", "d", "e", "f", "g", "h"})
	if _, err := tb.SetTypes("", 0, int8(0), int16(0), int32(0), int64(0), float32(0), 0.0); err != nil {
		t.Fatal(err)
	}
	tb.MustAddRows([][]any{
		{"b", 2, int8(2), int16(2), int32(2), int64(2), float32(2.5), 2.5},
		{"c", 10, int8(10), int16(10), int32(10), int64(10), float32(10), 10.0},
		{"a", -1, int8(-1), int16(-1), int32(-1), int64(-1), float32(-1), -1.0},
	})
	for index := 0; index < 8; index++ {
		tb.OrderByAsc(index)
		assertColumn(t, tb, 0, "a", "b", "c")
		tb.OrderByDesc(index)
		assertColumn(t, tb, 0, "c", "b", "a")
	}
	assertColumn(t, tb, 6, "10", "2.5", "-1")
}

// TestSortStaysAsRowsChange: rows added to a sorted table are sorted in, a sort set
// while the filter hides the rows applies to all of them, and ClearRows keeps the sort.
func TestSortStaysAsRowsChange(t *testing.T) {
	tb := people(40, 12).OrderByAsc(0)
	tb.MustAddRows([][]any{{50, "Joan", 5.0}, {1, "Ken", 6.0}})
	assertColumn(t, tb, 0, "1", "3", "7", "20", "50", "51", "100")

	tb.SetFilter(1, "ada").OrderByDesc(0).UnsetFilter()
	assertColumn(t, tb, 0, "100", "51", "50", "20", "7", "3", "1")

	tb.ClearRows().MustAddRows([][]any{{2, "b", 0.0}, {9, "a", 0.0}})
	assertColumn(t, tb, 0, "9", "2")
}

func TestFilterIgnoresCase(t *testing.T) {
	tb := people(40, 8)
	tb.SetFilter(1, "A")
	assertColumn(t, tb, 1, "ada", "Grace", "Barbara")
	if column, s := tb.GetFilter(); column != 1 || s != "A" {
		t.Fatalf("GetFilter = %d, %q", column, s)
	}

	tb.SetFilter(1, "é")
	assertColumn(t, tb, 1, "Édith")
	tb.SetFilter(1, "É")
	assertColumn(t, tb, 1, "Édith")

	// numbers match by their text
	tb.SetFilter(2, "9.")
	assertColumn(t, tb, 1, "ada", "Grace")
	tb.SetFilter(0, "0")
	assertColumn(t, tb, 0, "20", "100")

	tb.SetFilter(1, "zzz")
	assertColumn(t, tb, 1)
	if tb.GetCursorValue() != "" || tb.GetCursorRow() != nil {
		t.Fatal("a table without visible rows has a value under the cursor")
	}

	// an empty filter keeps every row, as does none
	tb.SetFilter(1, "")
	assertColumn(t, tb, 0, "20", "3", "100", "7", "51")
	tb.SetFilter(1, "ada").UnsetFilter()
	assertColumn(t, tb, 0, "20", "3", "100", "7", "51")
	if column, s := tb.GetFilter(); column != -1 || s != "" {
		t.Fatalf("GetFilter = %d, %q after UnsetFilter", column, s)
	}

	// an index out of range changes nothing
	tb.SetFilter(-1, "x").SetFilter(3, "x")
	assertColumn(t, tb, 0, "20", "3", "100", "7", "51")
}

func TestContainsFold(t *testing.T) {
	for _, tc := range []struct {
		s, lower string
		want     bool
	}{
		{"Hello World", "lo w", true},
		{"Hello", "hello", true},
		{"Hello", "hellos", false},
		{"Hello", "", true},
		{"", "a", false},
		{"ÉDITH", "édi", true},
		{"ascii only", "é", false},
		{"abcabd", "abd", true},
	} {
		if got := containsFold(tc.s, tc.lower); got != tc.want {
			t.Errorf("containsFold(%q, %q) = %v, want %v", tc.s, tc.lower, got, tc.want)
		}
	}
}

// TestHeaderAndFooterShowSortAndFilter: the header marks the sorted and the filtered
// column, and the footer names the filter while the cursor is on its column.
func TestHeaderAndFooterShowSortAndFilter(t *testing.T) {
	tb := people(36, 4).OrderByAsc(0).SetFilter(1, "a")
	got := plain(tb.Render())
	if want := "ID ▲        Name      ⑂ Score       "; got[0] != want {
		t.Fatalf("header %q, want %q", got[0], want)
	}
	if want := "                          0:0 / 36:2"; got[3] != want {
		t.Fatalf("footer %q, want %q", got[3], want)
	}
	tb.CursorRight().OrderByDesc(1)
	got = plain(tb.Render())
	if want := "ID          Name ▼    ⑂ Score       "; got[0] != want {
		t.Fatalf("header %q, want %q", got[0], want)
	}
	if want := "       filtered by: \"a\" / 1:0 / 36:2"; got[3] != want {
		t.Fatalf("footer %q, want %q", got[3], want)
	}

	// a footer without room drops the filter first and the size second
	tb.SetWidth(20)
	if got, want := plain(tb.Render())[3], "          1:0 / 20:2"; got != want {
		t.Fatalf("narrow footer %q, want %q", got, want)
	}
	tb.SetWidth(6)
	if got, want := plain(tb.Render())[3], "   1:0"; got != want {
		t.Fatalf("very narrow footer %q, want %q", got, want)
	}
}

// TestHeaderCutsWideTitlesByColumns: titles are cut by the columns they take on screen,
// never inside a character, and the marks stay.
func TestHeaderCutsWideTitlesByColumns(t *testing.T) {
	tb := NewTable(16, 3, []string{"Größenangabe", "日本語の列"}).MustAddRows([][]any{{"x", "y"}})
	if got, want := plain(tb.Render())[0], "Größenan日本語の"; got != want {
		t.Fatalf("header %q, want %q", got, want)
	}
	tb.OrderByAsc(0).CursorRight().SetFilter(1, "y")
	if got, want := plain(tb.Render())[0], "Größe ▲ 日本 ⑂  "; got != want {
		t.Fatalf("header %q, want %q", got, want)
	}
}

func TestCellsCutLongValues(t *testing.T) {
	tb := NewTable(20, 4, []string{"A", "B"}).MustAddRows([][]any{
		{"a very long value", "first line\nsecond line"},
		{"fits", "日本語のテキストです"},
	})
	got := plain(tb.Render())
	if want := "a very lo…first line"; got[1] != want {
		t.Fatalf("row %q, want %q", got[1], want)
	}
	if want := "fits      日本語の… "; got[2] != want {
		t.Fatalf("row %q, want %q", got[2], want)
	}
	if value := tb.GetCursorValue(); value != "a very long value" {
		t.Fatalf("value under the cursor is %q, want the whole value", value)
	}
}

func TestFormat(t *testing.T) {
	for _, tc := range []struct {
		value any
		want  string
	}{
		{"text", "text"},
		{-12, "-12"},
		{int8(-8), "-8"},
		{int16(16), "16"},
		{int32(32), "32"},
		{int64(1) << 40, "1099511627776"},
		{float32(9.5), "9.5"},
		{9.5, "9.5"},
		{123.456, "123.456"},
		{8.0, "8"},
		{1234567.0, "1234567"},
		{0.00025, "0.00025"},
		{true, ""},
	} {
		if got := format(tc.value); got != tc.want {
			t.Errorf("format(%v) = %q, want %q", tc.value, got, tc.want)
		}
	}
}

func TestCursorStaysOnTheRows(t *testing.T) {
	tb := people(40, 5)
	tb.CursorUp().CursorLeft()
	if x, y := tb.GetCursorLocation(); x != 0 || y != 0 {
		t.Fatalf("cursor at %d:%d, want 0:0", x, y)
	}
	for i := 0; i < 10; i++ {
		tb.CursorDown().CursorRight()
	}
	if x, y := tb.GetCursorLocation(); x != 2 || y != 4 {
		t.Fatalf("cursor at %d:%d, want 2:4", x, y)
	}
	if got := tb.GetCursorValue(); got != "1234567" {
		t.Fatalf("value under the cursor is %q, want 1234567", got)
	}

	row := tb.GetCursorRow()
	if !slices.Equal(row, []any{51, "Barbara", 1234567.0}) {
		t.Fatalf("row under the cursor is %v", row)
	}
	row[1] = "changed"
	if got := tb.GetCursorRow()[1]; got != "Barbara" {
		t.Fatalf("a change to the returned row reached the table: %v", got)
	}

	// the cursor follows the rows as they go
	tb.SetFilter(1, "ada")
	if x, y := tb.GetCursorLocation(); x != 2 || y != 0 {
		t.Fatalf("cursor at %d:%d on one filtered row, want 2:0", x, y)
	}
	if got := tb.GetCursorValue(); got != "9.5" {
		t.Fatalf("value under the cursor is %q right after the filter, want 9.5", got)
	}
	// a reload keeps the cursor's row, or the last row when fewer rows come back
	tb.UnsetFilter().CursorDown().CursorDown().CursorDown().CursorDown()
	tb.ClearRows().MustAddRows([][]any{{1, "one", 0.0}, {2, "two", 0.0}})
	if _, y := tb.GetCursorLocation(); y != 1 {
		t.Fatalf("cursor on row %d of two rows, want 1", y)
	}
	if got := tb.GetCursorValue(); got != "0" {
		t.Fatalf("value under the cursor is %q, want 0", got)
	}
	tb.ClearRows()
	if x, y := tb.GetCursorLocation(); x != 2 || y != 0 || tb.GetCursorValue() != "" {
		t.Fatalf("cursor at %d:%d with value %q in a table without rows", x, y, tb.GetCursorValue())
	}
}

// TestViewportFollowsTheCursor: the rows scroll as little as they have to, and leave no
// blank lines under the last row while rows are hidden above. A filter moves the cursor
// only while it hides the cursor's row.
func TestViewportFollowsTheCursor(t *testing.T) {
	tb := NewTable(4, 5, []string{"N"})
	for i := 0; i < 10; i++ {
		tb.MustAddRows([][]any{{strconv.Itoa(i)}})
	}
	rows := func() string {
		got := plain(tb.Render())
		return strings.Join(strings.Fields(strings.Join(got[1:len(got)-1], " ")), " ")
	}
	move := func(n int, step func() *Table) {
		for i := 0; i < n; i++ {
			step()
		}
	}
	for _, tc := range []struct {
		name string
		do   func()
		want string
	}{
		{"start", func() {}, "0 1 2"},
		{"down inside the rows", func() { move(2, tb.CursorDown) }, "0 1 2"},
		{"down past the last visible row", func() { move(2, tb.CursorDown) }, "2 3 4"},
		{"up inside the rows", func() { move(2, tb.CursorUp) }, "2 3 4"},
		{"up past the first visible row", func() { move(1, tb.CursorUp) }, "1 2 3"},
		{"to the last row", func() { move(20, tb.CursorDown) }, "7 8 9"},
		{"a taller table shows the rows above", func() { tb.SetHeight(7) }, "5 6 7 8 9"},
		{"a filter with fewer rows than lines", func() { tb.SetFilter(0, "1") }, "1"},
		{"no filter shows the cursor's row again", func() { tb.UnsetFilter() }, "5 6 7 8 9"},
		{"a shorter table keeps the cursor in view", func() { tb.SetHeight(4) }, "8 9"},
	} {
		tc.do()
		if got := rows(); got != tc.want {
			t.Fatalf("%s: rows %q, want %q", tc.name, got, tc.want)
		}
	}
}

// TestScrolledRowsUseTheirColumnsRatios: after scrolling right, the rows must size their
// cells with the ratios and minimum widths of the columns on screen, as the header does.
func TestScrolledRowsUseTheirColumnsRatios(t *testing.T) {
	tb := NewTable(30, 5, []string{"A", "B", "C", "D"})
	tb.SetRatio([]int{1, 1, 2, 1}).SetMinWidth([]int{12, 12, 12, 12})
	tb.MustAddRows([][]any{{"a1", "b1", "c1", "d1"}})
	if l, r := tb.GetVisibleColumnRange(); l != 0 || r != 1 {
		t.Fatalf("visible range %d-%d, want 0-1", l, r)
	}
	tb.CursorRight().CursorRight()
	if l, r := tb.GetVisibleColumnRange(); l != 1 || r != 2 {
		t.Fatalf("visible range %d-%d, want 1-2", l, r)
	}
	got := plain(tb.Render())
	if got[0] != "B           C                 " || got[1] != "b1          c1                " {
		t.Fatalf("header and row do not line up:\n%s\n%s", got[0], got[1])
	}
	tb.CursorLeft().CursorLeft()
	if l, r := tb.GetVisibleColumnRange(); l != 0 || r != 1 {
		t.Fatalf("visible range %d-%d after moving back, want 0-1", l, r)
	}
}

func TestVisibleColumnsFollowWidthsAndCursor(t *testing.T) {
	tb := NewTable(30, 5, []string{"A", "B", "C", "D"})
	if l, r := tb.GetVisibleColumnRange(); l != 0 || r != 3 {
		t.Fatalf("visible range %d-%d without minimum widths, want 0-3", l, r)
	}
	tb.SetMinWidth([]int{10, 10, 10, 10})
	if l, r := tb.GetVisibleColumnRange(); l != 0 || r != 2 {
		t.Fatalf("visible range %d-%d right after SetMinWidth, want 0-2", l, r)
	}
	tb.SetWidth(45)
	if l, r := tb.GetVisibleColumnRange(); l != 0 || r != 3 {
		t.Fatalf("visible range %d-%d at width 45, want 0-3", l, r)
	}
	// the cursor's column shows even when it alone is wider than the table
	tb.SetWidth(4).CursorRight()
	if l, r := tb.GetVisibleColumnRange(); l != 1 || r != 1 {
		t.Fatalf("visible range %d-%d at width 4, want 1-1", l, r)
	}
}

func TestSetTypesAndAddRowsReportBadInput(t *testing.T) {
	tb := NewTable(20, 5, []string{"A", "B"})
	var rowLen ErrorRowLen
	var badType ErrorBadType
	var badCellType ErrorBadCellType

	if _, err := tb.SetTypes(0); !errors.As(err, &rowLen) {
		t.Fatalf("SetTypes with one type for two columns returned %v", err)
	}
	if _, err := tb.SetTypes(nil, ""); !errors.As(err, &badType) {
		t.Fatalf("SetTypes with nil returned %v", err)
	}
	if _, err := tb.SetTypes(true, ""); !errors.As(err, &badType) {
		t.Fatalf("SetTypes with a bool returned %v", err)
	}
	if _, err := tb.SetTypes(0, ""); err != nil {
		t.Fatalf("SetTypes: %v", err)
	}

	if _, err := tb.AddRows([][]any{{1, "a"}, {2}}); !errors.As(err, &rowLen) {
		t.Fatalf("AddRows with a short row returned %v", err)
	}
	if _, err := tb.AddRows([][]any{{1, "a"}, {2, []string{"b"}}}); !errors.As(err, &badType) {
		t.Fatalf("AddRows with a slice value returned %v", err)
	}
	if _, err := tb.AddRows([][]any{{1, "a"}, {"2", "b"}}); !errors.As(err, &badCellType) {
		t.Fatalf("AddRows with a string in an int column returned %v", err)
	} else if want := "table: type string at index 0 does not match the column type int"; err.Error() != want {
		t.Fatalf("error %q, want %q", err, want)
	}
	if tb.GetCursorRow() != nil {
		t.Fatal("a failed AddRows added rows")
	}

	// SetTypes removes the rows
	tb.MustAddRows([][]any{{1, "a"}})
	if _, err := tb.SetTypes("", ""); err != nil {
		t.Fatal(err)
	}
	if tb.GetCursorRow() != nil {
		t.Fatal("SetTypes kept the rows")
	}
}

func TestMustAddRowsPanicsOnBadInput(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("MustAddRows did not panic")
		}
	}()
	NewTable(10, 5, []string{"A"}).MustAddRows([][]any{{1}})
}

func TestSetRatioAndSetMinWidthPanicOnBadInput(t *testing.T) {
	mustPanic := func(name string, f func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Errorf("%s did not panic", name)
			}
		}()
		f()
	}
	mustPanic("SetRatio with one value for two columns", func() { NewTable(10, 5, []string{"A", "B"}).SetRatio([]int{1}) })
	mustPanic("SetRatio with a zero ratio", func() { NewTable(10, 5, []string{"A", "B"}).SetRatio([]int{1, 0}) })
	mustPanic("SetMinWidth with one value for two columns", func() { NewTable(10, 5, []string{"A", "B"}).SetMinWidth([]int{1}) })
}

func TestSetRatioAndSetMinWidthCopyTheirInput(t *testing.T) {
	ratios, widths := []int{1, 3}, []int{0, 0}
	tb := NewTable(20, 3, []string{"A", "B"}).SetRatio(ratios).SetMinWidth(widths).MustAddRows([][]any{{"a", "b"}})
	before := tb.Render()
	ratios[0], widths[0] = 9, 30
	tb.SetWidth(20)
	if tb.Render() != before {
		t.Fatal("a change to the slices given to SetRatio and SetMinWidth reached the table")
	}
}

func TestSetStylesKeepsDefaultsForOtherTables(t *testing.T) {
	custom := lipgloss.NewStyle().Background(lipgloss.Color("#123456"))
	a := NewTable(10, 5, []string{"A"})
	a.SetStyles(map[StyleKey]lipgloss.Style{StyleKeyHeader: custom})
	b := NewTable(10, 5, []string{"A"})
	if b.styles[StyleKeyHeader].GetBackground() == custom.GetBackground() {
		t.Fatal("SetStyles on one table changed the header style of another")
	}
	if a.styles[StyleKeyHeader].GetBackground() != custom.GetBackground() {
		t.Fatal("SetStyles did not apply the override")
	}
	if a.styles[StyleKeyFooter].GetBackground() != b.styles[StyleKeyFooter].GetBackground() {
		t.Fatal("SetStyles changed a style it was not given")
	}
}

// TestStylesShowOnTheNextRender: alignment set through the styles reaches the header,
// the rows and the cursor cell, with and without style passing.
func TestStylesShowOnTheNextRender(t *testing.T) {
	right := lipgloss.NewStyle().Align(lipgloss.Right)
	tb := NewTable(8, 4, []string{"A", "B"}).MustAddRows([][]any{{"a", "b"}, {"c", "d"}})
	assertRendered(t, tb, "A   B   ", "a   b   ", "c   d   ", "     0:0")

	tb.SetStyles(map[StyleKey]lipgloss.Style{StyleKeyCellCursor: right})
	assertRendered(t, tb, "A   B   ", "   ab   ", "c   d   ", "     0:0")
	tb.CursorDown().CursorRight()
	assertRendered(t, tb, "A   B   ", "a   b   ", "c      d", "     1:1")

	tb.SetStyles(map[StyleKey]lipgloss.Style{StyleKeyRows: right, StyleKeyHeader: right, StyleKeyFooter: lipgloss.NewStyle()})
	assertRendered(t, tb, "A   B   ", "a   b   ", "c   d   ", "1:1     ")
	tb.SetStylePassing(true)
	assertRendered(t, tb, "   A   B", "a   b   ", "c   d   ", "1:1     ")
	tb.CursorUp()
	assertRendered(t, tb, "   A   B", "a   b   ", "   c   d", "1:0     ")
	tb.SetStylePassing(false)
	assertRendered(t, tb, "A   B   ", "a   b   ", "c   d   ", "1:0     ")
}

// TestRenderMatchesAFreshTable: a table that renders after every change must show what a
// table built from scratch with the same changes shows. That covers the reuse of rows
// and cells between renders.
func TestRenderMatchesAFreshTable(t *testing.T) {
	type op func(*Table)
	newTable := func() *Table {
		tb := NewTable(40, 12, []string{"ID", "Name", "City", "Score"})
		if _, err := tb.SetTypes(0, "", "", 0.0); err != nil {
			t.Fatal(err)
		}
		return tb.SetRatio([]int{1, 3, 2, 2}).SetMinWidth([]int{4, 10, 8, 6})
	}
	for seed := int64(0); seed < 25; seed++ {
		rng := rand.New(rand.NewSource(seed))
		randomRows := func() [][]any {
			var rows [][]any
			for i, n := 0, 1+rng.Intn(30); i < n; i++ {
				k := rng.Intn(500)
				rows = append(rows, []any{k, "name-" + strconv.Itoa(k%37), "city-" + strconv.Itoa(k%5), float64(k) / 8})
			}
			return rows
		}
		randomOp := func() (string, op) {
			n, column := 1+rng.Intn(15), rng.Intn(4)
			switch rng.Intn(15) {
			case 0:
				rows := randomRows()
				return "add rows", func(tb *Table) { tb.MustAddRows(rows) }
			case 1:
				return "sort ascending", func(tb *Table) { tb.OrderByAsc(column) }
			case 2:
				return "sort descending", func(tb *Table) { tb.OrderByDesc(column) }
			case 3:
				needle := strconv.Itoa(rng.Intn(10))
				return "filter", func(tb *Table) { tb.SetFilter(column, needle) }
			case 4:
				return "no filter", func(tb *Table) { tb.UnsetFilter() }
			case 5:
				width, height := 10+rng.Intn(60), rng.Intn(20)
				return "resize", func(tb *Table) { tb.SetWidth(width).SetHeight(height) }
			case 6:
				return "left", func(tb *Table) { tb.CursorLeft() }
			case 7:
				return "right", func(tb *Table) { tb.CursorRight() }
			case 8:
				passing := rng.Intn(2) == 0
				return "style passing", func(tb *Table) { tb.SetStylePassing(passing) }
			case 12:
				row := rng.Intn(40) - 5
				return "set cursor", func(tb *Table) { tb.SetCursorLocation(column, row) }
			case 9:
				style := lipgloss.NewStyle().Align(lipgloss.Position(float64(rng.Intn(3)) / 2))
				key := StyleKey(rng.Intn(6))
				return "styles", func(tb *Table) { tb.SetStyles(map[StyleKey]lipgloss.Style{key: style}) }
			case 10:
				if rng.Intn(4) == 0 {
					return "clear rows", func(tb *Table) { tb.ClearRows() }
				}
				fallthrough
			case 11:
				return "up", func(tb *Table) {
					for i := 0; i < n; i++ {
						tb.CursorUp()
					}
				}
			default:
				return "down", func(tb *Table) {
					for i := 0; i < n; i++ {
						tb.CursorDown()
					}
				}
			}
		}

		rows := randomRows()
		ops := []op{func(tb *Table) { tb.MustAddRows(rows) }}
		live := newTable()
		ops[0](live)
		for step := 0; step < 60; step++ {
			name, next := randomOp()
			ops = append(ops, next)
			next(live)
			got := live.Render()

			fresh := newTable()
			for _, o := range ops {
				o(fresh)
			}
			if want := fresh.Render(); got != want {
				t.Fatalf("seed %d step %d (%s): rendered\n%s\na fresh table renders\n%s", seed, step, name,
					strings.Join(plain(got), "\n"), strings.Join(plain(want), "\n"))
			}
			if again := live.Render(); again != got {
				t.Fatalf("seed %d step %d (%s): a second render differs", seed, step, name)
			}
		}
	}
}

// TestSetWidthWithTheSameWidthChangesNothing: a table sized on every frame, as one in a
// flexbox cell is, keeps its visible columns where the cursor left them.
func TestSetWidthWithTheSameWidthChangesNothing(t *testing.T) {
	tb := NewTable(30, 6, []string{"A", "B", "C", "D", "E"}).SetMinWidth([]int{10, 10, 10, 10, 10})
	tb.MustAddRows([][]any{{"a", "b", "c", "d", "e"}})
	tb.CursorRight().CursorRight().CursorRight().CursorLeft()
	if l, r := tb.GetVisibleColumnRange(); l != 1 || r != 3 {
		t.Fatalf("visible range %d-%d, want 1-3", l, r)
	}
	before := tb.Render()
	tb.SetWidth(30).SetHeight(6)
	if l, r := tb.GetVisibleColumnRange(); l != 1 || r != 3 {
		t.Fatalf("visible range %d-%d after SetWidth with the same width, want 1-3", l, r)
	}
	if tb.Render() != before {
		t.Fatal("SetWidth and SetHeight with the same size changed the render")
	}
	if allocs := testing.AllocsPerRun(20, func() { tb.SetWidth(30).SetHeight(6).Render() }); allocs != 0 {
		t.Fatalf("an unchanged table allocates %v times per render", allocs)
	}
}

// TestStylesKeepTheTableSize: whatever frame a style has, every part of the table stays
// at its size.
func TestStylesKeepTheTableSize(t *testing.T) {
	frames := map[string]lipgloss.Style{
		"top border": lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false),
		"border":     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()),
		"padding":    lipgloss.NewStyle().Padding(1, 2),
		"margin":     lipgloss.NewStyle().Margin(1),
		"height":     lipgloss.NewStyle().Height(3).Width(50),
	}
	for name, style := range frames {
		for key := StyleKeyHeader; key <= StyleKeyCellCursor; key++ {
			for _, size := range [][2]int{{20, 5}, {3, 4}, {0, 3}, {12, 2}} {
				tb := NewTable(size[0], size[1], []string{"A", "B"}).MustAddRows([][]any{{"a", "b"}, {"c", "d"}})
				tb.SetStyles(map[StyleKey]lipgloss.Style{key: style})
				got := strings.Split(tb.Render(), "\n")
				if len(got) != size[1] {
					t.Fatalf("%s on style %d at %dx%d: rendered %d lines", name, key, size[0], size[1], len(got))
				}
				for i, line := range got {
					if w := lipgloss.Width(line); w != size[0] {
						t.Fatalf("%s on style %d at %dx%d: line %d is %d columns wide", name, key, size[0], size[1], i, w)
					}
				}
			}
		}
	}
}

// TestCellsCutValuesToTheRoomTheyHave: the cut leaves room for the padding of the cursor
// cell's style and counts a tab as the four spaces it shows as.
func TestCellsCutValuesToTheRoomTheyHave(t *testing.T) {
	tb := NewTable(20, 4, []string{"A", "B"}).MustAddRows([][]any{
		{"a very long value here", "x"},
		{"tab\tbed value long", "first\r\nsecond"},
	})
	tb.SetStyles(map[StyleKey]lipgloss.Style{StyleKeyCellCursor: lipgloss.NewStyle().Padding(0, 1)})
	assertRendered(t, tb,
		"A         B         ",
		" a very … x         ",
		"tab    be…first     ",
		"          0:0 / 20:2")
	tb.CursorDown()
	assertRendered(t, tb,
		"A         B         ",
		"a very lo…x         ",
		" tab    … first     ",
		"          0:1 / 20:2")
}

// TestRowsShowNegativeZeroAfterAScroll: a row that scrolls into the place of a row with
// the value 0 shows its own -0.
func TestRowsShowNegativeZeroAfterAScroll(t *testing.T) {
	tb := NewTable(6, 3, []string{"N"})
	if _, err := tb.SetTypes(0.0); err != nil {
		t.Fatal(err)
	}
	negativeZero := math.Copysign(0, -1)
	tb.MustAddRows([][]any{{0.0}, {negativeZero}})
	assertRendered(t, tb, "N     ", "0     ", "   0:0")
	tb.CursorDown()
	assertRendered(t, tb, "N     ", "-0    ", "   0:1")
}

func TestNewTableCopiesItsHeaders(t *testing.T) {
	headers := []string{"A", "B"}
	tb := NewTable(8, 3, headers)
	headers[0] = "Z"
	tb.OrderByAsc(1)
	if got := plain(tb.Render())[0]; got != "A   B ▲ " {
		t.Fatalf("header %q, want %q", got, "A   B ▲ ")
	}
}

// TestMouseFindsRowsAndColumns: ColumnAt and RowAt turn a position on the rendered
// table into a column and a row, and SetCursorLocation moves the cursor there.
func TestMouseFindsRowsAndColumns(t *testing.T) {
	tb := NewTable(30, 6, []string{"A", "B", "C", "D", "E"}).SetMinWidth([]int{10, 10, 10, 10, 10})
	for i := 0; i < 10; i++ {
		n := strconv.Itoa(i)
		tb.MustAddRows([][]any{{"a" + n, "b" + n, "c" + n, "d" + n, "e" + n}})
	}

	for _, tc := range []struct {
		y    int
		row  int
		ok   bool
		name string
	}{
		{-1, 0, false, "above the table"},
		{0, 0, false, "the header"},
		{1, 0, true, "the first line of rows"},
		{4, 3, true, "the last line of rows"},
		{5, 0, false, "the footer"},
		{6, 0, false, "under the table"},
	} {
		if row, ok := tb.RowAt(tc.y); row != tc.row || ok != tc.ok {
			t.Errorf("RowAt(%d), %s, = %d, %v, want %d, %v", tc.y, tc.name, row, ok, tc.row, tc.ok)
		}
	}
	for _, tc := range []struct {
		x      int
		column int
		ok     bool
	}{{-1, 0, false}, {0, 0, true}, {9, 0, true}, {10, 1, true}, {29, 2, true}, {30, 0, false}} {
		if column, ok := tb.ColumnAt(tc.x); column != tc.column || ok != tc.ok {
			t.Errorf("ColumnAt(%d) = %d, %v, want %d, %v", tc.x, column, ok, tc.column, tc.ok)
		}
	}

	// a click on line 3, column 25 selects the third row and the third column
	column, row, ok := tb.CellAt(25, 3)
	if !ok {
		t.Fatal("CellAt found no cell at a row")
	}
	tb.SetCursorLocation(column, row)
	if x, y := tb.GetCursorLocation(); x != 2 || y != 2 || tb.GetCursorValue() != "c2" {
		t.Fatalf("cursor at %d:%d on %q, want 2:2 on c2", x, y, tb.GetCursorValue())
	}

	// the table scrolls to a cursor set outside the visible rows and columns
	tb.SetCursorLocation(4, 9)
	if l, r := tb.GetVisibleColumnRange(); l != 2 || r != 4 {
		t.Fatalf("visible range %d-%d with the cursor on the last column, want 2-4", l, r)
	}
	if got := plain(tb.Render())[1:5]; !slices.Equal(got, []string{
		"c6        d6        e6        ",
		"c7        d7        e7        ",
		"c8        d8        e8        ",
		"c9        d9        e9        ",
	}) {
		t.Fatalf("rows after the move:\n%s", strings.Join(got, "\n"))
	}
	if row, ok := tb.RowAt(1); row != 6 || !ok {
		t.Fatalf("RowAt(1) = %d, %v after scrolling, want 6, true", row, ok)
	}
	if column, ok := tb.ColumnAt(0); column != 2 || !ok {
		t.Fatalf("ColumnAt(0) = %d, %v after scrolling, want 2, true", column, ok)
	}

	// values out of range go to the nearest column and row
	tb.SetCursorLocation(-3, 99)
	if x, y := tb.GetCursorLocation(); x != 0 || y != 9 {
		t.Fatalf("cursor at %d:%d, want 0:9", x, y)
	}
	tb.SetCursorLocation(99, -3)
	if x, y := tb.GetCursorLocation(); x != 4 || y != 0 {
		t.Fatalf("cursor at %d:%d, want 4:0", x, y)
	}

	// lines without a row, and a table without rows or columns
	tb.SetFilter(0, "a1")
	if _, ok := tb.RowAt(2); ok {
		t.Fatal("RowAt found a row on a blank line")
	}
	empty := NewTable(10, 4, nil).SetCursorLocation(2, 2)
	if x, y := empty.GetCursorLocation(); x != 0 || y != 0 {
		t.Fatalf("cursor at %d:%d in a table without columns, want 0:0", x, y)
	}
	if _, ok := empty.ColumnAt(0); ok {
		t.Fatal("ColumnAt found a column in a table without columns")
	}
}

// TestCellAtFollowsTheRowsAndColumnAtTheHeader: with frames on the styles, the header's
// cells and the rows' cells are in different places, and each method answers for its
// own line.
func TestCellAtFollowsTheRowsAndColumnAtTheHeader(t *testing.T) {
	rows := func() *Table {
		return NewTable(20, 4, []string{"A", "B"}).MustAddRows([][]any{{"a", "b"}, {"c", "d"}})
	}
	at := func(tb *Table, x, y int) string {
		column, row, ok := tb.CellAt(x, y)
		if !ok {
			return "none"
		}
		return strconv.Itoa(column) + ":" + strconv.Itoa(row)
	}

	padded := rows().SetStyles(map[StyleKey]lipgloss.Style{StyleKeyHeader: lipgloss.NewStyle().Padding(0, 1)})
	assertRendered(t, padded, " A        B         ", "a         b         ", "c         d         ", "          0:0 / 20:2")
	if _, ok := padded.ColumnAt(0); ok {
		t.Error("ColumnAt found a header on the header's padding")
	}
	if column, ok := padded.ColumnAt(1); !ok || column != 0 {
		t.Errorf("ColumnAt(1) = %d, %v, want the first header", column, ok)
	}
	if got := at(padded, 0, 1); got != "0:0" {
		t.Errorf("CellAt(0, 1) = %s, want 0:0", got)
	}

	// a row style with padding on the left moves the row's cells, not the header's
	indented := rows().SetStyles(map[StyleKey]lipgloss.Style{
		StyleKeyRows:           lipgloss.NewStyle().PaddingLeft(4),
		StyleKeyRowsSubsequent: lipgloss.NewStyle().PaddingLeft(4),
		StyleKeyRowsCursor:     lipgloss.NewStyle().PaddingLeft(4),
	})
	assertRendered(t, indented, "A         B         ", "    a       b       ", "    c       d       ", "          0:0 / 20:2")
	for _, tc := range []struct {
		x, y int
		want string
	}{{0, 1, "none"}, {3, 1, "none"}, {4, 1, "0:0"}, {11, 2, "0:1"}, {12, 2, "1:1"}, {19, 1, "1:0"}, {20, 1, "none"}, {5, 0, "none"}, {5, 3, "none"}} {
		if got := at(indented, tc.x, tc.y); got != tc.want {
			t.Errorf("CellAt(%d, %d) = %s, want %s", tc.x, tc.y, got, tc.want)
		}
	}
	if column, ok := indented.ColumnAt(10); !ok || column != 1 {
		t.Errorf("ColumnAt(10) = %d, %v, want the second header", column, ok)
	}

	// a header without room for its frame shows no header, and a table without lines none
	framed := rows().SetStyles(map[StyleKey]lipgloss.Style{StyleKeyHeader: lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false)})
	if _, ok := framed.ColumnAt(3); ok {
		t.Error("ColumnAt found a header in a header line that shows a border")
	}
	if got := at(framed, 3, 1); got != "0:0" {
		t.Errorf("CellAt(3, 1) = %s under a framed header, want 0:0", got)
	}
	flat := rows().SetHeight(0)
	if _, ok := flat.ColumnAt(2); ok || at(flat, 2, 1) != "none" {
		t.Error("a table without lines has a header or a cell")
	}
}

// TestRenderFollowsTheHeightWithoutWidth: a table without columns to show renders its
// current number of lines.
func TestRenderFollowsTheHeightWithoutWidth(t *testing.T) {
	tb := NewTable(0, 1, []string{"A"}).MustAddRows([][]any{{"a"}})
	for _, height := range []int{1, 3, 2, 5, 1} {
		got := tb.SetHeight(height).Render()
		if n := strings.Count(got, "\n") + 1; n != height || lipgloss.Width(got) != 0 {
			t.Fatalf("at %d lines the table renders %d lines of %d columns: %q", height, n, lipgloss.Width(got), got)
		}
	}
}
