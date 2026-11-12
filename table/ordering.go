package table

import (
	"cmp"
	"slices"
)

// SortingOrderKey is the direction of a sort.
type SortingOrderKey int

// The two sort directions.
const (
	SortingOrderAscending  SortingOrderKey = iota // smallest first
	SortingOrderDescending                        // largest first
)

// GetOrder returns the sorted column, -1 when the rows are unsorted, and the direction.
func (t *Table) GetOrder() (int, SortingOrderKey) {
	return t.orderedColumnIndex, t.orderedColumnPhase
}

// OrderByAsc sorts the rows by the column at index, smallest first, comparing values of
// the column's type. Rows that compare equal keep their order, and rows added later are
// sorted in. An index out of range changes nothing.
func (t *Table) OrderByAsc(index int) *Table {
	return t.orderBy(index, SortingOrderAscending)
}

// OrderByDesc sorts the rows by the column at index, largest first, comparing values of
// the column's type. Rows that compare equal keep their order, and rows added later are
// sorted in. An index out of range changes nothing.
func (t *Table) OrderByDesc(index int) *Table {
	return t.orderBy(index, SortingOrderDescending)
}

// orderBy sorts the rows by the column at index in the given order. Rows added since the
// last sort are first sorted in by the order in place before the call, so they tie the
// way the others do.
func (t *Table) orderBy(index int, order SortingOrderKey) *Table {
	if index < 0 || index >= len(t.columnHeaders) {
		return t
	}
	if t.sortStale {
		t.sortRows()
	}
	t.orderedColumnIndex, t.orderedColumnPhase = index, order
	t.sortRows()
	t.sortStale, t.filterStale = false, true
	t.headerStale, t.rowsStale = true, true
	return t
}

// sortRows sorts the rows by the ordered column, if there is one.
func (t *Table) sortRows() {
	index := t.orderedColumnIndex
	if index < 0 || len(t.rows) < 2 {
		return
	}
	var compare func(a, b []any) int
	switch t.columnKinds[index] {
	case kindString:
		compare = compareColumn[string](index)
	case kindInt:
		compare = compareColumn[int](index)
	case kindInt8:
		compare = compareColumn[int8](index)
	case kindInt16:
		compare = compareColumn[int16](index)
	case kindInt32:
		compare = compareColumn[int32](index)
	case kindInt64:
		compare = compareColumn[int64](index)
	case kindFloat32:
		compare = compareColumn[float32](index)
	case kindFloat64:
		compare = compareColumn[float64](index)
	}
	if t.orderedColumnPhase == SortingOrderDescending {
		ascending := compare
		compare = func(a, b []any) int { return ascending(b, a) }
	}
	slices.SortStableFunc(t.rows, compare)
}

// compareColumn returns a function that compares two rows by their value of type T in
// the column at index.
func compareColumn[T Ordered](index int) func(a, b []any) int {
	return func(a, b []any) int {
		return cmp.Compare(a[index].(T), b[index].(T))
	}
}
