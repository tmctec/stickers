// Package table is a scrollable, sortable and filterable table for terminal apps, laid
// out with package flexbox and rendered with lipgloss.
//
// A Table renders a one-line header, one-line rows and a one-line footer that shows the
// cursor position and the size of the rows area. It scrolls vertically with the cursor
// and horizontally when the columns' minimum widths do not fit the table width.
//
//	t := table.NewTable(width, height, []string{"ID", "Name", "Score"})
//	t, err := t.SetTypes(0, "", 0.0) // one zero value per column
//	t.SetRatio([]int{1, 4, 2}).SetMinWidth([]int{4, 10, 6})
//	t.MustAddRows([][]any{{1, "Ada", 9.5}, {2, "Linus", 8.0}})
//	t.CursorDown()
//	t.OrderByAsc(2)
//	t.SetFilter(1, "li")
//	s := t.Render()
//
// # Data
//
// Columns are strings until SetTypes is called. Accepted column types are string, int,
// int8, int16, int32, int64, float32 and float64. Rows are validated against the column
// types when added. A value wider than its column is cut and ends in "…".
//
// # Sorting and filtering
//
// OrderByAsc and OrderByDesc sort by one column, comparing values of the column's type,
// and keep the rows in that order as more are added. SetFilter keeps the rows whose
// value in one column contains a string, ignoring case. The cursor and GetCursorValue,
// GetCursorRow and GetCursorLocation work on the rows the filter keeps, in sorted order.
//
// # Mouse
//
// CellAt turns a position on the rendered table into a column and a row, and
// SetCursorLocation moves the cursor to them. Line 0 is the header: ColumnAt names the
// column whose header shows at a position there, the column to sort by. RowAt names the
// row on a line.
//
// # Styles
//
// SetStyles overrides any of the StyleKey entries and keeps the defaults for the rest.
// SetStylePassing passes the row styles down to the cells.
//
// # Rendering
//
// Render returns exactly the table's height in lines, each the table's width in columns.
// It renders again only what changed since the last call, so an unchanged table costs
// next to nothing and a cursor move renders the two rows it touches. A Table is not safe
// for concurrent use.
package table
