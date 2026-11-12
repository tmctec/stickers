package table

import (
	"fmt"
	"slices"

	"charm.land/lipgloss/v2"
	"github.com/tmctec/stickers/v2/flexbox"
)

// Table is a responsive table with a one-line header, one-line rows and a one-line
// footer, laid out with flexbox. It scrolls vertically with the cursor and horizontally
// when the columns' minimum widths do not fit the table width.
//
// A Table is not safe for concurrent use.
type Table struct {
	columnHeaders []string
	// columnRatio and columnMinWidth size the header cells and the row cells alike
	columnRatio    []int
	columnMinWidth []int
	columnKinds    []kind

	rows [][]any
	// filteredRows are the rows the filter keeps: rows itself without a filter
	filteredRows   [][]any
	filteredColumn int
	filterString   string
	// kept is the storage filteredRows uses while a filter is set
	kept [][]any

	// orderedColumnIndex is the sorted column, -1 when the rows are unsorted
	orderedColumnIndex int
	orderedColumnPhase SortingOrderKey

	// sortStale and filterStale mark work that settle does before the rows are read:
	// rows were added or the order changed, and the rows, their order or the filter
	// changed
	sortStale   bool
	filterStale bool

	// cursorRow is the row the cursor was moved to and topRow the first visible row
	// that move left; cursorIndexY and rowsTopIndex are where the rows the filter keeps
	// put the two, as position works out
	cursorRow    int
	topRow       int
	cursorIndexY int
	rowsTopIndex int
	cursorIndexX int
	// cursorMovedRight tells from which side fitColumns fills the visible columns
	cursorMovedRight bool

	columnVisibleLeftIndex  int
	columnVisibleRightIndex int

	height int
	width  int
	// rowsBoxHeight is the height left for the rows under the header and over the footer
	rowsBoxHeight int

	styles map[StyleKey]lipgloss.Style
	// stylesVersion changes whenever the styles or the style passing do
	stylesVersion int
	stylePassing  bool

	// the header, the rows and the footer are a box each, which keeps every part at
	// exactly its size whatever its style
	headerBox  *flexbox.FlexBox
	rowsBox    *flexbox.FlexBox
	footerBox  *flexbox.FlexBox
	footerCell *flexbox.Cell

	// headerStale, rowsStale and columnsStale mark what Render builds again: the
	// header cells, the slots' contents and styles, and the slots themselves
	headerStale  bool
	rowsStale    bool
	columnsStale bool

	// slots are the lines of the rows area, slotsTop the row the first one showed last,
	// visibleRows the flexbox rows of the slots in use
	slots       []*slot
	slotsTop    int
	visibleRows []*flexbox.Row

	// the last footer text and the last render, and what they were made from
	footerOut string
	footerKey footerKey
	footerSet bool
	out       string
	parts     [3]string
	outHeight int
	rendered  bool
}

// NewTable initializes a Table of the given size with one string column per header, all
// with ratio 1 and no minimum width.
func NewTable(width, height int, columnHeaders []string) *Table {
	columns := len(columnHeaders)
	t := &Table{
		columnHeaders:  slices.Clone(columnHeaders),
		columnRatio:    slices.Repeat([]int{1}, columns),
		columnMinWidth: make([]int, columns),
		columnKinds:    make([]kind, columns),

		filteredColumn:     -1,
		orderedColumnIndex: -1,
		orderedColumnPhase: SortingOrderDescending,

		width:         width,
		height:        height,
		rowsBoxHeight: max(height-2, 0),

		headerBox: flexbox.New(width, 1),
		rowsBox:   flexbox.New(width, max(height-2, 0)).LockRowHeight(1),
		footerBox: flexbox.New(width, 1),

		headerStale: true,
		rowsStale:   true,
	}
	t.footerCell = flexbox.NewCell(1, 1).SetContentGenerator(func(width, _ int) string {
		return t.footerText(width)
	})
	t.footerBox.AddRows([]*flexbox.Row{t.footerBox.NewRow().AddCells(t.footerCell)})
	t.SetStyles(nil)
	t.fitColumns()
	return t
}

// SetRatio sets the width ratio of every column. It needs exactly one value per column,
// each greater than 0, and panics otherwise.
func (t *Table) SetRatio(values []int) *Table {
	if len(values) != len(t.columnHeaders) {
		panic(fmt.Sprintf("table: %d ratios for %d columns", len(values), len(t.columnHeaders)))
	}
	for _, value := range values {
		if value < 1 {
			panic("table: a column ratio must be greater than 0")
		}
	}
	t.columnRatio = slices.Clone(values)
	t.headerStale, t.rowsStale, t.columnsStale = true, true, true
	return t
}

// SetTypes sets the type of every column from one zero value per column, such as 0, ""
// or 0.0, and removes all rows, so call it before adding rows. Accepted types are string,
// int, int8, int16, int32, int64, float32 and float64. A wrong number of values returns
// an ErrorRowLen and a value of another type an ErrorBadType; both leave the table as it
// was.
func (t *Table) SetTypes(columnTypes ...any) (*Table, error) {
	if len(columnTypes) != len(t.columnHeaders) {
		return t, errRowLen("column types", len(columnTypes), len(t.columnHeaders))
	}
	kinds := make([]kind, len(columnTypes))
	for i, value := range columnTypes {
		k, ok := kindOf(value)
		if !ok {
			return t, errBadType(value, i)
		}
		kinds[i] = k
	}
	t.columnKinds = kinds
	t.cursorIndexX, t.cursorRow, t.topRow = 0, 0, 0
	t.ClearRows()
	t.fitColumns()
	return t, nil
}

// SetMinWidth sets the minimum width of every column. It needs exactly one value per
// column and panics otherwise. Minimum widths also decide how many columns are visible at
// once, see GetVisibleColumnRange.
func (t *Table) SetMinWidth(values []int) *Table {
	if len(values) != len(t.columnHeaders) {
		panic(fmt.Sprintf("table: %d minimum widths for %d columns", len(values), len(t.columnHeaders)))
	}
	t.columnMinWidth = slices.Clone(values)
	t.headerStale, t.rowsStale, t.columnsStale = true, true, true
	t.fitColumns()
	return t
}

// SetHeight sets the height of the table. The header and footer take one line each and
// the rows get the rest. A table of one line shows the header alone.
func (t *Table) SetHeight(value int) *Table {
	if value == t.height {
		return t
	}
	t.height = value
	t.rowsBoxHeight = max(value-2, 0)
	t.rowsBox.SetHeight(t.rowsBoxHeight)
	t.rowsStale = true
	return t
}

// SetWidth sets the width of the table. A new width recomputes which columns are
// visible, starting from the cursor's column.
func (t *Table) SetWidth(value int) *Table {
	if value == t.width {
		return t
	}
	t.width = value
	t.headerBox.SetWidth(value)
	t.rowsBox.SetWidth(value)
	t.footerBox.SetWidth(value)
	t.fitColumns()
	return t
}

// AddRows appends rows. Every row must have one value per column, of the column's type;
// the first row that does not returns an ErrorRowLen, ErrorBadType or ErrorBadCellType
// and nothing is added. A sorted table sorts the new rows in. The table keeps the rows
// it is given, so do not change them afterwards.
func (t *Table) AddRows(rows [][]any) (*Table, error) {
	for _, row := range rows {
		if err := t.validateRow(row); err != nil {
			return t, err
		}
	}
	t.rows = append(t.rows, rows...)
	t.sortStale, t.filterStale, t.rowsStale = true, true, true
	return t, nil
}

// MustAddRows is AddRows that panics on error.
func (t *Table) MustAddRows(rows [][]any) *Table {
	if _, err := t.AddRows(rows); err != nil {
		panic(err)
	}
	return t
}

// ClearRows removes every row, for example before reloading data in an update loop. The
// sort, the filter and the cursor's row stay and apply to the rows added next.
func (t *Table) ClearRows() *Table {
	clear(t.rows)
	t.rows = t.rows[:0]
	t.filterStale, t.rowsStale = true, true
	return t
}

// Render returns the table as exactly its height in lines, each its width in columns:
// the header, the visible rows and a footer that reads "x:y / w:h", the cursor position
// and the size of the rows area, prefixed by the filter when the cursor is on the
// filtered column. Only what changed since the last render is rendered again.
func (t *Table) Render() string {
	if t.height <= 0 {
		return ""
	}
	t.settle()
	t.updateHeader()
	t.updateRows()

	parts := [3]string{t.headerBox.Render()}
	if t.rowsBoxHeight > 0 {
		parts[1] = t.rowsBox.Render()
	}
	if t.height > 1 {
		parts[2] = t.footerBox.Render()
	}
	// a table without columns to show has three empty parts at every height
	if t.rendered && parts == t.parts && t.height == t.outHeight {
		return t.out
	}
	t.parts, t.outHeight, t.rendered = parts, t.height, true

	t.out = parts[0]
	if t.rowsBoxHeight > 0 {
		t.out += "\n" + parts[1]
	}
	if t.height > 1 {
		t.out += "\n" + parts[2]
	}
	return t.out
}

// settle brings the rows up to date with what changed since they were last read: it
// sorts in rows added to a sorted table, applies the filter, and works out where the
// cursor and the first visible row are.
func (t *Table) settle() {
	if t.sortStale {
		t.sortStale, t.filterStale = false, true
		t.sortRows()
	}
	if t.filterStale {
		t.filterStale = false
		t.applyFilter()
	}
	t.cursorIndexY, t.rowsTopIndex = t.position()
}
