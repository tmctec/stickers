package table

import "fmt"

// ErrorBadType reports a value whose type is not one a column can hold.
type ErrorBadType struct {
	msg string
}

// Error returns the message.
func (e ErrorBadType) Error() string {
	return e.msg
}

// ErrorRowLen reports a row, or a list of column types, whose length differs from the
// number of columns.
type ErrorRowLen struct {
	msg string
}

// Error returns the message.
func (e ErrorRowLen) Error() string {
	return e.msg
}

// ErrorBadCellType reports a value whose type differs from its column's type.
type ErrorBadCellType struct {
	msg string
}

// Error returns the message.
func (e ErrorBadCellType) Error() string {
	return e.msg
}

func errBadType(value any, index int) error {
	return ErrorBadType{msg: fmt.Sprintf("table: type %T at index %d is not one a column can hold", value, index)}
}

func errRowLen(what string, got, want int) error {
	return ErrorRowLen{msg: fmt.Sprintf("table: %d %s for %d columns", got, what, want)}
}

func errBadCellType(value any, index int, want kind) error {
	return ErrorBadCellType{msg: fmt.Sprintf("table: type %T at index %d does not match the column type %s", value, index, want)}
}
