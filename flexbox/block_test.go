package flexbox

import (
	"slices"
	"testing"

	"charm.land/lipgloss/v2"
)

const red = "\x1b[31m"
const reset = "\x1b[0m"

func TestLineWidth(t *testing.T) {
	for _, tc := range []struct {
		line string
		want int
	}{
		{"", 0},
		{"abc", 3},
		{red + "abc" + reset, 3},
		{"日本", 4},
		{"│a│", 3},
	} {
		if got := lineWidth(tc.line); got != tc.want {
			t.Errorf("lineWidth(%q) = %d, want %d", tc.line, got, tc.want)
		}
	}
}

func TestFitLine(t *testing.T) {
	for _, tc := range []struct {
		name, line string
		width      int
		want       string
	}{
		{"exact", "abc", 3, "abc"},
		{"padded", "ab", 4, "ab  "},
		{"cut", "abcdef", 3, "abc"},
		{"cut to nothing", "abc", 0, ""},
		{"escape sequences survive a cut", red + "abcdef" + reset, 3, red + "abc" + reset},
		{"a wide character at the cut leaves a space", "a日本", 2, "a "},
	} {
		if got := fitLine(tc.line, tc.width); got != tc.want {
			t.Errorf("%s: fitLine(%q, %d) = %q, want %q", tc.name, tc.line, tc.width, got, tc.want)
		}
	}
}

func TestTrimInside(t *testing.T) {
	frame := func() []string { return []string{"top", "a", "b", "c", "bottom"} }
	for _, tc := range []struct {
		name                string
		height, top, bottom int
		want                []string
	}{
		{"fits", 5, 1, 1, []string{"top", "a", "b", "c", "bottom"}},
		{"content loses its last lines", 3, 1, 1, []string{"top", "a", "bottom"}},
		{"the frame alone", 2, 1, 1, []string{"top", "bottom"}},
		{"no room for the frame keeps the top", 1, 1, 1, []string{"top"}},
		{"no frame keeps the first lines", 2, 0, 0, []string{"top", "a"}},
		{"no height", 0, 1, 1, []string{}},
	} {
		if got := trimInside(frame(), tc.height, tc.top, tc.bottom); !slices.Equal(got, tc.want) {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestBlockFill(t *testing.T) {
	var b block
	b.fill("ab\ncdef", 3, 4, 0, 0)
	if want := []string{"ab ", "cde", "   ", "   "}; !slices.Equal(b.lines, want) || b.width != 3 {
		t.Fatalf("lines %q width %d, want %q width 3", b.lines, b.width, want)
	}
	rev := b.rev
	b.fill("x", 1, 1, 0, 0)
	if b.rev == rev {
		t.Fatal("a rebuilt block kept its rev")
	}
}

func TestBlockJoin(t *testing.T) {
	left := &block{lines: []string{"ab", "cd", "ef"}, width: 2}
	short := &block{lines: []string{"x"}, width: 1}
	hidden := &block{lines: []string{"", "", ""}, width: 0}
	flat := &block{width: 3}

	var b block
	b.join([]*block{left, short, hidden, flat}, 3, 1, 2)
	want := []string{
		" abx     ",
		" cd      ",
		" ef      ",
	}
	if !slices.Equal(b.lines, want) || b.width != 9 {
		t.Fatalf("lines %q width %d, want %q width 9", b.lines, b.width, want)
	}

	b.clip(4)
	if want := []string{" abx", " cd ", " ef "}; !slices.Equal(b.lines, want) || b.width != 4 {
		t.Fatalf("clipped to %q width %d, want %q width 4", b.lines, b.width, want)
	}
}

func TestBlockStack(t *testing.T) {
	wide := &block{lines: []string{"abcdef"}, width: 6}
	narrow := &block{lines: []string{"xy", "zw"}, width: 2}
	empty := &block{width: 9}

	for _, tc := range []struct {
		align lipgloss.Position
		want  []string
	}{
		{lipgloss.Left, []string{"abcdef  ", "xy      ", "zw      "}},
		{lipgloss.Center, []string{"abcdef  ", "  xy    ", "  zw    "}},
		{lipgloss.Right, []string{"abcdef  ", "    xy  ", "    zw  "}},
	} {
		var b block
		b.stack([]*block{wide, empty, narrow}, 6, 8, tc.align, 10)
		if !slices.Equal(b.lines, tc.want) || b.width != 8 {
			t.Errorf("align %v: lines %q width %d, want %q width 8", tc.align, b.lines, b.width, tc.want)
		}
	}

	var b block
	b.stack([]*block{wide, narrow}, 6, 6, lipgloss.Left, 2)
	b.extend(4)
	if want := []string{"abcdef", "xy    ", "      ", "      "}; !slices.Equal(b.lines, want) {
		t.Fatalf("cut at 2 lines and extended to 4: %q, want %q", b.lines, want)
	}
}

func TestOffset(t *testing.T) {
	for _, tc := range []struct {
		gap  int
		pos  lipgloss.Position
		want int
	}{
		{6, lipgloss.Left, 0},
		{6, lipgloss.Center, 3},
		{5, lipgloss.Center, 3},
		{6, lipgloss.Right, 6},
		{8, 0.25, 2},
		{0, lipgloss.Right, 0},
		{-2, lipgloss.Right, 0},
		{4, 7, 4},
	} {
		if got := offset(tc.gap, tc.pos); got != tc.want {
			t.Errorf("offset(%d, %v) = %d, want %d", tc.gap, tc.pos, got, tc.want)
		}
	}
}

func TestSpaces(t *testing.T) {
	for _, n := range []int{-1, 0, 1, len(blanks), len(blanks) + 7} {
		if got := spaces(n); len(got) != max(n, 0) {
			t.Errorf("spaces(%d) has %d bytes", n, len(got))
		}
	}
}
