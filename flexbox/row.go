package flexbox

import (
	"image"

	"charm.land/lipgloss/v2"
)

// Row holds the cells of one row of a FlexBox. Its height comes from the box; its cells'
// widths and heights come from their ratios and fixed sizes. Rows stack vertically.
type Row struct {
	style styling
	cells []*Cell

	// width and height are the size the last layout gave the row, x and y where it put
	// the row's top left corner in the box, and cellsWidth the columns its cells take
	width, height int
	x, y          int
	cellsWidth    int
	// fixedHeight replaces the ratio-based height when greater than 0
	fixedHeight int

	// out is the last render, key and revs what it was rendered from
	out      block
	key      containerKey
	revs     []uint64
	rendered bool

	// tracks and parts are reused between renders
	tracks []track
	parts  []*block
}

// AddCells appends cells to the row. A cell without an ID gets its index as its ID,
// and a nil cell is left out.
func (r *Row) AddCells(cells ...*Cell) *Row {
	r.cells = appendCells(r.cells, cells)
	return r
}

// SetFixedHeight gives the row a fixed height instead of a ratio-based one. A value of 0
// or less restores ratio sizing.
func (r *Row) SetFixedHeight(value int) *Row {
	r.fixedHeight = value
	return r
}

// CellsLen returns the number of cells.
func (r *Row) CellsLen() int {
	return len(r.cells)
}

// GetCell returns the cell at the index, or nil.
func (r *Row) GetCell(index int) *Cell {
	if index >= 0 && index < len(r.cells) {
		return r.cells[index]
	}
	return nil
}

// GetCellCopy returns a copy of the cell at the index, or nil. Changes to the copy do
// not reach the row.
func (r *Row) GetCellCopy(index int) *Cell {
	if index >= 0 && index < len(r.cells) {
		c := r.cells[index].copy()
		return &c
	}
	return nil
}

// GetCellWithID returns the first cell with the ID, or nil.
func (r *Row) GetCellWithID(id string) *Cell {
	return cellWithID(r.cells, id)
}

// UpdateCellWithIndex replaces the cell at the index. An index out of range or a nil
// cell changes nothing.
func (r *Row) UpdateCellWithIndex(index int, cell *Cell) *Row {
	if cell != nil && index >= 0 && index < len(r.cells) {
		r.cells[index] = cell
		r.cells = appendCells(r.cells[:0], r.cells)
	}
	return r
}

// SetStyle replaces the style. Width and height keys are removed because the box owns
// the row's size. A border, padding or margin on the row shrinks the area its cells share.
func (r *Row) SetStyle(style lipgloss.Style) *Row {
	r.style.set(style)
	return r
}

// StylePassing sets whether the row's style, including what it inherits from the box, is
// passed to its cells.
func (r *Row) StylePassing(value bool) *Row {
	r.style.passing = value
	return r
}

// UseStyle sets the style and enables StylePassing in one call.
func (r *Row) UseStyle(style lipgloss.Style) *Row {
	return r.SetStyle(style).StylePassing(true)
}

// GetWidth returns the width the last layout gave the row, frame included.
func (r *Row) GetWidth() int {
	return r.width
}

// GetHeight returns the height the last layout gave the row, frame included.
func (r *Row) GetHeight() int {
	return r.height
}

// GetBounds returns the rectangle the row takes in its box, frame included, counted
// from the box's top left corner. A row of fixed-width cells is as wide as its cells,
// which can be less than GetWidth.
func (r *Row) GetBounds() image.Rectangle {
	return image.Rect(r.x, r.y, r.x+r.shownWidth(), r.y+max(r.height, 0))
}

// track returns how the row shares the height of its box: the largest ratioY, minimum
// height and fixed height among its cells, and its own fixed height.
func (r *Row) track() track {
	t := track{fixed: r.fixedHeight}
	for _, c := range r.cells {
		t.ratio = max(t.ratio, c.ratioY)
		t.minimum = max(t.minimum, c.minHeight)
		t.fixed = max(t.fixed, c.fixedHeight)
	}
	return t
}

// contentSize returns the area the cells share, inside the row's frame.
func (r *Row) contentSize() (width, height int) {
	f := r.style.frame
	return max(r.width-f.edgeX-f.paddingX, 0), max(r.height-f.edgeY()-f.paddingY, 0)
}

// layout sizes the cells for the row's size and the style resolved from the box. Cells
// share the width as tracks; each cell's height is its ratioY share of the row, the
// largest ratioY taking all of it.
func (r *Row) layout(from inherited) {
	r.style.resolve(from)
	width, height := r.contentSize()

	tracks, largest := r.tracks[:0], 0
	for _, c := range r.cells {
		tracks = append(tracks, track{ratio: c.ratioX, minimum: c.minWidth, fixed: c.fixedWidth})
		largest = max(largest, c.ratioY)
	}
	distribute(tracks, width)
	passed := r.style.passed()
	r.cellsWidth = 0
	for i, c := range r.cells {
		c.style.resolve(passed)
		tracks[i].cross = scale(c.ratioY, largest, height)
		c.width, c.height = tracks[i].size, tracks[i].cross
		r.cellsWidth += max(c.width, 0)
	}
	r.tracks = tracks
}

// laidOut reports whether every cell still has the size the last layout of this row
// gave it. A cell that sits in another row too has that row's size after the box laid
// out all of its rows.
func (r *Row) laidOut() bool {
	for i, c := range r.cells {
		if c.width != r.tracks[i].size || c.height != r.tracks[i].cross {
			return false
		}
	}
	return true
}

// shownWidth returns the columns the rendered row takes: its cells, cut at the row's
// edge, inside its frame, and never more than the row's width.
func (r *Row) shownWidth() int {
	width, _ := r.contentSize()
	f := r.style.frame
	return max(min(min(r.cellsWidth, width)+f.edgeX+f.paddingX, r.width), 0)
}

// contentBounds returns the rectangle the cells of the row show in.
func (r *Row) contentBounds() image.Rectangle {
	x, y := r.style.frame.inner(r.x, r.y)
	width, height := r.contentSize()
	return image.Rect(x, y, x+min(r.cellsWidth, width), y+height)
}

// place puts the row's top left corner at x and y in the box and its cells side by side
// inside its frame.
func (r *Row) place(x, y int) {
	r.x, r.y = x, y
	x, y = r.style.frame.inner(x, y)
	for _, c := range r.cells {
		c.x, c.y = x, y
		x += max(c.width, 0)
	}
}

// render returns the row as a block of exactly its height and at most its width,
// building it only when its size, its style or one of its cells changed since the last
// render.
//
// The cells are joined side by side, tops aligned, and anything wider than the row, such
// as a fixed-width cell wider than the box, is cut at the row's edge, never wrapped. A
// row narrower than the box keeps its width so the box can align it.
func (r *Row) render(from inherited) *block {
	// the box laid the row out already; a cell that sits in another row too has that
	// row's size by now and gets this row's back
	if !r.laidOut() {
		r.layout(from)
	}

	key := containerKey{width: max(r.width, 0), height: max(r.height, 0), style: r.style.id}
	stale := !r.rendered || key != r.key || len(r.revs) != len(r.cells)
	r.revs, r.parts = resize(r.revs, len(r.cells)), r.parts[:0]
	passed := r.style.passed()
	for i, c := range r.cells {
		part := c.render(passed)
		if part.rev != r.revs[i] {
			r.revs[i], stale = part.rev, true
		}
		r.parts = append(r.parts, part)
	}
	if !stale {
		return &r.out
	}
	r.key, r.rendered = key, true

	if key.height == 0 {
		r.out.collapse(key.width, 0)
		return &r.out
	}
	width, height := r.contentSize()
	r.out.join(r.parts, height, 0, 0)
	r.out.clip(width)
	if !r.style.plain() {
		content := r.out.String()
		if len(r.out.lines) == 0 {
			// a row too short for its cells still frames their width
			content = spaces(r.out.width)
		}
		f := r.style.frame
		style, content := r.style.dress(content)
		text := style.Height(key.height - f.marginY).Render(content)
		// a frame wider than the row is cut at the row's edge like its cells
		r.out.fill(text, min(textWidth(text), key.width), key.height, f.top, f.bottom)
	}
	return &r.out
}

// copy returns a copy of the row, with copies of its cells, that shares no render state
// with it.
func (r *Row) copy() Row {
	rowCopy := *r
	rowCopy.cells = copyCells(r.cells)
	rowCopy.out, rowCopy.rendered = block{}, false
	rowCopy.revs, rowCopy.tracks, rowCopy.parts = nil, nil, nil
	return rowCopy
}
