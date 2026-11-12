package table

import (
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/tmctec/stickers/v2/flexbox"
)

// slot is one line of the rows area: a flexbox row that shows whichever table row
// scrolls into it. Slots are reused between renders, so a row that stays on screen keeps
// its rendered cells.
type slot struct {
	row    *flexbox.Row
	values []any    // the value each cell shows
	texts  []string // the values as text
	fitted []string // the texts cut to the widths of the cells
	widths []int    // the widths fitted was cut for; -1 when a text changed

	style         StyleKey // the row style in use
	cursor        int      // the cell styled as the cursor's; -1 when none
	stylesVersion int      // the stylesVersion the styles were applied at
}

// newSlot returns a slot with one cell for each visible column.
func (t *Table) newSlot() *slot {
	n := t.columnVisibleRightIndex - t.columnVisibleLeftIndex + 1
	s := &slot{
		row:           t.rowsBox.NewRow(),
		values:        make([]any, n),
		texts:         make([]string, n),
		fitted:        make([]string, n),
		widths:        make([]int, n),
		cursor:        -1,
		stylesVersion: -1,
	}
	for i := 0; i < n; i++ {
		s.row.AddCells(t.newSlotCell(s, i))
	}
	return s
}

// newSlotCell returns the cell that shows the i-th visible column in the slot. Its
// content is the slot's text cut to the room the cell has for it, which is cut again
// only when the text or the room changes.
func (t *Table) newSlotCell(s *slot, i int) *flexbox.Cell {
	column := t.columnVisibleLeftIndex + i
	s.widths[i] = -1
	return flexbox.NewCell(t.columnRatio[column], 1).
		SetMinWidth(t.columnMinWidth[column]).
		SetContentGenerator(func(width, _ int) string {
			if s.widths[i] != width {
				s.fitted[i], s.widths[i] = cut(s.texts[i], width), width
			}
			return s.fitted[i]
		})
}

// show sets the values the slot's cells show.
func (s *slot) show(values []any) {
	for i, value := range values {
		if !same(s.values[i], value) {
			s.values[i], s.texts[i], s.widths[i] = value, format(value), -1
		}
	}
}

// same reports whether a and b show as the same text. Floats compare by their bits, so
// 0 and -0 differ and a NaN equals itself.
func same(a, b any) bool {
	switch x := a.(type) {
	case float64:
		y, ok := b.(float64)
		return ok && math.Float64bits(x) == math.Float64bits(y)
	case float32:
		y, ok := b.(float32)
		return ok && math.Float32bits(x) == math.Float32bits(y)
	default:
		return a == b
	}
}

// cut returns the first line of text, with tabs as four spaces, cut to width columns and
// ending in "…" when it is wider.
func cut(text string, width int) string {
	text, _, _ = strings.Cut(text, "\n")
	text = strings.TrimSuffix(text, "\r")
	if strings.Contains(text, "\t") {
		text = strings.ReplaceAll(text, "\t", "    ")
	}
	if ansi.StringWidth(text) <= width {
		return text
	}
	return ansi.Truncate(text, max(width, 0), "…")
}

// updateRows brings the slots of the rows area up to date with the visible rows, the
// cursor and the styles.
func (t *Table) updateRows() {
	if !t.rowsStale {
		return
	}
	t.rowsStale = false

	if t.columnsStale {
		t.columnsStale = false
		t.slots = t.slots[:0]
	}
	top := t.rowsTopIndex
	count := max(min(len(t.filteredRows)-top, t.rowsBoxHeight), 0)

	// a scroll moves the slots with their rows, so only the rows that scroll in render
	if shift := top - t.slotsTop; shift != 0 && len(t.slots) > 0 && shift > -len(t.slots) && shift < len(t.slots) {
		shift = (shift + len(t.slots)) % len(t.slots)
		slices.Reverse(t.slots[:shift])
		slices.Reverse(t.slots[shift:])
		slices.Reverse(t.slots)
	}
	t.slotsTop = top

	for len(t.slots) < count {
		t.slots = append(t.slots, t.newSlot())
	}
	rows := t.visibleRows[:0]
	for i := 0; i < count; i++ {
		s, index := t.slots[i], top+i
		s.show(t.filteredRows[index][t.columnVisibleLeftIndex : t.columnVisibleRightIndex+1])
		t.styleSlot(s, index)
		rows = append(rows, s.row)
	}
	t.visibleRows = rows
	t.rowsBox.SetRows(rows)
}

// styleSlot gives the slot the styles of the table row at index: the cursor row's style
// or one of the two alternating row styles, and the cursor cell's style on the cell
// under the cursor.
func (t *Table) styleSlot(s *slot, index int) {
	key, cursor := StyleKeyRows, -1
	switch {
	case index == t.cursorIndexY:
		key, cursor = StyleKeyRowsCursor, t.cursorIndexX-t.columnVisibleLeftIndex
	case index%2 == 0:
		key = StyleKeyRowsSubsequent
	}
	if cursor >= s.row.CellsLen() {
		cursor = -1
	}

	restyle := s.stylesVersion != t.stylesVersion
	if restyle || s.style != key {
		s.row.SetStyle(t.styles[key]).StylePassing(t.stylePassing)
		s.style = key
	}
	if restyle || s.cursor != cursor {
		if s.cursor >= 0 {
			// a new cell is back on the row's style alone
			s.row.UpdateCellWithIndex(s.cursor, t.newSlotCell(s, s.cursor))
		}
		if cursor >= 0 {
			s.row.GetCell(cursor).SetStyle(t.styles[StyleKeyCellCursor])
		}
		s.cursor = cursor
	}
	s.stylesVersion = t.stylesVersion
}

// updateHeader rebuilds the header row for the visible columns, the sort and the filter.
func (t *Table) updateHeader() {
	if !t.headerStale {
		return
	}
	t.headerStale = false

	row := t.headerBox.NewRow().StylePassing(t.stylePassing)
	for index := t.columnVisibleLeftIndex; index <= t.columnVisibleRightIndex; index++ {
		row.AddCells(t.newHeaderCell(index))
	}
	t.headerBox.SetStyle(t.styles[StyleKeyHeader]).SetRows([]*flexbox.Row{row})
}

// newHeaderCell returns the header cell of the column at index. Its title is built again
// only when the cell's width changes.
func (t *Table) newHeaderCell(index int) *flexbox.Cell {
	titleWidth, title := -1, ""
	return flexbox.NewCell(t.columnRatio[index], 1).
		SetMinWidth(t.columnMinWidth[index]).
		SetContentGenerator(func(width, _ int) string {
			if titleWidth != width {
				titleWidth, title = width, t.headerTitle(index, width)
			}
			return title
		})
}

// headerTitle returns the title of the column at index for a cell of the given width,
// followed by the sort mark when the column is sorted and, as far right as the width
// allows, by the filter mark when it is filtered. A title that does not fit is cut
// before the marks are.
func (t *Table) headerTitle(index, width int) string {
	title := t.columnHeaders[index]
	sorted := t.orderedColumnIndex == index
	filtered := t.filteredColumn == index && t.filterString != ""

	var marks string
	if sorted {
		marks = " " + sortAscendingMark
		if t.orderedColumnPhase == SortingOrderDescending {
			marks = " " + sortDescendingMark
		}
	}
	if filtered {
		// at least one space before the mark and one after it
		gap := max(width-ansi.StringWidth(title+marks)-2, 1)
		marks += strings.Repeat(" ", gap) + filterMark + " "
	}
	if ansi.StringWidth(title+marks) > width {
		if sorted && !filtered {
			// keep the sort mark off the next column
			marks += " "
		}
		title = ansi.Truncate(title, max(width-ansi.StringWidth(marks), 0), "")
	}
	return title + marks
}

// footerKey is everything the text of the footer depends on.
type footerKey struct {
	x, y          int
	width, height int
	room          int
	filter        string
}

// footerText returns the text of the footer for a line with room columns: "x:y / w:h",
// the cursor position and the size of the rows area, prefixed by the filter when the
// cursor is on the filtered column. A line without room for all of it drops the filter
// first and the size second.
func (t *Table) footerText(room int) string {
	key := footerKey{
		x: t.cursorIndexX, y: t.cursorIndexY,
		width: t.width, height: t.rowsBoxHeight,
		room: room,
	}
	if t.cursorIndexX == t.filteredColumn {
		key.filter = t.filterString
	}
	if key == t.footerKey && t.footerSet {
		return t.footerOut
	}
	t.footerKey, t.footerSet = key, true

	status := strconv.Itoa(key.x) + ":" + strconv.Itoa(key.y)
	if sized := status + " / " + strconv.Itoa(key.width) + ":" + strconv.Itoa(key.height); len(sized) <= room {
		status = sized
		if named := "filtered by: " + strconv.Quote(key.filter) + " / " + sized; key.filter != "" && ansi.StringWidth(named) <= room {
			status = named
		}
	}
	t.footerOut = ansi.Truncate(status, max(room, 0), "")
	return t.footerOut
}
