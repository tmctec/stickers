package flexbox

import (
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// blanks is a run of spaces that padding is sliced from.
const blanks = "                                                                " +
	"                                                                "

// block is rendered text of a known size: every line is width columns wide. rev changes
// whenever the block is built again, which tells a parent whether to build its own.
type block struct {
	lines []string
	width int
	rev   uint64
}

// containerKey is what the rendered block of a row, column or box depends on besides its
// children.
type containerKey struct {
	width, height int
	style         uint64
	align         lipgloss.Position
}

// resize returns revs with length n, reusing its storage when it is large enough.
func resize(revs []uint64, n int) []uint64 {
	if cap(revs) < n {
		return make([]uint64, n)
	}
	return revs[:n]
}

// linkStart begins the escape sequence that opens or closes a hyperlink.
const linkStart = "\x1b]8;"

// isolateLinks makes every line open and close its own hyperlink. A hyperlink stays
// open until it is closed, also across line ends, so in a line of a block that sits
// beside other blocks it would take in what follows on the same line of the screen.
// One link is open at a time: a new one replaces it and a close ends it.
func isolateLinks(lines []string) {
	open := "" // the sequence of the link open at this point
	for i, line := range lines {
		before := open
		for rest := line; ; {
			start := strings.Index(rest, linkStart)
			if start < 0 {
				break
			}
			sequence, address := linkSequence(rest[start:])
			if sequence == "" {
				break
			}
			open = ""
			if address != "" {
				open = sequence
			}
			rest = rest[start+len(sequence):]
		}
		line = before + line
		if open != "" {
			line += ansi.ResetHyperlink()
		}
		lines[i] = line
	}
}

// linkSequence returns the hyperlink sequence that text starts with and the address it
// links to, which is empty for the sequence that closes a link. It returns two empty
// strings for a sequence that does not end.
func linkSequence(text string) (sequence, address string) {
	// the sequence ends at a bell or at an escape and a backslash
	end := strings.IndexAny(text[len(linkStart):], "\a\x1b")
	if end < 0 {
		return "", ""
	}
	end += len(linkStart)
	stop := end + 1
	if text[end] == '\x1b' {
		if stop == len(text) || text[stop] != '\\' {
			return "", ""
		}
		stop++
	}
	// it holds parameters, a semicolon and the address
	_, address, _ = strings.Cut(text[len(linkStart):end], ";")
	return text[:stop], address
}

// fill sets the block to text at exactly width by height. Lines over height are cut from
// the inside, between the top frame lines (margin and border) and the bottom ones, so
// both survive whenever height has room for them; below that the top lines stay. Lines
// are cut or padded with spaces to width, and blank lines make up a missing height.
func (b *block) fill(text string, width, height, top, bottom int) {
	lines := appendLines(b.lines[:0], text)
	if strings.Contains(text, linkStart) {
		isolateLinks(lines)
	}
	lines = trimInside(lines, height, top, bottom)
	for i, line := range lines {
		lines[i] = fitLine(line, width)
	}
	for len(lines) < height {
		lines = append(lines, spaces(width))
	}
	b.lines, b.width, b.rev = lines, width, nextID()
}

// collapse sets the block to what an element without room renders as. Without columns it
// is height empty lines, which holds the lines of a column. Without lines it is no lines
// of the given width, which holds the columns of a row.
func (b *block) collapse(width, height int) {
	lines := b.lines[:0]
	if width <= 0 {
		width = 0
		for len(lines) < height {
			lines = append(lines, "")
		}
	}
	b.lines, b.width, b.rev = lines, width, nextID()
}

// join sets the block to parts side by side, tops aligned, at height lines, with left
// spaces before them and right spaces after. A part shorter than height leaves blank
// columns below it and a part without columns adds nothing.
func (b *block) join(parts []*block, height, left, right int) {
	lineLen := func(i int) int {
		n := left + right
		for _, p := range parts {
			if i < len(p.lines) {
				n += len(p.lines[i])
			} else {
				n += p.width
			}
		}
		return n
	}
	size := 0
	for i := 0; i < height; i++ {
		size += lineLen(i)
	}

	var text strings.Builder
	text.Grow(size)
	for i := 0; i < height; i++ {
		writeSpaces(&text, left)
		for _, p := range parts {
			if i < len(p.lines) {
				text.WriteString(p.lines[i])
			} else {
				writeSpaces(&text, p.width)
			}
		}
		writeSpaces(&text, right)
	}

	// every line is a piece of the one string the builder allocated
	lines, joined, start := b.lines[:0], text.String(), 0
	for i := 0; i < height; i++ {
		end := start + lineLen(i)
		lines, start = append(lines, joined[start:end]), end
	}
	width := left + right
	for _, p := range parts {
		width += p.width
	}
	b.lines, b.width, b.rev = lines, width, nextID()
}

// stack sets the block to parts on top of each other, cut at height lines. Every line is
// total columns wide: a part narrower than group gets its share of the spare columns on
// the left, as align says, and spaces on the right make up the rest.
func (b *block) stack(parts []*block, group, total int, align lipgloss.Position, height int) {
	lines := b.lines[:0]
	for _, p := range parts {
		left := offset(group-p.width, align)
		right := total - left - p.width
		for _, line := range p.lines {
			if len(lines) >= height {
				break
			}
			switch {
			case right < 0:
				// a part wider than the block is cut at the block's edge
				line = spaces(left) + fitLine(line, total-left)
			case left > 0 || right > 0:
				line = spaces(left) + line + spaces(right)
			}
			lines = append(lines, line)
		}
	}
	b.lines, b.width, b.rev = lines, total, nextID()
}

// clip cuts the lines of a block wider than width at width columns.
func (b *block) clip(width int) {
	if b.width <= width {
		return
	}
	for i, line := range b.lines {
		b.lines[i] = fitLine(line, width)
	}
	b.width = max(width, 0)
}

// extend appends blank lines until the block has height lines.
func (b *block) extend(height int) {
	for len(b.lines) < height {
		b.lines = append(b.lines, spaces(b.width))
	}
}

// widest returns the width of the widest of parts that has lines.
func widest(parts []*block) int {
	width := 0
	for _, p := range parts {
		if len(p.lines) > 0 {
			width = max(width, p.width)
		}
	}
	return width
}

// String joins the lines of the block.
func (b *block) String() string {
	return strings.Join(b.lines, "\n")
}

// appendLines appends the lines of text to dst.
func appendLines(dst []string, text string) []string {
	for {
		line, rest, more := strings.Cut(text, "\n")
		dst = append(dst, line)
		if !more {
			return dst
		}
		text = rest
	}
}

// trimInside cuts lines to height. It removes lines between the top frame lines and
// the bottom frame lines; when height has no room for both frames it keeps the top lines.
func trimInside(lines []string, height, top, bottom int) []string {
	if len(lines) <= height {
		return lines
	}
	if height <= 0 {
		return lines[:0]
	}
	if inner := height - top - bottom; inner >= 0 {
		copy(lines[top+inner:], lines[len(lines)-bottom:])
	}
	return lines[:height]
}

// fitLine returns line at exactly width columns: cut when wider, padded with spaces when
// narrower. Escape sequences stay intact.
func fitLine(line string, width int) string {
	w := lineWidth(line)
	if w == width {
		return line
	}
	if w > width {
		line = ansi.Truncate(line, max(width, 0), "")
		// a wide character at the cut leaves the line a column short
		w = lineWidth(line)
	}
	return line + spaces(width-w)
}

// lineWidth returns the number of columns line takes on screen. Printable ASCII takes
// one column per byte; anything else is measured by ansi.StringWidth.
func lineWidth(line string) int {
	for i := 0; i < len(line); i++ {
		if c := line[i]; c < ' ' || c > '~' {
			return ansi.StringWidth(line)
		}
	}
	return len(line)
}

// textWidth returns the width of the widest line of text.
func textWidth(text string) int {
	widest := 0
	for {
		line, rest, more := strings.Cut(text, "\n")
		widest = max(widest, lineWidth(line))
		if !more {
			return widest
		}
		text = rest
	}
}

// spaces returns n spaces, or "" when n is 0 or less.
func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	if n <= len(blanks) {
		return blanks[:n]
	}
	return strings.Repeat(" ", n)
}

// writeSpaces writes n spaces to sb.
func writeSpaces(sb *strings.Builder, n int) {
	for ; n > len(blanks); n -= len(blanks) {
		sb.WriteString(blanks)
	}
	if n > 0 {
		sb.WriteString(blanks[:n])
	}
}

// offset returns how many of gap spare columns go to the left of text placed at pos,
// where 0 is the left edge and 1 the right edge.
func offset(gap int, pos lipgloss.Position) int {
	if gap <= 0 {
		return 0
	}
	p := math.Min(1, math.Max(0, float64(pos)))
	return int(math.Round(float64(gap) * p))
}

// lead returns how many of spare columns or lines lipgloss puts before content that a
// style aligns at pos: none at the start, half of them at the center, rounded down, and
// all of them at the end.
func lead(spare int, pos lipgloss.Position) int {
	switch {
	case spare <= 0:
		return 0
	case pos == lipgloss.Right:
		return spare
	case pos == lipgloss.Center:
		return spare / 2
	default:
		return 0
	}
}
