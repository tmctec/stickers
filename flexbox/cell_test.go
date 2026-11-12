package flexbox

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func bordered() lipgloss.Style { return lipgloss.NewStyle().Border(lipgloss.NormalBorder()) }

// renderAt renders a cell at the given size and returns its lines.
func renderAt(c *Cell, width, height int) []string {
	c.width, c.height = width, height
	return c.render(inherited{}).lines
}

func assertLines(t *testing.T, got []string, want ...string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("rendered\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestCellRenderFitsContent(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered()).SetContent("ab\ncd")
	assertLines(t, renderAt(cell, 6, 4),
		"┌────┐",
		"│ab  │",
		"│cd  │",
		"└────┘")
}

func TestCellRenderKeepsTrailingSpacesOnLastLine(t *testing.T) {
	// Both lines are 4 columns wide; centered in 10 columns they start at column 3.
	cell := NewCell(1, 1).SetStyle(lipgloss.NewStyle().Align(lipgloss.Center)).SetContent("a   \nb   ")
	assertLines(t, renderAt(cell, 10, 2),
		"   a      ",
		"   b      ")
}

// TestCellContentAlignsVertically: the content sits where the style's vertical alignment
// puts it, also when its last line ends in a space, which centers one column further
// left.
func TestCellContentAlignsVertically(t *testing.T) {
	for _, tc := range []struct {
		pos     lipgloss.Position
		content string
		want    []string
	}{
		{lipgloss.Top, "x", []string{"  x  ", "     ", "     "}},
		{lipgloss.Center, "x", []string{"     ", "  x  ", "     "}},
		{lipgloss.Bottom, "x", []string{"     ", "     ", "  x  "}},
		{lipgloss.Top, "x ", []string{" x   ", "     ", "     "}},
		{lipgloss.Center, "x ", []string{"     ", " x   ", "     "}},
		{lipgloss.Bottom, "x ", []string{"     ", "     ", " x   "}},
	} {
		cell := NewCell(1, 1).SetStyle(lipgloss.NewStyle().Align(lipgloss.Center, tc.pos)).SetContent(tc.content)
		assertLines(t, renderAt(cell, 5, 3), tc.want...)
	}
}

func TestRenderStyledKeepsTrailingSpaces(t *testing.T) {
	right := lipgloss.NewStyle().Width(6).Align(lipgloss.Right)
	if got := renderStyled(right, "ab  "); got != "  ab  " {
		t.Fatalf("rendered %q, want %q", got, "  ab  ")
	}
	if got := renderStyled(right, "ab"); got != "    ab" {
		t.Fatalf("rendered %q, want %q", got, "    ab")
	}
	if got := renderStyled(right.Underline(true), "a "); strings.Contains(got, keeper) {
		t.Fatalf("rendered %q, which still holds the marker", got)
	}
}

func TestCellRenderClipsContentNotBorder(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered()).SetContent("ab\ncd")
	assertLines(t, renderAt(cell, 6, 3),
		"┌────┐",
		"│ab  │",
		"└────┘")
}

func TestCellRenderFrameOnly(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered()).SetContent("ab\ncd")
	assertLines(t, renderAt(cell, 6, 2),
		"┌────┐",
		"└────┘")
}

func TestCellRenderBelowFrameKeepsTop(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered()).SetContent("ab\ncd")
	assertLines(t, renderAt(cell, 6, 1), "┌────┐")
}

func TestCellRenderZeroHeightHoldsWidth(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered()).SetContent("ab")
	cell.width, cell.height = 5, 0
	if out := cell.render(inherited{}); len(out.lines) != 0 || out.width != 5 {
		t.Fatalf("rendered %d lines of width %d, want no lines of width 5", len(out.lines), out.width)
	}
}

func TestCellRenderZeroWidthHoldsHeight(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered()).SetContent("ab")
	assertLines(t, renderAt(cell, 0, 3), "", "", "")
}

func TestCellRenderClipsWrappedContent(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered()).SetContent("aaaa bbbb cccc")
	assertLines(t, renderAt(cell, 6, 4),
		"┌────┐",
		"│aaaa│",
		"│bbbb│",
		"└────┘")
}

func TestCellRenderPaddingLargerThanContent(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered().Padding(1, 0)).SetContent("ab")
	assertLines(t, renderAt(cell, 6, 4),
		"┌────┐",
		"│    │",
		"│    │",
		"└────┘")
}

func TestCellRenderKeepsMargins(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered().Margin(1, 0)).SetContent("ab\ncd")
	assertLines(t, renderAt(cell, 6, 4),
		"      ",
		"┌────┐",
		"└────┘",
		"      ")
}

// TestRowShortCellsKeepBottomBorder is demo-7's bottom row at a 4-line height: the cells
// with the smaller height ratios get 3 and 2 lines and still close their borders.
func TestRowShortCellsKeepBottomBorder(t *testing.T) {
	box := New(70, 4)
	row := box.NewRow()
	for _, c := range []struct {
		label  string
		ratioY int
	}{{"P", 4}, {"Q", 3}, {"R", 2}, {"S", 3}, {"T", 4}} {
		row.AddCells(NewCell(1, c.ratioY).SetStyle(bordered()).
			SetContent(c.label + "\n[1:" + string(rune('0'+c.ratioY)) + "]"))
	}
	box.AddRows([]*Row{row})

	got := strings.Split(box.Render(), "\n")
	assertLines(t, got,
		"┌────────────┐┌────────────┐┌────────────┐┌────────────┐┌────────────┐",
		"│P           ││Q           │└────────────┘│S           ││T           │",
		"│[1:4]       │└────────────┘              └────────────┘│[1:4]       │",
		"└────────────┘                                          └────────────┘")
}

// TestColumnSkipsZeroHeightCells: with 2 lines for ratios 1:1:2 the first two cells get
// no lines and the third gets both, so its content must be visible.
func TestColumnSkipsZeroHeightCells(t *testing.T) {
	box := NewHorizontal(3, 2)
	col := box.NewColumn()
	col.AddCells(
		NewCell(1, 1).SetContent("a"),
		NewCell(1, 1).SetContent("b"),
		NewCell(1, 2).SetContent("c"),
	)
	box.AddColumns([]*Column{col})

	assertLines(t, strings.Split(box.Render(), "\n"),
		"c  ",
		"   ")
}

// plainAndStyled renders content in a cell with no style and in a cell with an empty
// style set, which goes through lipgloss, and fails unless both show the same lines.
func plainAndStyled(t *testing.T, content string, width, height int) []string {
	t.Helper()
	plain := renderAt(NewCell(1, 1).SetContent(content), width, height)
	styled := renderAt(NewCell(1, 1).SetStyle(lipgloss.NewStyle()).SetContent(content), width, height)
	if strings.Join(plain, "\n") != strings.Join(styled, "\n") {
		t.Fatalf("content %q at %dx%d renders\n%q\nwithout a style and\n%q\nthrough lipgloss", content, width, height, plain, styled)
	}
	return plain
}

// TestCellWrapsWithoutLosingContent: a cell breaks its content at words, breaks a word
// that is too long, keeps a character with its combining marks, and shows the same
// lines with and without a style.
func TestCellWrapsWithoutLosingContent(t *testing.T) {
	assertLines(t, plainAndStyled(t, " Configuration", 8, 3), "        ", "Configur", "ation   ")
	assertLines(t, plainAndStyled(t, "山田太郎-san", 8, 3), "山田太郎", "-san    ", "        ")
	assertLines(t, plainAndStyled(t, "cafe\u0301s", 4, 2), "cafe\u0301", "s   ")
	assertLines(t, plainAndStyled(t, "a-very-long-word", 6, 3), "a-    ", "very-l", "ong-  ")
	assertLines(t, plainAndStyled(t, "abc\r\ndef", 10, 1), "abc       ")
	assertLines(t, plainAndStyled(t, "abc\r\ndef", 10, 2), "abc       ", "def       ")
	assertLines(t, plainAndStyled(t, "a\tb", 8, 1), "a    b  ")

	contents := []string{
		"", " ", "x", " lead", "trail ", "two  spaces", "a-hyphenated-word", "averyveryverylongword",
		"one two three four five six", "line one\nline two\n\nline four", "  indented\n    more",
		"東京tower", "日本語のテキスト", "mixed 日本語 and ascii", "e\u0301e\u0301e\u0301e\u0301e\u0301", "👾 invaders 👾👾",
		"tab\there", "crlf\r\nline", "zero\u200bwidth",
	}
	for _, content := range contents {
		for width := 1; width <= 12; width++ {
			plainAndStyled(t, content, width, 4)
		}
	}
}

// TestLipglossKeepsFittedContent: the wrap inside lipgloss must leave content that
// fitContent prepared as it is, so a cell shows the lines it counted.
func TestLipglossKeepsFittedContent(t *testing.T) {
	trimmed := func(text string) string {
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimRight(line, " ")
		}
		return strings.Join(lines, "\n")
	}
	for _, content := range []string{
		" Configuration", "山田太郎-san", "cafe\u0301s", "one two three four", "a-b-c-d-e-f-g", "longwordwithoutanybreak",
		"  two leading", "x  y  z", "日本語 text 日本語", "👾👾👾👾", "tab\there and\tthere", "a-very-long-hyphenated-word",
	} {
		for width := 1; width <= 10; width++ {
			fitted := fitContent(content, width, 99)
			for _, line := range strings.Split(fitted, "\n") {
				if w := lipgloss.Width(line); w > width {
					t.Errorf("content %q at width %d has the line %q of %d columns", content, width, line, w)
				}
			}
			if got := trimmed(lipgloss.Wrap(fitted, width, "")); got != trimmed(fitted) {
				t.Errorf("content %q at width %d is fitted as\n%q\nand lipgloss breaks that again as\n%q", content, width, fitted, got)
			}
		}
	}
}

// TestStyledContentInAPlainCellStaysInTheCell: escape sequences in the content are
// closed at every line end, so a color never runs into the next cell.
func TestStyledContentInAPlainCellStaysInTheCell(t *testing.T) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render("long red text that wraps")
	box := New(20, 3)
	box.AddRows([]*Row{box.NewRow().AddCells(NewCell(1, 1).SetContent(red), NewCell(1, 1).SetContent("plain"))})
	for i, line := range lines(box) {
		resets := strings.Count(line, "\x1b[m") + strings.Count(line, "\x1b[0m")
		if strings.Count(line, "\x1b[38") != resets {
			t.Fatalf("line %d leaves a color open: %q", i, line)
		}
	}
	plainAndStyled(t, red, 10, 3)
}

// TestContentLeftOpenTakesNoExtraLine: content that ends inside a color or a hyperlink
// sits where plain content sits, also at the bottom of its cell.
func TestContentLeftOpenTakesNoExtraLine(t *testing.T) {
	bottom := lipgloss.NewStyle().AlignVertical(lipgloss.Bottom)
	for _, content := range []string{"red", "\x1b[31mred", "\x1b[31mred\x1b[0m", "\x1b]8;;https://x.y\ared", "red  "} {
		for _, cell := range []*Cell{NewCell(1, 1).SetStyle(bottom), NewCell(1, 1).SetStyle(bordered().AlignVertical(lipgloss.Bottom))} {
			box := New(6, 5)
			box.AddRows([]*Row{box.NewRow().AddCells(cell.SetContent(content))})
			got := strings.Split(ansi.Strip(box.Render()), "\n")
			last := got[len(got)-1]
			if cell.GetStyle().GetBorderTopSize() > 0 {
				last = got[len(got)-2]
			}
			if !strings.Contains(last, "red") {
				t.Errorf("content %q does not sit on the last line of its cell:\n%s", content, strings.Join(got, "\n"))
			}
		}
	}
}
