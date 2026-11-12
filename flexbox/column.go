package flexbox

import (
	"image"

	"charm.land/lipgloss/v2"
)

// Column holds the cells of one column of a HorizontalFlexBox. Its width comes from the
// box; its cells' widths and heights come from their ratios and fixed sizes. Columns sit
// side by side.
type Column struct {
	style styling
	cells []*Cell

	// width and height are the size the last layout gave the column, x and y where it
	// put the column's top left corner in the box
	width, height int
	x, y          int
	// fixedWidth replaces the ratio-based width when greater than 0
	fixedWidth int

	// out is the last render, key and revs what it was rendered from
	out      block
	key      containerKey
	revs     []uint64
	rendered bool

	// tracks and parts are reused between renders
	tracks []track
	parts  []*block
}

// AddCells appends cells to the column. A cell without an ID gets its index as its ID,
// and a nil cell is left out.
func (c *Column) AddCells(cells ...*Cell) *Column {
	c.cells = appendCells(c.cells, cells)
	return c
}

// SetFixedWidth gives the column a fixed width instead of a ratio-based one. A value of
// 0 or less restores ratio sizing.
func (c *Column) SetFixedWidth(value int) *Column {
	c.fixedWidth = value
	return c
}

// CellsLen returns the number of cells.
func (c *Column) CellsLen() int {
	return len(c.cells)
}

// GetCell returns the cell at the index, or nil.
func (c *Column) GetCell(index int) *Cell {
	if index >= 0 && index < len(c.cells) {
		return c.cells[index]
	}
	return nil
}

// GetCellCopy returns a copy of the cell at the index, or nil. Changes to the copy do
// not reach the column.
func (c *Column) GetCellCopy(index int) *Cell {
	if index >= 0 && index < len(c.cells) {
		cellCopy := c.cells[index].copy()
		return &cellCopy
	}
	return nil
}

// GetCellWithID returns the first cell with the ID, or nil.
func (c *Column) GetCellWithID(id string) *Cell {
	return cellWithID(c.cells, id)
}

// UpdateCellWithIndex replaces the cell at the index. An index out of range or a nil
// cell changes nothing.
func (c *Column) UpdateCellWithIndex(index int, cell *Cell) *Column {
	if cell != nil && index >= 0 && index < len(c.cells) {
		c.cells[index] = cell
		c.cells = appendCells(c.cells[:0], c.cells)
	}
	return c
}

// SetStyle replaces the style. Width and height keys are removed because the box owns
// the column's size. A border, padding or margin on the column shrinks the area its
// cells share.
func (c *Column) SetStyle(style lipgloss.Style) *Column {
	c.style.set(style)
	return c
}

// StylePassing sets whether the column's style, including what it inherits from the
// box, is passed to its cells.
func (c *Column) StylePassing(value bool) *Column {
	c.style.passing = value
	return c
}

// UseStyle sets the style and enables StylePassing in one call.
func (c *Column) UseStyle(style lipgloss.Style) *Column {
	return c.SetStyle(style).StylePassing(true)
}

// GetWidth returns the width the last layout gave the column, frame included.
func (c *Column) GetWidth() int {
	return c.width
}

// GetHeight returns the height the last layout gave the column, frame included.
func (c *Column) GetHeight() int {
	return c.height
}

// GetBounds returns the rectangle the column takes in its box, frame included, counted
// from the box's top left corner.
func (c *Column) GetBounds() image.Rectangle {
	return image.Rect(c.x, c.y, c.x+max(c.width, 0), c.y+max(c.height, 0))
}

// track returns how the column shares the width of its box: the largest ratioX, minimum
// width and fixed width among its cells, and its own fixed width.
func (c *Column) track() track {
	t := track{fixed: c.fixedWidth}
	for _, cell := range c.cells {
		t.ratio = max(t.ratio, cell.ratioX)
		t.minimum = max(t.minimum, cell.minWidth)
		t.fixed = max(t.fixed, cell.fixedWidth)
	}
	return t
}

// contentSize returns the area the cells share, inside the column's frame.
func (c *Column) contentSize() (width, height int) {
	f := c.style.frame
	return max(c.width-f.edgeX-f.paddingX, 0), max(c.height-f.edgeY()-f.paddingY, 0)
}

// layout sizes the cells for the column's size and the style resolved from the box.
// Cells share the height as tracks; each cell's width is its ratioX share of the column,
// the largest ratioX taking all of it.
func (c *Column) layout(from inherited) {
	c.style.resolve(from)
	width, height := c.contentSize()

	tracks, largest := c.tracks[:0], 0
	for _, cell := range c.cells {
		tracks = append(tracks, track{ratio: cell.ratioY, minimum: cell.minHeight, fixed: cell.fixedHeight})
		largest = max(largest, cell.ratioX)
	}
	distribute(tracks, height)
	passed := c.style.passed()
	for i, cell := range c.cells {
		cell.style.resolve(passed)
		tracks[i].cross = scale(cell.ratioX, largest, width)
		cell.width, cell.height = tracks[i].cross, tracks[i].size
	}
	c.tracks = tracks
}

// laidOut reports whether every cell still has the size the last layout of this column
// gave it. A cell that sits in another column too has that column's size after the box
// laid out all of its columns.
func (c *Column) laidOut() bool {
	for i, cell := range c.cells {
		if cell.width != c.tracks[i].cross || cell.height != c.tracks[i].size {
			return false
		}
	}
	return true
}

// contentBounds returns the rectangle the cells of the column show in.
func (c *Column) contentBounds() image.Rectangle {
	x, y := c.style.frame.inner(c.x, c.y)
	width, height := c.contentSize()
	return image.Rect(x, y, x+width, y+height)
}

// place puts the column's top left corner at x and y in the box and its cells on top of
// each other inside its frame. A column with a style moves the cells as a group to where
// the style aligns them, when they leave columns or lines spare.
func (c *Column) place(x, y int) {
	c.x, c.y = x, y
	width, height := c.contentSize()
	widest, lines := 0, 0
	for _, cell := range c.cells {
		if cell.height > 0 {
			widest, lines = max(widest, cell.width), lines+cell.height
		}
	}
	f := c.style.frame
	x, y = f.inner(x, y)
	x, y = x+lead(width-widest, f.alignX), y+lead(height-lines, f.alignY)
	for _, cell := range c.cells {
		cell.x, cell.y = x, y
		y += max(cell.height, 0)
	}
}

// render returns the column as a block of exactly its width by its height, building it
// only when its size, its style or one of its cells changed since the last render.
//
// The cells are stacked, left edges aligned. A cell without lines takes no room, and
// anything taller than the column is cut at the column's bottom edge.
func (c *Column) render(from inherited) *block {
	// the box laid the column out already; a cell that sits in another column too has
	// that column's size by now and gets this column's back
	if !c.laidOut() {
		c.layout(from)
	}

	key := containerKey{width: max(c.width, 0), height: max(c.height, 0), style: c.style.id}
	stale := !c.rendered || key != c.key || len(c.revs) != len(c.cells)
	c.revs, c.parts = resize(c.revs, len(c.cells)), c.parts[:0]
	passed := c.style.passed()
	for i, cell := range c.cells {
		part := cell.render(passed)
		if part.rev != c.revs[i] {
			c.revs[i], stale = part.rev, true
		}
		c.parts = append(c.parts, part)
	}
	if !stale {
		return &c.out
	}
	c.key, c.rendered = key, true

	if key.width == 0 || key.height == 0 {
		c.out.collapse(key.width, key.height)
		return &c.out
	}
	width, height := c.contentSize()
	if c.style.plain() {
		c.out.stack(c.parts, width, width, lipgloss.Left, height)
		c.out.extend(height)
		return &c.out
	}
	group := widest(c.parts)
	c.out.stack(c.parts, group, group, lipgloss.Left, height)
	f := c.style.frame
	style, content := c.style.dress(c.out.String())
	text := renderStyled(f.sized(style, key.width, key.height), content)
	c.out.fill(text, key.width, key.height, f.top, f.bottom)
	return &c.out
}

// copy returns a copy of the column, with copies of its cells, that shares no render
// state with it.
func (c *Column) copy() Column {
	columnCopy := *c
	columnCopy.cells = copyCells(c.cells)
	columnCopy.out, columnCopy.rendered = block{}, false
	columnCopy.revs, columnCopy.tracks, columnCopy.parts = nil, nil, nil
	return columnCopy
}
