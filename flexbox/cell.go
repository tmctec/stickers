package flexbox

import (
	"image"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

// Cell is the building block of a grid. In a FlexBox the cells of a row sit side by
// side; in a HorizontalFlexBox the cells of a column stack vertically.
type Cell struct {
	style styling
	// id names the cell for GetCellWithID
	id string

	// ratioX and ratioY are the cell's share of the width and height
	ratioX, ratioY int
	// minWidth and minHeight are the least size of a ratio-sized cell
	minWidth, minHeight int
	// fixedWidth and fixedHeight replace the ratio when greater than 0
	fixedWidth, fixedHeight int

	// width and height are the size the last layout gave the cell, x and y where it
	// put the cell's top left corner in the box
	width, height int
	x, y          int

	// contentGenerator builds the content for the cell's content size
	contentGenerator func(maxX, maxY int) string

	// out is the last render and key what it was rendered from
	out      block
	key      cellKey
	rendered bool
}

// cellKey is everything the rendered block of a cell depends on.
type cellKey struct {
	width, height int
	style         uint64
	content       string
}

// CellOption configures a Cell as NewCell creates it.
type CellOption func(*Cell)

// WithStyle sets the cell's style, with width and height keys removed, as NewCell creates
// it.
func WithStyle(style lipgloss.Style) CellOption {
	return func(c *Cell) {
		c.style.set(style)
	}
}

// NewCell initializes a Cell with the given width and height ratios and an empty style.
// Options such as WithStyle configure the cell before it is returned.
func NewCell(ratioX, ratioY int, opts ...CellOption) *Cell {
	c := &Cell{
		style:  styling{own: lipgloss.NewStyle()},
		ratioX: ratioX,
		ratioY: ratioY,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SetID names the cell for GetCellWithID. Unnamed cells get their index in the row or
// column as their ID.
func (c *Cell) SetID(id string) *Cell {
	c.id = id
	return c
}

// GetID returns the cell's ID: the name SetID gave it, or its index in the row or column
// it was added to.
func (c *Cell) GetID() string {
	return c.id
}

// SetContent sets static content. The cell wraps it at its content width and clips it
// to its content height when it renders. A tab shows as four spaces.
func (c *Cell) SetContent(content string) *Cell {
	c.contentGenerator = func(_, _ int) string {
		return content
	}
	return c
}

// SetContentGenerator sets a function that builds the content each time the cell
// renders. It receives the width and height the content has room for, inside the cell's
// margins, border and padding, never less than 0. A cell without columns or lines does
// not call it.
func (c *Cell) SetContentGenerator(generator func(maxX, maxY int) string) *Cell {
	c.contentGenerator = generator
	return c
}

// GetContent returns the content as set or as the generator builds it for the cell's
// current content size, before wrapping and clipping.
func (c *Cell) GetContent() string {
	if c.contentGenerator == nil {
		return ""
	}
	if !c.style.resolved {
		c.style.resolve(inherited{})
	}
	return c.contentGenerator(c.contentSize())
}

// SetMinWidth keeps a ratio-sized cell at least this wide. In a FlexBox it applies to
// the cell. In a HorizontalFlexBox it applies to the cell's column, which takes the
// largest minimum among its cells.
func (c *Cell) SetMinWidth(value int) *Cell {
	c.minWidth = value
	return c
}

// SetFixedWidth gives the cell a fixed width instead of a ratio-based one. In a FlexBox
// it applies to the cell. In a HorizontalFlexBox it applies to the cell's column, which
// takes the largest fixed width among its cells. A value of 0 or less restores ratio
// sizing.
func (c *Cell) SetFixedWidth(value int) *Cell {
	c.fixedWidth = value
	return c
}

// SetMinHeigth is the misspelled name of SetMinHeight.
//
// Deprecated: use [Cell.SetMinHeight].
func (c *Cell) SetMinHeigth(value int) *Cell {
	return c.SetMinHeight(value)
}

// SetMinHeight keeps a ratio-sized cell at least this tall. In a HorizontalFlexBox it
// applies to the cell. In a FlexBox it applies to the cell's row, which takes the
// largest minimum among its cells.
func (c *Cell) SetMinHeight(value int) *Cell {
	c.minHeight = value
	return c
}

// SetFixedHeight gives the cell a fixed height instead of a ratio-based one. In a
// HorizontalFlexBox it applies to the cell. In a FlexBox it applies to the cell's row,
// which takes the largest fixed height among its cells; each cell then gets its ratioY
// share of that row. A value of 0 or less restores ratio sizing.
func (c *Cell) SetFixedHeight(value int) *Cell {
	c.fixedHeight = value
	return c
}

// SetStyle replaces the style. Width and height keys are removed because the grid owns
// the cell's size.
func (c *Cell) SetStyle(style lipgloss.Style) *Cell {
	c.style.set(style)
	return c
}

// GetStyle returns a copy of the style set on the cell. What a row or column passes
// down is not part of it.
func (c *Cell) GetStyle() lipgloss.Style {
	return c.style.own
}

// GetWidth returns the width the last layout gave the cell, frame included. A content
// generator receives the content area instead.
func (c *Cell) GetWidth() int {
	return c.width
}

// GetHeight returns the height the last layout gave the cell, frame included. A content
// generator receives the content area instead.
func (c *Cell) GetHeight() int {
	return c.height
}

// GetBounds returns the rectangle the last layout gave the cell, frame included. It
// counts columns and lines from the top left corner of the cell's box, so the cell at
// the corner of a box without a frame starts at (0, 0).
func (c *Cell) GetBounds() image.Rectangle {
	return image.Rect(c.x, c.y, c.x+max(c.width, 0), c.y+max(c.height, 0))
}

// GetContentBounds returns the rectangle the content has, inside the cell's margins,
// border and padding, in the coordinates of GetBounds. It is the area a content
// generator fills, and where to place what the content positions itself, such as a
// cursor or a box nested in the cell.
func (c *Cell) GetContentBounds() image.Rectangle {
	width, height := c.contentSize()
	x, y := c.style.frame.inner(c.x, c.y)
	return image.Rect(x, y, x+width, y+height)
}

// contentSize returns the room the content has, inside the cell's margins, border and
// padding.
func (c *Cell) contentSize() (width, height int) {
	f := c.style.frame
	return max(c.width-f.edgeX-f.paddingX, 0), max(c.height-f.edgeY()-f.paddingY, 0)
}

// render returns the cell as a block of exactly its width by its height, building it
// only when its size, style or content changed since the last render.
//
// The frame (margins and border) renders complete whenever the cell has room for it:
// the content is wrapped the way lipgloss wraps it and clipped to the lines left between
// the paddings, so content taller than the cell loses its bottom lines, never the border.
func (c *Cell) render(from inherited) *block {
	c.style.resolve(from)

	width, height := max(c.width, 0), max(c.height, 0)
	var content string
	if width > 0 && height > 0 && c.contentGenerator != nil {
		content = c.contentGenerator(c.contentSize())
	}
	key := cellKey{width: width, height: height, style: c.style.id, content: content}
	if c.rendered && key == c.key {
		return &c.out
	}
	c.key, c.rendered = key, true

	switch {
	case width == 0 || height == 0:
		c.out.collapse(width, height)
	case c.style.plain():
		c.out.fill(fitPlain(content, width, height), width, height, 0, 0)
	default:
		f := c.style.frame
		style, fitted := c.style.dress(fitContent(content, width-f.edgeX-f.paddingX, height-f.edgeY()-f.paddingY))
		text := renderStyled(f.sized(style, width, height), fitted)
		c.out.fill(text, width, height, f.top, f.bottom)
	}
	return &c.out
}

// copy returns a copy of the cell that shares no render state with it.
func (c *Cell) copy() Cell {
	cellCopy := *c
	cellCopy.out, cellCopy.rendered = block{}, false
	return cellCopy
}

// appendCells appends the cells that are not nil to dst and gives every cell without an
// ID its index as its ID.
func appendCells(dst []*Cell, cells []*Cell) []*Cell {
	dst = appendElements(dst, cells)
	for i, c := range dst {
		if c.id == "" {
			c.id = strconv.Itoa(i)
		}
	}
	return dst
}

// cellWithID returns the first of cells with the ID, or nil.
func cellWithID(cells []*Cell, id string) *Cell {
	for _, c := range cells {
		if c.id == id {
			return c
		}
	}
	return nil
}

// copyCells returns copies of cells.
func copyCells(cells []*Cell) []*Cell {
	copies := make([]*Cell, len(cells))
	for i, c := range cells {
		cellCopy := c.copy()
		copies[i] = &cellCopy
	}
	return copies
}

// fitContent prepares content for a cell: it wraps the content at width and keeps the
// first height lines. Tabs become four spaces and Windows line ends plain ones.
//
// Lipgloss wraps what it renders once more, at the same width. Content prepared here
// already fits, so that second wrap changes nothing; left to itself it breaks some lines
// too late and loses what sticks out past the cell.
func fitContent(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	if strings.ContainsAny(content, "\t\r") {
		content = strings.ReplaceAll(content, "\r\n", "\n")
		content = strings.ReplaceAll(content, "\t", "    ")
	}
	// the wrap drops the spaces at the very end of its input; a newline after the
	// last line keeps them. After that newline the wrap writes nothing but the escape
	// sequences of a color or link the content left open, which go with it.
	wrapped := cellbuf.Wrap(content+"\n", width, "")
	if last := strings.LastIndexByte(wrapped, '\n'); last >= 0 && ansi.Strip(wrapped[last+1:]) == "" {
		wrapped = wrapped[:last]
	}
	return cutLines(firstLines(wrapped, height), width)
}

// cutLines cuts every line of text that is wider than width, which a wrap leaves when
// a single character is: a wide character in a cell of one column.
func cutLines(text string, width int) string {
	fitting := true
	for line := range strings.SplitSeq(text, "\n") {
		if lineWidth(line) > width {
			fitting = false
			break
		}
	}
	if fitting {
		return text
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if lineWidth(line) > width {
			lines[i] = ansi.Truncate(line, width, "")
		}
	}
	return strings.Join(lines, "\n")
}

// fitPlain is fitContent for a cell with the empty style. Content that fits as it is
// skips the wrap.
func fitPlain(content string, width, height int) string {
	if fits(content, width) {
		return content
	}
	return fitContent(content, width, height)
}

// fits reports whether every line of text is printable ASCII and at most width columns
// wide. Wrapping leaves such text as it is.
func fits(text string, width int) bool {
	column := 0
	for i := 0; i < len(text); i++ {
		switch c := text[i]; {
		case c == '\n':
			column = 0
		case c < ' ' || c > '~':
			return false
		default:
			if column++; column > width {
				return false
			}
		}
	}
	return true
}

// firstLines returns at most n lines of s, for n of 1 or more.
func firstLines(s string, n int) string {
	end := 0
	for i := 0; i < n; i++ {
		next := strings.IndexByte(s[end:], '\n')
		if next < 0 {
			return s
		}
		end += next + 1
	}
	return s[:end-1]
}
