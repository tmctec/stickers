package table

import "strconv"

// Ordered is the set of types a column can hold and sort by.
type Ordered interface {
	int | int8 | int32 | int16 | int64 | float32 | float64 | string
}

// kind is the type of a column.
type kind uint8

const (
	kindString kind = iota
	kindInt
	kindInt8
	kindInt16
	kindInt32
	kindInt64
	kindFloat32
	kindFloat64
)

// String returns the name of the Go type the kind stands for.
func (k kind) String() string {
	return [...]string{"string", "int", "int8", "int16", "int32", "int64", "float32", "float64"}[k]
}

// kindOf returns the kind of value, and false when a column cannot hold it.
func kindOf(value any) (kind, bool) {
	switch value.(type) {
	case string:
		return kindString, true
	case int:
		return kindInt, true
	case int8:
		return kindInt8, true
	case int16:
		return kindInt16, true
	case int32:
		return kindInt32, true
	case int64:
		return kindInt64, true
	case float32:
		return kindFloat32, true
	case float64:
		return kindFloat64, true
	default:
		return 0, false
	}
}

// format returns the text a table shows for value, or "" for a type a column cannot
// hold. Floats print as plain decimals with the fewest digits that read back as the same
// number.
func format(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return ""
	}
}

// validateRow checks that the row has one value per column, each of its column's type.
func (t *Table) validateRow(row []any) error {
	if len(row) != len(t.columnKinds) {
		return errRowLen("values in a row", len(row), len(t.columnKinds))
	}
	for i, value := range row {
		k, ok := kindOf(value)
		if !ok {
			return errBadType(value, i)
		}
		if k != t.columnKinds[i] {
			return errBadCellType(value, i, t.columnKinds[i])
		}
	}
	return nil
}
