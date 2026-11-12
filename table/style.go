package table

import "charm.land/lipgloss/v2"

// StyleKey names a part of the table whose style SetStyles can override.
type StyleKey int

// The parts of the table that take a style.
const (
	StyleKeyHeader         StyleKey = iota // the header line
	StyleKeyFooter                         // the footer line with the cursor position
	StyleKeyRows                           // every second row, starting with the second
	StyleKeyRowsSubsequent                 // every second row, starting with the first
	StyleKeyRowsCursor                     // the row under the cursor
	StyleKeyCellCursor                     // the cell under the cursor
)

// The marks the header shows on a sorted or filtered column.
const (
	sortAscendingMark  = "▲"
	sortDescendingMark = "▼"
	filterMark         = "⑂"
)

// defaultStyles returns a new map of the default styles, so overrides on one table never
// reach another.
func defaultStyles() map[StyleKey]lipgloss.Style {
	header := lipgloss.NewStyle().
		Background(lipgloss.Color("#7158e2")).
		Foreground(lipgloss.Color("#ffffff"))
	return map[StyleKey]lipgloss.Style{
		StyleKeyHeader: header,
		StyleKeyFooter: header.Align(lipgloss.Right),
		StyleKeyRows: lipgloss.NewStyle().
			Background(lipgloss.Color("#4b4b4b")).
			Foreground(lipgloss.Color("#ffffff")),
		StyleKeyRowsSubsequent: lipgloss.NewStyle().
			Background(lipgloss.Color("#3d3d3d")).
			Foreground(lipgloss.Color("#ffffff")),
		StyleKeyRowsCursor: lipgloss.NewStyle().
			Background(lipgloss.Color("#f7b731")).
			Foreground(lipgloss.Color("#000000")).
			Bold(true),
		StyleKeyCellCursor: lipgloss.NewStyle().
			Background(lipgloss.Color("#f6e58d")).
			Foreground(lipgloss.Color("#000000")),
	}
}

// SetStyles overrides the styles of the given parts of the table. Parts left out keep
// their default style.
func (t *Table) SetStyles(styles map[StyleKey]lipgloss.Style) *Table {
	merged := defaultStyles()
	for key, style := range styles {
		merged[key] = style
	}
	t.styles = merged
	t.stylesVersion++
	t.footerCell.SetStyle(merged[StyleKeyFooter])
	t.headerStale, t.rowsStale = true, true
	return t
}

// SetStylePassing sets whether the header and row styles are passed down to their cells.
func (t *Table) SetStylePassing(value bool) *Table {
	t.stylePassing = value
	t.headerBox.StylePassing(value)
	t.rowsBox.StylePassing(value)
	t.stylesVersion++
	t.headerStale, t.rowsStale = true, true
	return t
}
