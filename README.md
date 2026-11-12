# Stickers 👾

Stickers is a set of layout elements for terminal apps built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea) v2 and styled with
[Lip Gloss](https://github.com/charmbracelet/lipgloss) v2. It ships two packages:

- `flexbox`: a responsive grid inspired by CSS flexbox. A `FlexBox` stacks rows
  of cells; a `HorizontalFlexBox` lines up columns of cells. Sizes come from
  ratios, fixed values, or both, and follow the terminal as it resizes.
- `table`: a scrollable, sortable, filterable table that uses `flexbox` for its
  layout.

Both render only what changed since the last frame, so calling `Render` from
every `View` is cheap. See [Performance](#performance).

This module lives at `github.com/tmctec/stickers/v2`. It is a fork of
[76creates/stickers](https://github.com/76creates/stickers).

![FlexBox with Table](.github/images/flex-box-with-table.gif)

## Contents

- [Install](#install)
- [FlexBox](#flexbox)
  - [Sizing rules](#sizing-rules)
  - [What renders when space runs out](#what-renders-when-space-runs-out)
  - [Content](#content)
  - [Styles](#styles)
  - [Reading and changing the grid](#reading-and-changing-the-grid)
  - [Positions and the mouse](#positions-and-the-mouse)
  - [HorizontalFlexBox](#horizontalflexbox)
  - [Recipes](#recipes)
- [Table](#table)
- [Performance](#performance)
- [Demos](#demos)
- [Development](#development)
- [Security](#security)

## Install

```bash
go get github.com/tmctec/stickers/v2@latest
```

The library requires Go 1.25 or later and takes Lip Gloss v2 styles
(`charm.land/lipgloss/v2`). It has no dependency on Bubble Tea: a box is a
value that turns a size into a string, so it works in a Bubble Tea v2 app and
outside Bubble Tea.

For Lip Gloss v1 and Bubble Tea v1, use the upstream module,
`github.com/76creates/stickers`.

The package documentation, with runnable examples, is on
[pkg.go.dev](https://pkg.go.dev/github.com/tmctec/stickers/v2).

## FlexBox

A `FlexBox` has a width and a height, holds rows, and each row holds cells. You
give the box the terminal size, give rows and cells ratios or fixed sizes, and
call `Render()`.

```go
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tmctec/stickers/v2/flexbox"
)

type model struct{ box *flexbox.FlexBox }

func newModel() model {
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	box := flexbox.New(0, 0)
	box.AddRows([]*flexbox.Row{
		box.NewRow().AddCells(
			flexbox.NewCell(1, 1).SetStyle(border).SetContent("sidebar").SetFixedWidth(20),
			flexbox.NewCell(3, 1).SetStyle(border).SetContent("main"),
		),
		box.NewRow().AddCells(
			flexbox.NewCell(1, 1).SetStyle(border).SetContent("footer"),
		).SetFixedHeight(3),
	})
	return model{box: box}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.box.SetWidth(msg.Width).SetHeight(msg.Height)
	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	view := tea.NewView(m.box.Render())
	view.AltScreen = true
	return view
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
```

Every setter returns its receiver, so calls chain. A setter takes effect on the
next `Render`; nothing else has to be called.

### Sizing rules

Every cell is created with `NewCell(ratioX, ratioY)`.

Rows share the height of a box, and the cells of a row share its width, by the
same four steps:

1. Fixed sizes are taken first.
2. What is left is split in proportion to the ratios. Each share is rounded
   down and the leftover goes one unit each to the largest ratios.
3. A share that falls under its minimum takes the minimum, and the others
   split the rest.
4. When the fixed sizes leave nothing, the ratio-sized entries get their
   minimum, or nothing without one.

**Row heights.**

- `row.SetFixedHeight(n)` pins a row. `0` or less restores ratio sizing.
- `cell.SetFixedHeight(n)` asks for a fixed row height from a cell. The row
  takes the largest request among its cells.
- `cell.SetMinHeight(n)` keeps the cell's row at least `n` lines tall. The row
  takes the largest minimum among its cells.
- A row's ratio is the largest `ratioY` among its cells.
- `box.LockRowHeight(n)` pins every row to `n` lines. `0` or less restores
  ratio sizing.

**Cell widths within a row.**

- `cell.SetFixedWidth(n)` pins a cell. `0` or less restores ratio sizing.
- `cell.SetMinWidth(n)` keeps a ratio-sized cell at least `n` columns wide.
- The other cells split the rest by `ratioX`.

**Cell heights within a row.** Each cell gets `ratioY / maxRatioY` of the row
height, rounded down, and sits at the top of the row. Cells with the largest
ratio fill the row; the others are shorter.

**Rows of different widths.** A row whose cells add up to less than the box
width, because of fixed widths, keeps that width. `box.SetRowAlign(pos)` with
`lipgloss.Left`, `lipgloss.Center` or `lipgloss.Right` places such rows against
the widest row. The default is left.

### What renders when space runs out

`Render` always returns exactly `GetHeight()` lines of `GetWidth()` columns,
whatever the rows and cells ask for.

- A cell renders its frame first and its content second. The border and
  margins render complete whenever the cell has room for them. Content taller
  than the cell loses its bottom lines; the bottom border stays.
- Content is wrapped at the content width and clipped to the lines between the
  paddings.
- A cell too short for its frame shows its top lines. A bordered cell with two
  lines is a closed, empty frame; with one line it is the top border.
- A row without lines takes no line of the box, so the rows that have lines
  keep theirs. A cell without lines leaves its columns blank, and a cell
  without columns adds nothing to its row.
- A row never wraps its cells. Anything wider than the row, such as a
  fixed-width cell wider than the box, is cut at the row's edge.
- Rows that together are taller than the box are cut at its bottom edge.

### Content

- `cell.SetContent(s)` sets static text. A tab shows as four spaces.
- `cell.SetContentGenerator(func(w, h int) string)` builds the text each time
  the box renders. `w` and `h` are the room the content has, inside the cell's
  margins, border and padding, never less than 0, so the generator can lay out
  text for the space it has. A cell without columns or lines does not call its
  generator.
- `cell.GetWidth()` and `cell.GetHeight()` return the full allocated size,
  including the frame.
- Content keeps its alignment: the style's horizontal and vertical alignment
  place it, and trailing spaces count as part of a line.
- Content can carry its own colors and hyperlinks. They end with the cell on
  every line, and a cell or row that underlines or strikes through draws its
  line through styled content too.

### Styles

Styles are `lipgloss.Style` values. Width and height keys are removed when a
style is set; the grid owns sizes.

- `SetStyle(style)` on a box, row, column or cell. Borders, padding and
  margins on a box or row shrink the area its children share.
- `StylePassing(true)` on a box passes its style to its rows; on a row or
  column it passes the style to its cells. A row with passing enabled inherits
  the box style first, then hands the result to its cells. Children inherit
  only keys they have not set themselves; margins and padding never pass.
  Changing a passed style changes the children on the next render.
- `UseStyle(style)` sets the style and enables passing in one call.
- `NewCell(x, y, flexbox.WithStyle(style))` sets a cell style at creation.
- `GetStyle()` returns the style set on the cell itself, without what a row or
  column passes down.
- Styles render in full color. Bubble Tea fits the colors to the terminal;
  outside Bubble Tea, print with `lipgloss.Println`, which does the same.

### Reading and changing the grid

- `GetRow`, `GetCell` and `GetCellWithID` return the elements themselves.
  Change them and the next render shows it.
- `GetRowCopy`, `GetRowCellCopy` and `GetCellCopy` return copies. Changes to a
  copy do not reach the grid.
- `UpdateRow` and `UpdateCellWithIndex` replace an element in place.
- All of them return `nil`, or change nothing, for an index that is not there.
  A `nil` row, column or cell is left out wherever one is added or set.
- Add each row, column and cell in one place. One that sits in two places has
  one size, the size of its last place.
- Cell IDs default to the cell's index in its row; `SetID` names one for
  `GetCellWithID`, and `GetID` reads it back.
- Sizes are computed when the box renders. `box.ForceRecalculate()` computes
  them without rendering, so `GetWidth` and `GetHeight` on rows and cells can
  be read first.

### Positions and the mouse

A layout places every element as well as sizing it. Positions count columns and
lines from the top left corner of the box.

- `GetBounds()` on a cell, row or column returns its rectangle, frame included,
  as an `image.Rectangle`.
- `cell.GetContentBounds()` returns the rectangle the content has, inside the
  cell's margins, border and padding: the area a content generator fills.
- `box.CellAt(x, y)` returns the cell that shows at a point, or `nil` on a
  frame, in a blank, or outside the box.

A mouse click carries the pointer's position, so `CellAt` turns it into a cell:

```go
case tea.MouseClickMsg:
	if cell := m.box.CellAt(msg.X, msg.Y); cell != nil {
		m.selected = cell.GetID()
	}
```

Set `view.MouseMode = tea.MouseModeCellMotion` in `View` to receive mouse
messages. The same positions place the terminal's cursor in a cell:

```go
origin := cell.GetContentBounds().Min
view.Cursor = tea.NewCursor(origin.X, origin.Y)
```

The coordinates belong to the box. A box that does not start at the top left
corner of the screen has its own position to subtract from the pointer's. A box
nested in a cell starts at that cell's `GetContentBounds().Min`.

### HorizontalFlexBox

`NewHorizontal(width, height)` lines up columns side by side; cells stack
vertically inside each column. It mirrors `FlexBox` with the two directions
swapped: columns share the box width, and the cells of a column share its
height.

```go
box := flexbox.NewHorizontal(width, height)
box.AddColumns([]*flexbox.Column{
	box.NewColumn().SetFixedWidth(24).AddCells(
		flexbox.NewCell(1, 1).SetContent("menu"),
	),
	box.NewColumn().AddCells(
		flexbox.NewCell(1, 1).SetContent("title").SetFixedHeight(3),
		flexbox.NewCell(1, 1).SetContent("body"),
	),
})
```

The same cell setters apply in both boxes. Along the direction the cells are
laid out they size the cell; across it they size the cell's row or column,
which takes the largest value among its cells.

| Cell setting | In a `FlexBox` row | In a `HorizontalFlexBox` column |
|---|---|---|
| `ratioX` | share of the row's width | width as a fraction of the column; the largest sets the column's share of the box |
| `ratioY` | height as a fraction of the row; the largest sets the row's share of the box | share of the column's height |
| `SetFixedWidth` | width of the cell | width of the column |
| `SetMinWidth` | minimum width of the cell | minimum width of the column |
| `SetFixedHeight` | height of the row | height of the cell |
| `SetMinHeight` | minimum height of the row | minimum height of the cell |

- `column.SetFixedWidth(n)` pins a column, as `row.SetFixedHeight(n)` pins a
  row. `LockColumnWidth(n)` pins every column to `n` columns.
- `SetColumnAlign(pos)` places the columns inside the box when together they
  are narrower than it, at `lipgloss.Left`, `lipgloss.Center` or
  `lipgloss.Right`. The default is left. Columns that together are wider than
  the box are cut at its edge.
- `GetColumn`, `GetColumnCopy`, `GetColumnCellCopy`, `UpdateColumn`,
  `ColumnsLen` and `CellAt` mirror the `FlexBox` methods.

### Recipes

**Nest a box in a cell.** A content generator receives the cell's content
area. Size the inner box to it and render:

```go
inner := flexbox.NewHorizontal(0, 0)
// add columns to inner

cell := flexbox.NewCell(1, 1).SetContentGenerator(func(w, h int) string {
	return inner.SetWidth(w).SetHeight(h).Render()
})
```

**Put a component in a cell.** The same pattern fits anything that renders to a
size: a `table.Table`, or a Bubble Tea component such as a viewport.

```go
flexbox.NewCell(3, 1).SetContentGenerator(func(w, h int) string {
	return m.table.SetWidth(w).SetHeight(h).Render()
})

flexbox.NewCell(3, 1).SetContentGenerator(func(w, h int) string {
	m.viewport.SetWidth(w)
	m.viewport.SetHeight(h)
	return m.viewport.View()
})
```

**Float a box over the grid.** Lip Gloss layers put one render on top of
another, which keeps the grid visible around a dialog:

```go
dialog := dialogStyle.Render("Save changes?")
w, h := lipgloss.Size(dialog)
layers := lipgloss.NewCompositor(
	lipgloss.NewLayer(m.box.Render()),
	lipgloss.NewLayer(dialog).X((m.width-w)/2).Y((m.height-h)/2).Z(1),
)
content := lipgloss.NewCanvas(m.width, m.height).Compose(layers).Render()
```

**Show live data.** A content generator runs on every render, so content that
reads from the model is always current:

```go
flexbox.NewCell(1, 1).SetContentGenerator(func(_, _ int) string {
	return fmt.Sprintf("%d requests", m.requests)
})
```

**A sidebar, a main area and a status line.**

```go
box := flexbox.New(width, height)
box.AddRows([]*flexbox.Row{
	box.NewRow().AddCells(
		flexbox.NewCell(1, 1).SetContent("sidebar").SetFixedWidth(24),
		flexbox.NewCell(1, 1).SetContent("main").SetMinWidth(40),
	),
	box.NewRow().SetFixedHeight(1).AddCells(
		flexbox.NewCell(1, 1).SetContent("status"),
	),
})
```

## Table

A `Table` renders a one-line header, one-line rows, and a one-line footer with
the cursor position and the size of the rows area. It scrolls vertically with
the cursor and horizontally when the columns' minimum widths do not fit.

```go
t := table.NewTable(0, 0, []string{"ID", "Name", "Score"})
t, err := t.SetTypes(0, "", 0.0) // int, string, float64 columns
if err != nil {
	return err
}
t.SetRatio([]int{1, 4, 2}).SetMinWidth([]int{4, 10, 6})
t.MustAddRows([][]any{
	{1, "Ada", 9.5},
	{2, "Linus", 8.0},
})

// on tea.WindowSizeMsg
t.SetWidth(w)
t.SetHeight(h) // includes the header and footer lines

// on keys
t.CursorDown()       // also CursorUp, CursorLeft, CursorRight
t.OrderByAsc(2)      // or OrderByDesc; GetOrder() reports the current sort
t.SetFilter(1, "li") // case-insensitive substring on column 1; UnsetFilter() clears
row := t.GetCursorRow() // the values of the row under the cursor

// in View
s := t.Render()
```

- **Types.** Every column is `string` until `SetTypes` is called with one zero
  value per column. Accepted types are `string`, `int`, `int8`, `int16`,
  `int32`, `int64`, `float32` and `float64`. `SetTypes` clears existing rows.
- **Rows.** `AddRows` validates each row's length and cell types and returns
  the first error without adding anything: an `ErrorRowLen`, `ErrorBadType` or
  `ErrorBadCellType`. `MustAddRows` panics instead. `ClearRows` empties the
  table and keeps the sort, the filter and the cursor's row, so a reload shows
  the new data the same way.
- **Values.** Numbers show as plain decimals. A value wider than its column is
  cut and ends in `…`; a value of several lines shows its first, and a tab
  shows as four spaces.
- **Sorting.** `OrderByAsc(i)` and `OrderByDesc(i)` sort by column `i` using
  the column's type, so numbers sort numerically. Rows that compare equal keep
  their order, so sorting by one column and then by another orders by both.
  Rows added to a sorted table are sorted in. An unknown index is ignored.
  The header marks the sorted column with ▲ or ▼.
- **Filtering.** `SetFilter(i, s)` keeps rows whose column `i` contains `s`,
  ignoring case. The header marks the filtered column with ⑂ and the footer
  names the filter while the cursor is on that column. `GetFilter` returns the
  column, `-1` when none, and the string.
- **Cursor.** `GetCursorLocation` returns `x, y`, `GetCursorValue` the text
  under the cursor and `GetCursorRow` a copy of the row's values. All three
  count over the rows the filter keeps, in sorted order; without rows the
  value is `""` and the row `nil`. While a filter hides the cursor's row the
  cursor sits on the last row, and it is back on its row when the row is.
- **Mouse.** `CellAt(x, y)` returns the column and the row of the cell at a
  position on the rendered table, and false where there is none.
  `SetCursorLocation(column, row)` moves the cursor there and scrolls to it. A
  click on line 0 is on the header: `ColumnAt(x)` gives the column whose
  header shows there, the column to sort by. `RowAt(y)` gives the row on a
  line. A table that does not start at the top left corner of the screen has
  its position to subtract first.
- **Horizontal scrolling.** Columns are shown while their minimum widths fit
  the table width; moving the cursor past the visible range scrolls.
  `GetVisibleColumnRange` returns the first and last visible column. With all
  minimum widths at `0`, every column is always visible. The cursor's column
  is always visible.
- **Styles.** `SetStyles(map[table.StyleKey]lipgloss.Style)` overrides any of
  `StyleKeyHeader`, `StyleKeyFooter`, `StyleKeyRows`, `StyleKeyRowsSubsequent`
  (every other row), `StyleKeyRowsCursor` and `StyleKeyCellCursor`; keys you
  leave out keep their defaults. `SetStylePassing(true)` passes the row styles
  down to the cells.
- **Size.** `Render` returns exactly the table's height in lines, each the
  table's width in columns, whatever frames the styles have: the header, every
  row and the footer are one line each. A table of one line is the header
  alone; of two, the header and the footer. A footer without room drops the
  filter first and the size second.
- **In a cell.** `SetWidth` and `SetHeight` with the size the table already
  has do nothing, so a content generator can size the table on every render.
- **Sizing checks.** `SetRatio` and `SetMinWidth` must receive exactly one
  value per column, and ratios must be greater than zero. A mismatch panics,
  since it is a programming error.

![Table Multi-Type](.github/images/table-multi-type.gif)

## Performance

Bubble Tea calls `View` after every message, so `Render` runs far more often
than the layout changes. Both packages are built for that:

- A box computes its layout, sizes and positions, on every render with
  integer math and no allocation.
- A cell renders again only when its size, its style or its content changed. A
  row renders again only when one of its cells did, and the box only when one
  of its rows did.
- Rows, columns and boxes without a style of their own are assembled directly,
  without a Lip Gloss pass.
- A table keeps the rows on screen between renders. A cursor move renders the
  two rows it touches, and a scroll renders the row that scrolls in.
- Sorting is a stable O(n log n) sort in place. The filter runs when the
  filter or the rows change, never on a cursor move.

Content generators run on every render, so keep them cheap or cache what they
compute.

Measured on an Apple M2 Pro with `make bench`, for a 160×48 grid of 17 cells
and a 160×48 table in its default colors:

| Benchmark | Time per render |
|---|---|
| Grid, nothing changed | 0.5 µs, no allocation |
| Grid, one cell's content changed | 35 µs |
| Grid, resized, every cell rendered | 300 µs |
| Table of 1,000 rows, nothing changed | 10 µs, no allocation |
| The same table in a flexbox cell, nothing changed | 10 µs, no allocation |
| Table, cursor moved one row | 45 µs |
| Table of 100,000 rows, scrolled one row | 45 µs |
| Table of 100,000 rows, cursor moved under a filter | 43 µs |
| Sort 10,000 rows | 0.8 ms |
| Filter 100,000 rows and render | 2.5 ms |

A box and a table are not safe for concurrent use. Bubble Tea runs `Update`
and `View` on one goroutine, so a model needs no locking.

## Demos

The `example` directory holds thirteen demos and a combined program that cycles
through them. It is a Go module of its own, so the demos' dependencies stay
out of the library's `go.mod`.

```bash
make help       # lists every target, with one line for each demo
make run        # all demos: tab / shift+tab cycle, a about, q quit
make demo-7     # one demo
```

Every demo answers `a` for an about box, which floats over the demo as a Lip
Gloss layer, and `q` to quit. In `make run` the purple header names the demo,
lists its own keys, then the shared keys. Each demo names itself in the
terminal's title.

| Demo | Shows | Keys |
|---|---|---|
| 1 FlexBox Simple | ratio-based rows and cells | |
| 2 FlexBox Horizontal | `HorizontalFlexBox` columns | |
| 3 FlexBox with Table | a `Table` inside a cell through a content generator, clicks placed by the cell's position | arrows, click, wheel, ctrl+s or a click on a header sorts, enter copies the cell into the grid, letters or a paste filter |
| 4 Table Simple String | string columns, sorting, filtering, the mouse | arrows, click, wheel, ctrl+s or a click on a header, enter copies the cell to the clipboard, letters or a paste |
| 5 Table Multi-Type | typed columns via gocsv, numeric sorting, the mouse | arrows, click, wheel, ctrl+s or a click on a header, enter copies the cell to the clipboard, letters or a paste |
| 6 Nested Borders | recursive boxes through a content generator | b border |
| 7 Simple Borders | labeled cells, every border style, short cells | b border |
| 8 Fixed Row Heights | `SetFixedHeight` on rows | m fixed/dynamic |
| 9 Fixed Column Widths | `SetFixedWidth` on cells | m fixed/dynamic |
| 10 Mixed Fixed/Dynamic | fixed and ratio sizing in one grid | b border, m fixed/dynamic, h text |
| 11 Cell Config | changing a cell's fixed sizes at run time | b border, h text, C-W select a cell, then w, h, digits, enter, esc |
| 12 Row Alignment | `SetRowAlign` with rows narrower than the box | r row align |
| 13 Mouse and Cursor | `CellAt` for clicks and `GetContentBounds` for the terminal's cursor | click, arrows |

The table demos show `example/sample/sample.csv`, which is compiled into them.

![FlexBox Simple](.github/images/flex-box-simple.gif)

## Development

```bash
make prereqs   # check for Go, install golangci-lint when it is missing
make test      # the library's tests, then the demos' tests
make lint      # golangci-lint on both modules
make bench     # the render, sort and filter benchmarks
make cover     # coverage of the library, function by function
make check     # lint and test
```

The repository holds two Go modules. The root module is the library: packages
`flexbox` and `table`, which depend on Lip Gloss and need Go 1.25. The
`example` module holds the demos, adds Bubble Tea and gocsv, and needs Go 1.26.

The tests cover:

- **`flexbox`.** The sizing rules, the block operations behind rendering, the
  style features of Lip Gloss v2 (gradient borders, underline styles,
  hyperlinks), and every rule under
  [What renders when space runs out](#what-renders-when-space-runs-out). For
  random grids, the text a box renders inside a cell's bounds must be the
  cell's own block, which ties positions to the screen.
  A reference renderer composes random grids with plain Lip Gloss joins and
  styles; the boxes must return the same text after every random change, which
  covers both the direct assembly of unstyled elements and the reuse of what
  did not change.
- **`table`.** Typed sorting, filtering, the cursor and the viewport, the
  header and footer at every size, value formatting, bad input and style
  overrides. A table that renders after every random change must show what a
  table built from scratch with the same changes shows.
- **`example`.** A test cycles every demo at several terminal sizes and checks
  that each fits under the header. Others click cells and table rows, turn the
  wheel and paste, through the combined program's header and into a nested
  table.

Continuous integration runs the tests and the linter on every push and pull
request.

## Security

Report a vulnerability privately, as [SECURITY.md](SECURITY.md) describes.

## License

MIT, see [LICENSE](LICENSE).

***
Made with [Charm](https://charm.sh).

<a href="https://charm.sh/"><img alt="The Charm logo" src="https://stuff.charm.sh/charm-badge-unrounded.jpg" width="400"></a>
