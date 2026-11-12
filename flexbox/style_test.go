package flexbox

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// TestStyleFeaturesKeepSizeAndText: the style features of lipgloss v2 show when set on a
// cell, a row or a box, and leave sizes and text as they are.
func TestStyleFeaturesKeepSizeAndText(t *testing.T) {
	base := bordered().Padding(0, 1)
	render := func(style lipgloss.Style) string {
		box := New(24, 5)
		box.AddRows([]*Row{box.NewRow().AddCells(
			NewCell(1, 1).SetStyle(style).SetContent("left cell with some text"),
			NewCell(1, 1).SetStyle(base).SetContent("right"),
		)})
		return box.Render()
	}
	want := ansi.Strip(render(base))

	for _, tc := range []struct {
		name   string
		style  lipgloss.Style
		marker string // an escape sequence the feature writes
	}{
		{"gradient border", base.BorderForegroundBlend(lipgloss.Color("#ff0000"), lipgloss.Color("#0000ff")), "\x1b[38;2;"},
		{"curly underline in a color", base.UnderlineStyle(lipgloss.UnderlineCurly).UnderlineColor(lipgloss.Color("#00ff00")), "4:3"},
		{"hyperlink", base.Hyperlink("https://example.com"), "\x1b]8;;https://example.com"},
		{"background and bold", base.Background(lipgloss.Color("#222222")).Bold(true), "\x1b[1;"},
	} {
		rendered := render(tc.style)
		assertSize(t, rendered, 24, 5)
		if got := ansi.Strip(rendered); got != want {
			t.Errorf("%s changes the text:\n%s\nwant\n%s", tc.name, got, want)
		}
		if !strings.Contains(rendered, tc.marker) {
			t.Errorf("%s does not show in the render: no %q in %q", tc.name, tc.marker, rendered)
		}
	}

	// a gradient has more than one color along the border
	gradient := render(base.BorderForegroundBlend(lipgloss.Color("#ff0000"), lipgloss.Color("#0000ff")))
	if !strings.Contains(gradient, "\x1b[38;2;255;0;0m") || strings.Count(gradient, "\x1b[38;2;") < 4 {
		t.Errorf("the gradient border has too few colors: %q", gradient)
	}

	// padding and margin characters fill the frame and keep the size
	dotted := render(lipgloss.NewStyle().Padding(1).PaddingChar('.').Margin(0, 1).MarginChar('~'))
	assertSize(t, dotted, 24, 5)
	if second := strings.Split(ansi.Strip(dotted), "\n")[1]; !strings.HasPrefix(second, "~.left.") {
		t.Errorf("padding and margin characters do not show: %q", second)
	}
}

// TestUnderlinedContainersKeepTheirCellsWhole: a row, column or box that underlines or
// strikes through leaves the escape sequences of its styled cells intact, and cells that
// inherit the style show it themselves.
func TestUnderlinedContainersKeepTheirCellsWhole(t *testing.T) {
	lined := lipgloss.NewStyle().Underline(true).Strikethrough(true)
	colored := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	whole := func(name, rendered string) {
		t.Helper()
		assertSize(t, rendered, 12, 2)
		if got := ansi.Strip(rendered); got != "ab    cd    \n            " {
			t.Errorf("%s shows %q", name, got)
		}
	}
	cells := func() []*Cell {
		return []*Cell{NewCell(1, 1).SetStyle(colored).SetContent("ab"), NewCell(1, 1).SetContent("cd")}
	}

	box := New(12, 2)
	row := box.NewRow().SetStyle(lined).AddCells(cells()...)
	box.AddRows([]*Row{row})
	whole("an underlined row", box.Render())
	row.StylePassing(true)
	whole("an underlined row that passes its style", box.Render())
	if rendered := box.Render(); !strings.Contains(rendered, "\x1b[4;") && !strings.Contains(rendered, ";4;") {
		t.Errorf("cells that inherit an underline do not show it: %q", rendered)
	}
	box.SetStyle(lined)
	whole("an underlined box", box.Render())

	horizontal := NewHorizontal(12, 2).SetStyle(lined)
	horizontal.AddColumns([]*Column{
		horizontal.NewColumn().SetStyle(lined).AddCells(cells()[0]),
		horizontal.NewColumn().AddCells(cells()[1]),
	})
	whole("an underlined column in an underlined box", horizontal.Render())

	// plain cells in an underlined row are underlined by the row
	plainBox := New(4, 1)
	plainBox.AddRows([]*Row{plainBox.NewRow().SetStyle(lipgloss.NewStyle().Underline(true)).AddCells(NewCell(1, 1).SetContent("ab"))})
	if rendered := plainBox.Render(); !strings.Contains(rendered, "\x1b[4") {
		t.Errorf("an underlined row does not underline its plain cells: %q", rendered)
	}
}

// TestHyperlinksStayInTheirElement: a hyperlink set on a cell or a column covers that
// element on every line it has text on, and nothing beside it.
func TestHyperlinksStayInTheirElement(t *testing.T) {
	link := lipgloss.NewStyle().Hyperlink("https://example.com")
	open, closed := "\x1b]8;;https://example.com\a", ansi.ResetHyperlink()
	// linked returns how many columns of line the link covers, from the left edge
	linked := func(line string) int {
		t.Helper()
		before, _, found := strings.Cut(line, closed)
		if strings.Count(line, open) != 1 || strings.Count(line, closed) != 1 || !strings.HasPrefix(line, open) || !found {
			t.Fatalf("the line does not open one link at its start and close it: %q", line)
		}
		return ansi.StringWidth(before)
	}

	box := New(12, 3)
	box.AddRows([]*Row{box.NewRow().AddCells(
		NewCell(1, 1).SetStyle(link).SetContent("link\nmore"),
		NewCell(1, 1).SetContent("plain"),
	)})
	got := lines(box)
	assertLines(t, strings.Split(ansi.Strip(box.Render()), "\n"), "link  plain ", "more        ", "            ")
	for i, line := range got[:2] {
		if columns := linked(line); columns < 4 || columns > 6 {
			t.Errorf("line %d links %d columns, want the text of the first cell and no more than its 6 columns: %q", i, columns, line)
		}
	}
	if strings.Contains(got[2], linkStart) {
		t.Errorf("a line without text carries a link: %q", got[2])
	}

	horizontal := NewHorizontal(12, 2)
	horizontal.AddColumns([]*Column{
		horizontal.NewColumn().SetStyle(link).AddCells(NewCell(1, 1).SetContent("link\nmore")),
		horizontal.NewColumn().AddCells(NewCell(1, 1).SetContent("plain")),
	})
	for i, line := range lines(horizontal) {
		if columns := linked(line); columns != 6 {
			t.Errorf("line %d links %d columns, want the 6 columns of the first column: %q", i, columns, line)
		}
	}
}

func TestLinkSequence(t *testing.T) {
	for _, tc := range []struct{ text, sequence, address string }{
		{"\x1b]8;;https://a\atext", "\x1b]8;;https://a\a", "https://a"},
		{"\x1b]8;id=1;https://b\x1b\\text", "\x1b]8;id=1;https://b\x1b\\", "https://b"},
		{"\x1b]8;;\atext", "\x1b]8;;\a", ""},
		{"\x1b]8;;\x1b\\", "\x1b]8;;\x1b\\", ""},
		{"\x1b]8;;https://cut", "", ""},
		{"\x1b]8;;https://cut\x1b", "", ""},
		{"\x1b]8;;https://cut\x1b[m", "", ""},
	} {
		if sequence, address := linkSequence(tc.text); sequence != tc.sequence || address != tc.address {
			t.Errorf("linkSequence(%q) = %q, %q, want %q, %q", tc.text, sequence, address, tc.sequence, tc.address)
		}
	}
}

// TestIsolateLinks: every line closes the link it has open and the next line opens it
// again, and a link around another one comes back when the inner one closes.
func TestIsolateLinks(t *testing.T) {
	a, b, closed := "\x1b]8;;https://a\a", "\x1b]8;;https://b\a", ansi.ResetHyperlink()
	for _, tc := range []struct {
		name     string
		in, want []string
	}{
		{"no links", []string{"plain", "text"}, []string{"plain", "text"}},
		{"closed on its line", []string{a + "x" + closed + "y"}, []string{a + "x" + closed + "y"}},
		{"open over three lines", []string{a + "x", "y", "z" + closed + "w", "v"},
			[]string{a + "x" + closed, a + "y" + closed, a + "z" + closed + "w", "v"}},
		{"a second link replaces the first and its close ends both", []string{a + "x" + b + "y" + closed + "z", "w"},
			[]string{a + "x" + b + "y" + closed + "z", "w"}},
		{"a second link over a line end", []string{a + "x" + b + "y", "z" + closed + "w"},
			[]string{a + "x" + b + "y" + closed, b + "z" + closed + "w"}},
		{"a close without an open", []string{"x" + closed + "y"}, []string{"x" + closed + "y"}},
		{"never closed", []string{a + "x", "y"}, []string{a + "x" + closed, a + "y" + closed}},
	} {
		got := append([]string(nil), tc.in...)
		isolateLinks(got)
		if strings.Join(got, "\n") != strings.Join(tc.want, "\n") {
			t.Errorf("%s:\n got %q\nwant %q", tc.name, got, tc.want)
		}
		for i := range got {
			if ansi.Strip(got[i]) != ansi.Strip(tc.in[i]) {
				t.Errorf("%s: line %d changed its text", tc.name, i)
			}
		}
	}
}

// TestLinkedContentWrapsInsideItsCell: content that carries its own hyperlink and colors
// wraps like plain text, and every line closes what it opens, so neither a link nor a
// color runs into the next cell.
func TestLinkedContentWrapsInsideItsCell(t *testing.T) {
	linked := lipgloss.NewStyle().Hyperlink("https://example.com").Foreground(lipgloss.Color("#ff0000")).
		Render("a linked text that wraps over lines")
	for _, style := range []*lipgloss.Style{nil, ptr(bordered())} {
		cell := NewCell(1, 1).SetContent(linked)
		width := 12
		if style != nil {
			cell.SetStyle(*style)
			width = 14
		}
		box := New(width+6, 6)
		box.AddRows([]*Row{box.NewRow().AddCells(cell.SetFixedWidth(width), NewCell(1, 1).SetContent("plain"))})
		rendered := box.Render()
		assertSize(t, rendered, width+6, 6)

		text := ansi.Strip(rendered)
		for _, word := range []string{"a linked", "text that", "wraps over", "lines", "plain"} {
			if !strings.Contains(text, word) {
				t.Errorf("the render lacks %q:\n%s", word, text)
			}
		}
		for i, line := range strings.Split(rendered, "\n") {
			opened := strings.Count(line, "\x1b]8;;https://example.com")
			closed := strings.Count(line, "\x1b]8;;\x1b\\") + strings.Count(line, "\x1b]8;;\a")
			if opened != closed {
				t.Errorf("line %d opens %d links and closes %d: %q", i, opened, closed, line)
			}
			if at := strings.Index(line, "plain"); at >= 0 && strings.LastIndex(line[:at], "\x1b[38;2;255;0;0m") > strings.LastIndex(line[:at], "\x1b[m") {
				t.Errorf("line %d colors the next cell: %q", i, line)
			}
		}
	}
}

func ptr[T any](v T) *T { return &v }

// TestUnderlineRunsThroughStyledContent: a cell, or a row, that underlines or strikes
// through keeps the escape sequences of styled content whole and draws its lines under
// all of it, also after the content resets its own color.
func TestUnderlineRunsThroughStyledContent(t *testing.T) {
	under := lipgloss.NewStyle().Underline(true)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render("red")
	linked := lipgloss.NewStyle().Hyperlink("https://example.com").Render("link")

	for name, style := range map[string]lipgloss.Style{
		"underline":        under,
		"strikethrough":    lipgloss.NewStyle().Strikethrough(true),
		"curly underline":  lipgloss.NewStyle().UnderlineStyle(lipgloss.UnderlineCurly).UnderlineColor(lipgloss.Color("#00ff00")),
		"underline spaces": lipgloss.NewStyle().UnderlineSpaces(true),
	} {
		for _, content := range []string{red + " and plain", linked, "plain " + red + "\nsecond " + red} {
			// the style on the cell itself, and passed down by the row
			own := New(18, 2)
			own.AddRows([]*Row{own.NewRow().AddCells(NewCell(1, 1).SetStyle(style).SetContent(content))})
			passed := New(18, 2)
			passed.AddRows([]*Row{passed.NewRow().UseStyle(style).AddCells(NewCell(1, 1).SetContent(content))})
			want := New(18, 2)
			want.AddRows([]*Row{want.NewRow().AddCells(NewCell(1, 1).SetContent(content))})
			for kind, box := range map[string]*FlexBox{"on the cell": own, "passed by the row": passed} {
				rendered := box.Render()
				assertSize(t, rendered, 18, 2)
				if got := ansi.Strip(rendered); got != ansi.Strip(want.Render()) {
					t.Errorf("%s %s shows %q for %q", name, kind, got, content)
				}
			}
		}
	}

	// the line is drawn under the colored word and again after its reset
	box := New(14, 1)
	box.AddRows([]*Row{box.NewRow().AddCells(NewCell(1, 1).SetStyle(under).SetContent(red + " after"))})
	rendered := box.Render()
	on := "\x1b[4m"
	if !strings.HasPrefix(rendered, on) || !strings.Contains(rendered, "red\x1b[m"+on+" after") {
		t.Errorf("the underline does not run through the styled content: %q", rendered)
	}

	// a row keeps the underline on its plain cells when another cell is colored
	row := New(12, 1)
	row.AddRows([]*Row{row.NewRow().SetStyle(under).AddCells(NewCell(1, 1).SetContent("ab"), NewCell(1, 1).SetContent(red))})
	if rendered := row.Render(); !strings.HasPrefix(rendered, on+"ab") || ansi.Strip(rendered) != "ab    red   " {
		t.Errorf("an underlined row with a colored cell renders %q", rendered)
	}
}

// TestLinesResumeAfterAnySequenceThatEndsThem: the underline of an element comes back
// after content turns it off in any way, and a row that passes it has no gap after a
// styled cell.
func TestLinesResumeAfterAnySequenceThatEndsThem(t *testing.T) {
	on, off := "\x1b[4m", "\x1b[24m"
	for _, end := range []string{"\x1b[m", "\x1b[0m", "\x1b[00m", "\x1b[24m", "\x1b[29m", "\x1b[59m", "\x1b[4:0m", "\x1b[01;00m"} {
		if got := relined("a"+end+"b", on, off); got != on+"a"+end+on+"b"+off {
			t.Errorf("after %q the line does not resume: %q", end, got)
		}
	}
	for _, keep := range []string{"\x1b[31m", "\x1b[39m", "\x1b[38;5;0m", "\x1b[38;2;0;0;0m", "\x1b[4:3m", "\x1b[38:2::0:0:0m", "\x1b[1;34m"} {
		if got := relined("a"+keep+"b", on, off); got != on+"a"+keep+"b"+off {
			t.Errorf("after %q the line is written again: %q", keep, got)
		}
	}

	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render("red")
	box := New(24, 1)
	box.AddRows([]*Row{box.NewRow().UseStyle(lipgloss.NewStyle().Underline(true)).AddCells(
		NewCell(1, 1).SetContent("ab"), NewCell(1, 1).SetContent(red), NewCell(1, 1).SetContent("cd"))})
	// follow the underline through the line: every column of text and padding has it
	underlined, lined := 0, false
	for rest := box.Render(); rest != ""; {
		if strings.HasPrefix(rest, "\x1b[") {
			end := strings.IndexByte(rest, 'm')
			params := rest[2:end]
			switch {
			case strings.HasPrefix(params, "4"):
				lined = true
			case resetsAttributes(params):
				lined = false
			}
			rest = rest[end+1:]
			continue
		}
		if lined {
			underlined++
		}
		rest = rest[1:]
	}
	if underlined != 24 {
		t.Errorf("%d of 24 columns are underlined in %q", underlined, box.Render())
	}
}

// TestTwoLinksInOneCell: a second link in a cell's content replaces the first, and its
// close ends linking, as on a terminal.
func TestTwoLinksInOneCell(t *testing.T) {
	a, b, off := "\x1b]8;;https://a\a", "\x1b]8;;https://b\a", ansi.ResetHyperlink()
	box := New(16, 2)
	box.AddRows([]*Row{box.NewRow().AddCells(NewCell(1, 1).SetContent(a + "one" + b + "two" + off + " three\nnext line"))})
	got := lines(box)
	if strings.Contains(got[1], linkStart) || !strings.HasSuffix(got[0], off+" three    ") {
		t.Errorf("text after the close is linked: %q", got)
	}
}

func TestRelined(t *testing.T) {
	on, off := "\x1b[4m", "\x1b[24m"
	for _, tc := range []struct{ text, want string }{
		{"plain", on + "plain" + off},
		{"a\n\nb", on + "a" + off + "\n\n" + on + "b" + off},
		{"\x1b[31mred\x1b[m after", on + "\x1b[31mred\x1b[m" + on + " after" + off},
		{"\x1b[0;1mbold", on + "\x1b[0;1m" + on + "bold" + off},
		{"\x1b[38;2;0;0;0mblack\x1b[38;5;0m", on + "\x1b[38;2;0;0;0mblack\x1b[38;5;0m" + off},
		{"\x1b[2Kcleared", on + "\x1b[2Kcleared" + off},
		{"cut \x1b[31", on + "cut \x1b[31" + off},
	} {
		if got := relined(tc.text, on, off); got != tc.want {
			t.Errorf("relined(%q) = %q, want %q", tc.text, got, tc.want)
		}
	}
	if got := relined("text", "", ""); got != "text" {
		t.Errorf("relined without lines changed the text to %q", got)
	}
}

// TestRowLinkComesBackAfterACellLink: a hyperlink on a row covers the cells after one
// that has a link of its own.
func TestRowLinkComesBackAfterACellLink(t *testing.T) {
	rowLink, cellLink := "\x1b]8;;https://row.example\a", "\x1b]8;;https://cell.example\a"
	box := New(9, 1)
	box.AddRows([]*Row{box.NewRow().SetStyle(lipgloss.NewStyle().Hyperlink("https://row.example")).AddCells(
		NewCell(1, 1).SetContent("aa"),
		NewCell(1, 1).SetStyle(lipgloss.NewStyle().Hyperlink("https://cell.example")).SetContent("bb"),
		NewCell(1, 1).SetContent("cc"),
	)})
	rendered := box.Render()
	if ansi.Strip(rendered) != "aa bb cc " {
		t.Fatalf("the row shows %q", ansi.Strip(rendered))
	}
	inRow := func(text string) bool {
		before, _, _ := strings.Cut(rendered, text)
		return strings.LastIndex(before, rowLink) > strings.LastIndex(before, ansi.ResetHyperlink()) &&
			strings.LastIndex(before, rowLink) > strings.LastIndex(before, cellLink)
	}
	if !inRow("aa") || !inRow("cc") || inRow("bb") {
		t.Errorf("the row's link does not cover the first and the last cell alone: %q", rendered)
	}
}
