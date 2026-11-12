// Package flexbox is a responsive grid for terminal apps, inspired by CSS flexbox and
// rendered with lipgloss.
//
// A FlexBox has a width and a height, holds rows, and each row holds cells. A
// HorizontalFlexBox holds columns of cells instead. Give the box the terminal size, give
// rows and cells ratios or fixed sizes, and call Render.
//
//	box := flexbox.New(width, height)
//	box.AddRows([]*flexbox.Row{
//		box.NewRow().AddCells(
//			flexbox.NewCell(1, 1).SetContent("sidebar").SetFixedWidth(20),
//			flexbox.NewCell(3, 1).SetContent("main"),
//		),
//		box.NewRow().AddCells(flexbox.NewCell(1, 1).SetContent("footer")).SetFixedHeight(3),
//	})
//	s := box.Render()
//
// Every setter returns its receiver, so calls chain, and takes effect on the next Render.
//
// # Sizing
//
// Rows share the height of a FlexBox, and the cells of a row share its width, the same
// way. Fixed sizes are taken first. What is left is split in proportion to the ratios,
// each share rounded down and the leftover going one unit each to the largest ratios. A
// share under its minimum takes the minimum and the others split the rest. When the
// fixed sizes leave nothing, the ratio-sized entries get their minimum, or nothing
// without one.
//
// A cell's ratioX, SetFixedWidth and SetMinWidth size it along its row. Across the row,
// each cell is ratioY/maxRatioY of the row height, rounded down, and sits at the top;
// the largest ratioY, SetFixedHeight and SetMinHeight among the cells size the row
// itself, as Row.SetFixedHeight does.
//
// A HorizontalFlexBox is the mirror image. Columns share the width of the box and the
// cells of a column share its height: ratioY, SetFixedHeight and SetMinHeight size a
// cell along its column, each cell is ratioX/maxRatioX of the column width, and the
// largest ratioX, SetFixedWidth and SetMinWidth among the cells size the column, as
// Column.SetFixedWidth does.
//
// # Rendering
//
// Render returns exactly the box's height in lines, each the box's width in columns.
//
// A cell renders its frame first and its content second. The border and margins render
// complete whenever the cell has room for them; content is wrapped at the content width
// and clipped to the lines between the paddings, so content taller than the cell loses
// its bottom lines, never the border. A row or cell without lines takes no lines, and a
// cell or column without columns takes no columns. A row never wraps its cells: anything
// wider than the row is cut at the row's edge, and a row narrower than the box keeps its
// width so SetRowAlign can place it. Columns that together are narrower than a
// HorizontalFlexBox are placed by SetColumnAlign, and cut at its edge when they are
// wider.
//
// Render lays out the grid every time and renders again only what changed since the
// last call: a cell when its size, style or content changed, a row or column when one
// of its cells did, the box when one of its rows or columns did. A content generator
// runs on every render, so content that reads outside state is always current.
//
// # Styles
//
// Styles are lipgloss values. Width and height keys are removed when a style is set,
// because the grid owns sizes. StylePassing on a box passes its style to the rows, and
// on a row or column to the cells; UseStyle sets a style and enables passing in one call.
// Children inherit only the keys they have not set; margins and padding never pass.
//
// # Reading the grid
//
// GetRow, GetColumn, GetCell and GetCellWithID return the elements themselves; a change
// to one shows on the next render. GetRowCopy, GetRowCellCopy, GetColumnCopy,
// GetColumnCellCopy and GetCellCopy return copies that leave the grid alone. Sizes are
// computed by Render, or by ForceRecalculate when they are needed before it.
//
// Add each row, column and cell in one place: one that sits in two places has one size,
// the size of its last place. A nil row, column or cell is left out wherever one is
// added or set.
//
// # Positions
//
// GetBounds on a cell, row or column returns the rectangle the last layout gave it,
// counted in columns and lines from the top left corner of its box. GetContentBounds on
// a cell returns the rectangle its content has, inside its margins, border and padding.
// CellAt returns the cell that shows at a point, which turns the position of a mouse
// click into a cell.
//
// The coordinates belong to the box. A box that renders somewhere else than at the top
// left corner of the screen has its own position to add: a box nested in a cell starts
// at the GetContentBounds of that cell.
//
// # Nesting
//
// A content generator receives the room its content has, inside the cell's margins,
// border and padding. A box, a table or any other component nests in a cell by taking
// that size and rendering:
//
//	cell.SetContentGenerator(func(width, height int) string {
//		return inner.SetWidth(width).SetHeight(height).Render()
//	})
//
// # Concurrency
//
// A box, its rows, columns and cells are not safe for concurrent use.
package flexbox
