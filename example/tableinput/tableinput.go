// Package tableinput holds what the table demos share: how the keyboard, the mouse and
// pasted text drive a table.
package tableinput

import (
	"image"
	"strings"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/tmctec/stickers/v2/table"
)

// Update applies a message to the table and reports whether the message was one the
// table takes.
//
//   - Arrow keys move the cursor and ctrl+s sorts by the cursor's column.
//   - A letter or digit extends the filter on the cursor's column and backspace shortens
//     it.
//   - A left click on a row moves the cursor to the cell under the pointer, and on the
//     header sorts by its column. The wheel moves the cursor a row.
//   - Pasted text extends the filter on the cursor's column.
//
// origin is where the table's top left corner is on screen, which mouse messages are
// counted from.
func Update(t *table.Table, msg tea.Msg, origin image.Point) bool {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return key(t, msg)
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			click(t, msg.X-origin.X, msg.Y-origin.Y)
		}
	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelUp:
			t.CursorUp()
		case tea.MouseWheelDown:
			t.CursorDown()
		}
	case tea.PasteMsg:
		paste(t, msg.Content)
	default:
		return false
	}
	return true
}

func key(t *table.Table, msg tea.KeyPressMsg) bool {
	switch msg.String() {
	case "down":
		t.CursorDown()
	case "up":
		t.CursorUp()
	case "left":
		t.CursorLeft()
	case "right":
		t.CursorRight()
	case "ctrl+s":
		x, _ := t.GetCursorLocation()
		SortBy(t, x)
	case "backspace":
		// the filter loses its last character, which can be more than a byte
		if column, filter := t.GetFilter(); utf8.RuneCountInString(filter) > 1 {
			_, size := utf8.DecodeLastRuneInString(filter)
			t.SetFilter(column, filter[:len(filter)-size])
		} else {
			t.UnsetFilter()
		}
	default:
		typed := []rune(msg.Text)
		if len(typed) != 1 || !unicode.IsLetter(typed[0]) && !unicode.IsDigit(typed[0]) {
			return false
		}
		extendFilter(t, msg.Text)
	}
	return true
}

// SortBy sorts the table by the column: largest first when it is sorted by that column
// smallest first, and smallest first otherwise.
func SortBy(t *table.Table, column int) {
	if sorted, order := t.GetOrder(); sorted == column && order == table.SortingOrderAscending {
		t.OrderByDesc(column)
		return
	}
	t.OrderByAsc(column)
}

// click moves the cursor to the cell at column x and line y of the rendered table, or
// sorts by the column of a header at line 0.
func click(t *table.Table, x, y int) {
	if column, row, ok := t.CellAt(x, y); ok {
		t.SetCursorLocation(column, row)
	} else if column, ok := t.ColumnAt(x); ok && y == 0 {
		SortBy(t, column)
	}
}

// paste extends the filter on the cursor's column by the first line of text.
func paste(t *table.Table, text string) {
	text, _, _ = strings.Cut(strings.TrimSpace(text), "\n")
	if text != "" {
		extendFilter(t, strings.TrimSpace(text))
	}
}

// extendFilter adds text to the filter when it is on the cursor's column, and starts a
// filter on that column with it otherwise.
func extendFilter(t *table.Table, text string) {
	x, _ := t.GetCursorLocation()
	if column, filter := t.GetFilter(); column == x {
		text = filter + text
	}
	t.SetFilter(x, text)
}
