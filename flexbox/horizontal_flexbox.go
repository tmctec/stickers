package flexbox

import (
	"image"

	"charm.land/lipgloss/v2"
)

// HorizontalFlexBox is a responsive grid of columns, each holding cells stacked
// vertically. Columns sit side by side and share the box width by ratio or fixed size.
//
// A HorizontalFlexBox is not safe for concurrent use.
type HorizontalFlexBox struct {
	style   styling
	columns []*Column

	width, height int
	// fixedColumnWidth pins every column to a width when greater than 0
	fixedColumnWidth int
	// columnAlign places the columns when together they are narrower than the box
	columnAlign lipgloss.Position

	// out is the last render, key and revs what it was rendered from
	out      string
	key      containerKey
	revs     []uint64
	rendered bool

	// tracks, parts and frame are reused between renders
	tracks []track
	parts  []*block
	frame  block
}

// NewHorizontal initializes a HorizontalFlexBox of the given size with no columns and an
// empty style.
func NewHorizontal(width, height int) *HorizontalFlexBox {
	return &HorizontalFlexBox{
		style:       styling{own: lipgloss.NewStyle()},
		width:       width,
		height:      height,
		columnAlign: lipgloss.Left,
	}
}

// SetColumnAlign places the columns inside the box when together they are narrower than
// it, such as columns of a fixed width, at lipgloss.Left, lipgloss.Center or
// lipgloss.Right. The default is lipgloss.Left.
func (b *HorizontalFlexBox) SetColumnAlign(align lipgloss.Position) *HorizontalFlexBox {
	b.columnAlign = align
	return b
}

// SetStyle replaces the style. Width and height keys are removed because the box's size
// is set with SetWidth and SetHeight. A border, padding or margin shrinks the area the
// columns share.
func (b *HorizontalFlexBox) SetStyle(style lipgloss.Style) *HorizontalFlexBox {
	b.style.set(style)
	return b
}

// StylePassing sets whether the box's style is passed to its columns. A column passes it
// on to its cells only when its own StylePassing is enabled.
func (b *HorizontalFlexBox) StylePassing(value bool) *HorizontalFlexBox {
	b.style.passing = value
	return b
}

// UseStyle sets the style and enables StylePassing in one call.
func (b *HorizontalFlexBox) UseStyle(style lipgloss.Style) *HorizontalFlexBox {
	return b.SetStyle(style).StylePassing(true)
}

// NewColumn initializes an empty Column. Add it with AddColumns or SetColumns.
func (b *HorizontalFlexBox) NewColumn() *Column {
	return &Column{style: styling{own: lipgloss.NewStyle()}}
}

// AddColumns appends columns to the box. A nil column is left out.
func (b *HorizontalFlexBox) AddColumns(columns []*Column) *HorizontalFlexBox {
	b.columns = appendElements(b.columns, columns)
	return b
}

// SetColumns replaces all columns of the box. A nil column is left out. The box keeps
// its own list, so later changes to the slice do not reach it.
func (b *HorizontalFlexBox) SetColumns(columns []*Column) *HorizontalFlexBox {
	b.columns = replaceElements(b.columns, columns)
	return b
}

// ColumnsLen returns the number of columns.
func (b *HorizontalFlexBox) ColumnsLen() int {
	return len(b.columns)
}

// GetColumn returns the column at the index, or nil.
func (b *HorizontalFlexBox) GetColumn(index int) *Column {
	if index >= 0 && index < len(b.columns) {
		return b.columns[index]
	}
	return nil
}

// GetColumnCopy returns a copy of the column at the index, with copies of its cells, or
// nil. Changes to the copy do not reach the box.
func (b *HorizontalFlexBox) GetColumnCopy(index int) *Column {
	if index >= 0 && index < len(b.columns) {
		columnCopy := b.columns[index].copy()
		return &columnCopy
	}
	return nil
}

// GetColumnCellCopy returns a copy of the cell at cellIndex in the column at
// columnIndex, or nil. Changes to the copy do not reach the box.
func (b *HorizontalFlexBox) GetColumnCellCopy(columnIndex, cellIndex int) *Cell {
	if column := b.GetColumn(columnIndex); column != nil {
		return column.GetCellCopy(cellIndex)
	}
	return nil
}

// UpdateColumn replaces the column at the index. An index out of range or a nil column
// changes nothing.
func (b *HorizontalFlexBox) UpdateColumn(index int, column *Column) *HorizontalFlexBox {
	if column != nil && index >= 0 && index < len(b.columns) {
		b.columns[index] = column
	}
	return b
}

// LockColumnWidth pins every column to the given number of columns, which disables
// horizontal scaling. A value of 0 or less restores ratio sizing.
func (b *HorizontalFlexBox) LockColumnWidth(value int) *HorizontalFlexBox {
	b.fixedColumnWidth = value
	return b
}

// SetHeight sets the box height, which every column gets in full.
func (b *HorizontalFlexBox) SetHeight(value int) *HorizontalFlexBox {
	b.height = value
	return b
}

// SetWidth sets the box width, which its columns share.
func (b *HorizontalFlexBox) SetWidth(value int) *HorizontalFlexBox {
	b.width = value
	return b
}

// GetHeight returns the box height.
func (b *HorizontalFlexBox) GetHeight() int {
	return b.height
}

// GetWidth returns the box width.
func (b *HorizontalFlexBox) GetWidth() int {
	return b.width
}

// Render lays out the grid and returns it as exactly GetHeight lines of GetWidth columns.
// Only what changed since the last render is rendered again: a cell when its size, style
// or content changed, a column when one of its cells did. An unchanged grid costs a call
// to each content generator.
func (b *HorizontalFlexBox) Render() string {
	b.arrange()

	key := containerKey{width: max(b.width, 0), height: max(b.height, 0), style: b.style.id, align: b.columnAlign}
	stale := !b.rendered || key != b.key || len(b.revs) != len(b.columns)
	b.revs, b.parts = resize(b.revs, len(b.columns)), b.parts[:0]
	passed := b.style.passed()
	for i, column := range b.columns {
		part := column.render(passed)
		if part.rev != b.revs[i] {
			b.revs[i], stale = part.rev, true
		}
		b.parts = append(b.parts, part)
	}
	if !stale {
		return b.out
	}
	b.key, b.rendered = key, true

	// Columns that leave spare room are placed in it; columns wider than the box are
	// cut at its edge.
	width, height := b.contentSize()
	left, right := b.spare()
	b.frame.join(b.parts, height, left, right)
	b.frame.clip(width)
	if !b.style.plain() {
		f := b.style.frame
		style, content := b.style.dress(b.frame.String())
		text := renderStyled(f.sized(style, key.width, key.height), content)
		b.frame.fill(text, key.width, key.height, f.top, f.bottom)
	}
	b.out = b.frame.String()
	return b.out
}

// ForceRecalculate lays out the box and every column now, so the sizes and positions of
// columns and cells can be read before the next render.
func (b *HorizontalFlexBox) ForceRecalculate() {
	b.arrange()
}

// arrange sizes and places the columns and their cells. Render does it before anything
// renders, so a content generator that reads a position reads the one of this render.
func (b *HorizontalFlexBox) arrange() {
	b.layout()
	passed := b.style.passed()
	for _, column := range b.columns {
		column.layout(passed)
	}
	b.place()
}

// CellAt returns the cell that shows at column x and line y of the box, counted from the
// box's top left corner, or nil when no cell shows there: on a frame, beside a cell
// narrower than its column, or outside the box. It answers for the box as it is now,
// which is what the last render showed unless something changed since.
func (b *HorizontalFlexBox) CellAt(x, y int) *Cell {
	b.arrange()
	point := image.Pt(x, y)
	inside := b.contentBounds()
	for _, column := range b.columns {
		if !point.In(column.contentBounds().Intersect(inside)) {
			continue
		}
		for _, cell := range column.cells {
			if point.In(cell.GetBounds()) {
				return cell
			}
		}
	}
	return nil
}

// contentBounds returns the rectangle the columns show in, inside the box's frame.
func (b *HorizontalFlexBox) contentBounds() image.Rectangle {
	x, y := b.style.frame.inner(0, 0)
	width, height := b.contentSize()
	return image.Rect(x, y, x+width, y+height)
}

// spare returns the columns left and right of the columns of the box when together they
// are narrower than it, split the way the column alignment says.
func (b *HorizontalFlexBox) spare() (left, right int) {
	right, _ = b.contentSize()
	for _, column := range b.columns {
		right -= max(column.width, 0)
	}
	if right <= 0 {
		return 0, 0
	}
	left = right - offset(right, 1-b.columnAlign)
	return left, right - left
}

// place positions the columns and their cells in the box, the way Render joins them.
func (b *HorizontalFlexBox) place() {
	x, y := b.style.frame.inner(0, 0)
	left, _ := b.spare()
	x += left
	for _, column := range b.columns {
		column.place(x, y)
		x += max(column.width, 0)
	}
}

// contentSize returns the area the columns share, inside the box's frame.
func (b *HorizontalFlexBox) contentSize() (width, height int) {
	f := b.style.frame
	return max(b.width-f.edgeX-f.paddingX, 0), max(b.height-f.edgeY()-f.paddingY, 0)
}

// layout sizes the columns. Every column takes the content height, and the columns
// share the content width as tracks.
func (b *HorizontalFlexBox) layout() {
	b.style.resolve(inherited{})
	width, height := b.contentSize()

	tracks := b.tracks[:0]
	for _, column := range b.columns {
		t := column.track()
		if b.fixedColumnWidth > 0 {
			t.fixed = b.fixedColumnWidth
		}
		tracks = append(tracks, t)
	}
	distribute(tracks, width)
	for i, column := range b.columns {
		column.width, column.height = tracks[i].size, height
	}
	b.tracks = tracks
}
