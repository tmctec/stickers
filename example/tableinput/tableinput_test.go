package tableinput

import (
	"image"
	"strconv"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/tmctec/stickers/v2/table"
)

func people() *table.Table {
	t := table.NewTable(30, 7, []string{"ID", "Name", "City"})
	for i, name := range []string{"Linus", "Ada", "Grace", "Barbara", "Ken", "Dennis", "Margaret"} {
		t.MustAddRows([][]any{{strconv.Itoa(i), name, "city " + strconv.Itoa(i%3)}})
	}
	return t
}

func press(code rune, text string, mod tea.KeyMod) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Text: text, Mod: mod}
}

func at(t *testing.T, tb *table.Table, x, y int, value string) {
	t.Helper()
	if gotX, gotY := tb.GetCursorLocation(); gotX != x || gotY != y || tb.GetCursorValue() != value {
		t.Fatalf("cursor at %d:%d on %q, want %d:%d on %q", gotX, gotY, tb.GetCursorValue(), x, y, value)
	}
}

func TestKeysMoveSortAndFilter(t *testing.T) {
	tb := people()
	for _, k := range []tea.KeyPressMsg{press(tea.KeyDown, "", 0), press(tea.KeyDown, "", 0), press(tea.KeyRight, "", 0)} {
		if !Update(tb, k, image.Point{}) {
			t.Fatalf("%s is not a key the table takes", k)
		}
	}
	at(t, tb, 1, 2, "Grace")
	Update(tb, press(tea.KeyUp, "", 0), image.Point{})
	Update(tb, press(tea.KeyLeft, "", 0), image.Point{})
	at(t, tb, 0, 1, "1")

	// ctrl+s sorts by the cursor's column, smallest first, then largest first
	Update(tb, press(tea.KeyRight, "", 0), image.Point{})
	Update(tb, press('s', "", tea.ModCtrl), image.Point{})
	at(t, tb, 1, 1, "Barbara")
	Update(tb, press('s', "", tea.ModCtrl), image.Point{})
	at(t, tb, 1, 1, "Linus")

	// letters and digits extend the filter on the cursor's column, backspace shortens it
	Update(tb, press('a', "a", 0), image.Point{})
	Update(tb, press('r', "r", 0), image.Point{})
	if column, filter := tb.GetFilter(); column != 1 || filter != "ar" {
		t.Fatalf("filter %q on column %d, want ar on 1", filter, column)
	}
	at(t, tb, 1, 1, "Barbara")
	Update(tb, press(tea.KeyBackspace, "", 0), image.Point{})
	if _, filter := tb.GetFilter(); filter != "a" {
		t.Fatalf("filter %q after backspace, want a", filter)
	}
	Update(tb, press(tea.KeyBackspace, "", 0), image.Point{})
	if column, _ := tb.GetFilter(); column != -1 {
		t.Fatal("backspace on a filter of one character did not remove it")
	}

	// a filter on another column starts over
	Update(tb, press('k', "k", 0), image.Point{})
	Update(tb, press(tea.KeyLeft, "", 0), image.Point{})
	Update(tb, press('2', "2", 0), image.Point{})
	if column, filter := tb.GetFilter(); column != 0 || filter != "2" {
		t.Fatalf("filter %q on column %d, want 2 on 0", filter, column)
	}

	for _, k := range []tea.KeyPressMsg{press(tea.KeyEnter, "", 0), press(tea.KeySpace, " ", 0), press('!', "!", 0), press(tea.KeyTab, "", 0)} {
		if Update(tb, k, image.Point{}) {
			t.Errorf("%s is taken by the table", k)
		}
	}
	if Update(tb, tea.WindowSizeMsg{Width: 10, Height: 10}, image.Point{}) {
		t.Error("a window size message is taken by the table")
	}
}

func TestMouseSelectsSortsAndScrolls(t *testing.T) {
	tb := people()
	origin := image.Pt(4, 2)

	// the table shows 5 rows under its header at line 2 of the screen
	Update(tb, tea.MouseClickMsg{X: 4 + 12, Y: 2 + 3, Button: tea.MouseLeft}, origin)
	at(t, tb, 1, 2, "Grace")
	Update(tb, tea.MouseClickMsg{X: 4 + 29, Y: 2 + 1, Button: tea.MouseLeft}, origin)
	at(t, tb, 2, 0, "city 0")

	// clicks beside the table, on the footer and with another button change nothing
	for _, click := range []tea.MouseClickMsg{
		{X: 3, Y: 4, Button: tea.MouseLeft},
		{X: 40, Y: 4, Button: tea.MouseLeft},
		{X: 10, Y: 2 + 6, Button: tea.MouseLeft},
		{X: 10, Y: 4, Button: tea.MouseRight},
	} {
		Update(tb, click, origin)
		at(t, tb, 2, 0, "city 0")
	}

	// a click on a header sorts by its column, a second one the other way
	Update(tb, tea.MouseClickMsg{X: 4 + 12, Y: 2, Button: tea.MouseLeft}, origin)
	if column, order := tb.GetOrder(); column != 1 || order != table.SortingOrderAscending {
		t.Fatalf("sorted by column %d in order %d, want 1 ascending", column, order)
	}
	Update(tb, tea.MouseClickMsg{X: 4 + 12, Y: 2, Button: tea.MouseLeft}, origin)
	if column, order := tb.GetOrder(); column != 1 || order != table.SortingOrderDescending {
		t.Fatalf("sorted by column %d in order %d, want 1 descending", column, order)
	}

	// the wheel moves the cursor a row and the table follows
	for i := 0; i < 6; i++ {
		Update(tb, tea.MouseWheelMsg{Button: tea.MouseWheelDown}, origin)
	}
	at(t, tb, 2, 6, "city 1")
	if row, ok := tb.RowAt(1); !ok || row != 2 {
		t.Fatalf("the first visible row is %d, want 2", row)
	}
	Update(tb, tea.MouseWheelMsg{Button: tea.MouseWheelUp}, origin)
	at(t, tb, 2, 5, "city 0")
}

func TestPasteExtendsTheFilter(t *testing.T) {
	tb := people()
	Update(tb, press(tea.KeyRight, "", 0), image.Point{})
	Update(tb, tea.PasteMsg{Content: "  ar\nsecond line"}, image.Point{})
	if column, filter := tb.GetFilter(); column != 1 || filter != "ar" {
		t.Fatalf("filter %q on column %d, want ar on 1", filter, column)
	}
	Update(tb, tea.PasteMsg{Content: "b"}, image.Point{})
	if _, filter := tb.GetFilter(); filter != "arb" {
		t.Fatalf("filter %q, want arb", filter)
	}
	Update(tb, tea.PasteMsg{Content: " \n "}, image.Point{})
	if _, filter := tb.GetFilter(); filter != "arb" {
		t.Fatalf("an empty paste changed the filter to %q", filter)
	}
}

// TestBackspaceRemovesACharacter: the filter loses its last character whole, whatever
// its size in bytes.
func TestBackspaceRemovesACharacter(t *testing.T) {
	tb := people()
	Update(tb, tea.PasteMsg{Content: "a日é"}, image.Point{})
	for _, want := range []string{"a日", "a"} {
		Update(tb, press(tea.KeyBackspace, "", 0), image.Point{})
		if _, filter := tb.GetFilter(); filter != want {
			t.Fatalf("filter %q after backspace, want %q", filter, want)
		}
	}
	Update(tb, press(tea.KeyBackspace, "", 0), image.Point{})
	Update(tb, press('é', "é", 0), image.Point{})
	Update(tb, press(tea.KeyBackspace, "", 0), image.Point{})
	if column, filter := tb.GetFilter(); column != -1 || filter != "" {
		t.Fatalf("filter %q on column %d after removing its only character, want none", filter, column)
	}
	// backspace without a filter, and a click on a table without rows, change nothing
	Update(tb, press(tea.KeyBackspace, "", 0), image.Point{})
	empty := table.NewTable(20, 5, []string{"A"})
	Update(empty, tea.MouseClickMsg{X: 3, Y: 2, Button: tea.MouseLeft}, image.Point{})
	Update(empty, tea.PasteMsg{Content: "x"}, image.Point{})
	if x, y := empty.GetCursorLocation(); x != 0 || y != 0 {
		t.Fatalf("cursor at %d:%d in a table without rows", x, y)
	}
}
