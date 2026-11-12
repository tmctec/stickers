package table

import (
	"strings"
	"unicode/utf8"
)

// SetFilter keeps only the rows whose value in the column contains s, ignoring case. An
// empty s keeps every row. An index out of range changes nothing. The header marks the
// column and the footer names the filter while the cursor is on that column.
func (t *Table) SetFilter(columnIndex int, s string) *Table {
	if columnIndex < 0 || columnIndex >= len(t.columnHeaders) {
		return t
	}
	t.filteredColumn, t.filterString = columnIndex, s
	t.filterStale = true
	t.headerStale, t.rowsStale = true, true
	return t
}

// UnsetFilter removes the filter so every row is visible again.
func (t *Table) UnsetFilter() *Table {
	t.filteredColumn, t.filterString = -1, ""
	t.filterStale = true
	t.headerStale, t.rowsStale = true, true
	return t
}

// GetFilter returns the filtered column, -1 when there is none, and the filter string.
func (t *Table) GetFilter() (columnIndex int, s string) {
	return t.filteredColumn, t.filterString
}

// applyFilter sets filteredRows to the rows the filter keeps: all of them without a
// filter.
func (t *Table) applyFilter() {
	if t.filterString == "" || t.filteredColumn < 0 {
		t.filteredRows = t.rows
		return
	}
	needle := strings.ToLower(t.filterString)
	kept := t.kept[:0]
	for _, row := range t.rows {
		if containsFold(format(row[t.filteredColumn]), needle) {
			kept = append(kept, row)
		}
	}
	// rows the last filter kept and this one does not must not stay reachable
	if len(kept) < len(t.kept) {
		clear(t.kept[len(kept):])
	}
	t.kept, t.filteredRows = kept, kept
}

// containsFold reports whether s contains lower, which must be lower case, ignoring the
// case of s.
func containsFold(s, lower string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return strings.Contains(strings.ToLower(s), lower)
		}
	}
	// s is ASCII, so it is compared byte by byte without a lower-case copy
	for start := 0; start+len(lower) <= len(s); start++ {
		i := 0
		for ; i < len(lower); i++ {
			c := s[start+i]
			if 'A' <= c && c <= 'Z' {
				c += 'a' - 'A'
			}
			if c != lower[i] {
				break
			}
		}
		if i == len(lower) {
			return true
		}
	}
	return false
}
