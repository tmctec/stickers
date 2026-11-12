# Changelog

All notable changes to this project will be documented in this file.
## [2.0.0] - 2026-11-11
### ⚠ BREAKING CHANGES
- The module is `github.com/tmctec/stickers/v2` and takes Lip Gloss v2 styles from `charm.land/lipgloss/v2`. For Lip Gloss v1, use the upstream module, `github.com/76creates/stickers`. The library needs Go 1.25.
- Styles render in full color whatever the terminal, as Lip Gloss v2 styles do. Bubble Tea v2 fits the colors to the terminal, and `lipgloss.Println` does outside it.
- A tab in the content of a cell shows as four spaces, whatever tab width the style sets.
- A content generator receives the room its content has, inside the cell's padding too. It used to receive the padding's columns and lines as well, so a generator that filled its area in a padded cell was wrapped.
- The demos are a Go module of their own, `github.com/tmctec/stickers/example`, and run on Bubble Tea v2. The library module requires Lip Gloss and nothing from Bubble Tea or gocsv.
- `Cell.GetStyle` returns the style set on the cell. A render no longer merges the keys a row or column passes down into it.
- A row without lines takes no line of its box. It used to take one blank line, which pushed the rows under it out of a short box.
- Table cells cut a value that is too wide and end it in `…`. They used to wrap it at a word and show the first line.
- A table keeps its sort: rows added after `OrderByAsc` or `OrderByDesc` are sorted in, and `ClearRows` keeps the sort, the filter and the cursor's row.
- The cursor of a table sits on the last row only while a filter hides its own row, and is back on its row when the filter goes. It used to stay on the row the filter moved it to.
- `Table.SetTypes` returns an `ErrorRowLen` for a wrong number of types. The messages of the table's errors start with `table:`.
- A minimum width or height holds next to fixed sizes that leave nothing. It used to drop to 0 there.
- `SetRows` and `SetColumns` copy the slice they are given.
### Fixes
#### flexbox
- Every setter takes effect on the next render. `SetWidth` or `SetHeight` alone on a box with a border, `LockRowHeight`, `LockColumnWidth`, `SetStyle` on a box, row or column, and setters on a cell kept from before the first render used to wait for another call to trigger a new layout.
- A style passed down with `StylePassing` or `UseStyle` reaches the children again after it changes, and stops when passing is turned off. Children used to keep the first style they inherited.
- A box or column with padding is its full width. The padding was subtracted twice, so the box rendered narrower and its content wrapped.
- A box with a row or column that has no cells, or whose cells all have a ratio of 0, renders. It used to panic.
- Columns that together are wider than a `HorizontalFlexBox` are cut at its edge. They used to wrap onto extra lines.
- Cell content sits where the style's vertical alignment puts it. Since the trailing-space fix below, centered content sat at the top and bottom-aligned content one line too high.
- A cell keeps the trailing spaces of its last content line, so a centered or right-aligned block of lines keeps every line at the same offset. The same holds for the last line of a styled column or box.
- A cell's border and margins always render complete when the cell has room for them. Content taller than the cell is clipped to the content area instead of pushing the bottom border out of the cell. A styled row, column or box keeps its bottom border the same way.
- A row clips cells that stick out past its edge, such as a fixed-width cell wider than the box, instead of word-wrapping them onto extra lines. A cell with zero width adds no columns to its row.
- A cell is always its full width. A wide character that does not fit before the cell's edge used to leave the line one column short and shift the cells to its right.
- Shares are computed with integers. Floating point rounding gave some cells one unit too few, such as each of 49 equal cells in 49 columns.
- A content generator never receives a negative size, and a cell never gets one.
- A cell wraps its content without losing any of it. A leading space before a long word, wide characters that end at the cell's edge, and combining marks used to lose characters or blank the cell. The cell wraps the content itself before Lip Gloss does, because the wrap inside Lip Gloss v2 breaks such lines too late.
- Escape sequences in the content of a cell without a style are closed at every line end. A color used to run into the next cell.
- Content with Windows line ends leaves no carriage return in a cell.
- A bordered row too short for its cells closes its frame over its full width instead of two columns.
- A nil row, column or cell, such as a getter returns for an index that is not there, is left out instead of causing a panic at the next render.
- A hyperlink set on a cell, row or column stays in that element. It used to run on into the cells beside it on every line but the last.
- A cell, row, column or box that underlines or strikes through keeps styled content whole and draws its line through it. Lip Gloss writes the two one rune at a time, which cut the escape sequences of colored or linked content apart and showed them as text.
- Content that ends inside a color or a hyperlink takes no extra line, so it sits where plain content sits when it is centered or at the bottom.
- A row whose frame alone is wider than its box is cut at the box's edge and leaves the other rows in place.
- A content generator that reads its cell's position gets the position of the render it runs in, from the first render on.
- `Render` returns exactly the box's height in lines and width in columns for every size and style.
#### table
- `SetFilter`, `OrderByAsc` and `OrderByDesc` ignore a negative index. They used to panic.
- `SetTypes` returns an `ErrorBadType` for a nil type. It used to panic.
- `GetCursorValue` and `GetCursorLocation` reflect `SetFilter`, `ClearRows` and `AddRows` at once. They used to answer from the rows of the last render, and `GetCursorValue` panicked after a reload with fewer rows.
- A table renders exactly its height: one line is the header, two the header and footer. It used to render three lines at any height under three, and a footer wider than the table wrapped onto a second line.
- Floats show all their digits, `9.5` for 9.5. They used to show one significant digit, `1E+01`.
- A sort applies whatever the filter shows. It used to be ignored while fewer than two rows were visible.
- `OrderByAsc` sorts smallest first and `OrderByDesc` largest first; the two were swapped.
- Header titles are cut by the columns they take on screen. Titles with multi-byte or wide characters used to be cut inside a character.
- The footer names the filter only while one is set, not for an empty filter string.
- `SetMinWidth` recomputes the visible columns, and the cursor's column is visible even when it alone is wider than the table.
- A table without columns renders. It used to panic.
- `SetWidth` with the width the table already has changes nothing. It used to choose the visible columns again from the cursor's column, which scrolled a table that is sized on every render.
- The table is its size whatever frame the footer's style has; a border or padding on it used to add lines.
- A table without width renders the number of lines of its current height.
- A value is cut to the room its cell has: the padding of the cursor cell's style, tabs and Windows line ends are counted, so a cut value always ends in `…`.
- `int64` columns are accepted by `SetTypes` and `AddRows`, as `Ordered` always listed them.
- `SetStyles` overrides apply to that table only; the package defaults and other tables keep their styles.
- A table rendered before `SetHeight` is the height given to `NewTable`, not one line taller.
- After horizontal scrolling, rows size their cells with the ratios and minimum widths of the columns on screen, as the header does.
### Performance
- `Render` reuses what did not change. A cell renders again only when its size, style or content changed, a row or column only when one of its cells did. On a 160×48 grid of 17 cells an unchanged render takes 0.5 µs instead of 1.8 ms and allocates nothing; a render after one cell's content changed takes 35 µs.
- Rows, columns and boxes without a style are assembled directly instead of through a Lip Gloss style and join, which makes a full render after a resize about five times faster.
- A table keeps the rows on screen between renders. A cursor move takes about 45 µs instead of 1 ms, also while scrolling and under a filter, and an unchanged table 10 µs.
- Sorting is a stable O(n log n) sort in place instead of a bubble sort: 0.8 ms instead of 102 ms for 10,000 rows.
- The filter runs when the filter or the rows change, not on every cursor move, and compares ASCII text without allocating.
### Features
- Positions and hit testing. `GetBounds` on a cell, row or column returns its rectangle in the box, `Cell.GetContentBounds` the rectangle inside the cell's border and margins, and `CellAt` on a box the cell that shows at a point. A mouse click finds its cell, and the terminal's cursor finds its place in one.
- Mouse support for the table. `CellAt` turns a position on the rendered table into a column and a row, `ColumnAt` names the header at a position and `RowAt` the row on a line, and `SetCursorLocation` moves the cursor.
- The cell setters apply in both boxes. In a `FlexBox`, `SetMinHeight` on a cell sets the minimum height of its row. In a `HorizontalFlexBox`, `SetFixedHeight` on a cell pins its height, and `SetFixedWidth` and `SetMinWidth` set the width of its column.
- `Column.SetFixedWidth` pins a column, as `Row.SetFixedHeight` pins a row.
- `Table.GetCursorRow` returns the values of the row under the cursor.
- `Cell.GetID`, and `GetWidth` and `GetHeight` on `Row` and `Column`.
- `UpdateCellWithIndex` returns its row or column, so it chains like the other setters. `UpdateRow` and `UpdateColumn` ignore an index out of range instead of panicking.
- `HorizontalFlexBox.SetColumnAlign` places columns that together are narrower than the box at left, center or right.
### Changes
- `SetRatio` and `SetMinWidth` panic on a wrong number of values or a ratio under 1 instead of calling `log.Fatalf`, and copy the slices they are given.
### Examples
- demo-13 selects cells with the mouse through `CellAt` and puts the terminal's cursor in the selected cell.
- The table demos take the mouse: a click selects a cell, a click on a header sorts, the wheel moves the cursor. A paste extends the filter, and enter in demo-4 and demo-5 copies the cell to the clipboard. The three demos share their input handling in `example/tableinput`.
- The about box floats over its demo as a Lip Gloss layer, from the shared package `example/about`, and every demo names itself in the terminal's title.
- `make run` runs every demo in one program. The Makefile follows the two-layer convention with `build`, `test`, `lint`, `bench`, `cover` and `check` targets.
- The table demos compile `example/sample/sample.csv` in, so every demo runs from any directory.
- demo-3 embeds its table through a content generator.
- The directory of demo-2 is `demo-2-flex-box-horizontal`.
- The size-logging copy of demo-10 and the `demo-10-log`, `log` and `log-follow` make targets are removed.
- `make demo-all` opens on a splash screen, shows one header format for every demo (name, the demo's own keys, then `a`, `tab`, `shift+tab`, `q`), and every demo fits under the header at any terminal size.
- demo-7 cells show only their label when one line fits and nothing when none does; the border cycle runs Fill, Hidden, Normal, Rounded, Thick, Double.
- demo-10's header row keeps its text in dynamic mode; demo-11 no longer crashes on its first resize; demo-12's about text explains row alignment and the cut on narrow terminals.
- All demo sources are tracked; the `.gitignore` patterns for demo binaries are anchored to the repository root.
### Documentation
- README is a developer guide to both packages: sizing rules, what renders when space runs out, styles, recipes for nesting boxes and components, the table API, performance, demos and development.
- Package documentation for `flexbox` and `table`, a doc comment on every exported identifier, and runnable examples.
### Tooling
- Tests for the sizing rules and every rendering rule, and two randomized checks: the boxes against a reference built from plain Lip Gloss joins, and a table that renders after every change against one built from scratch.
- Benchmarks for rendering, sorting and filtering.
- A golangci-lint configuration and a GitHub Actions workflow that tests and lints both modules.
### Dependencies
- Lip Gloss is `charm.land/lipgloss/v2` `v2.0.6`, replacing `github.com/charmbracelet/lipgloss` `v0.13.0`.
- Updated `github.com/charmbracelet/x/ansi` to `v0.11.8` and added `github.com/charmbracelet/x/cellbuf` `v0.0.15` for wrapping.
- The demos use `charm.land/bubbletea/v2` `v2.1.1`, replacing `github.com/charmbracelet/bubbletea` `v1.1.1`.
## [v1.5.0](https://github.com/76creates/stickers/compare/v1.4.1...v1.5.0) (2025-09-29)
### ⚠ BREAKING CHANGES
- `TableStyleKey` is now `StyleKey`, and names of the keys better reflect their purpose within their package
- `TableSorting*` is now `SortingOrder*`, and use `SortingOrderKey` as type.
- Removed `OrderByColumn`, deprecated in v1.4.1; use `OrderByAsc` and `OrderByDesc`.
- Updated table package errors to have `Error` prefix instead of `Table` as a prefix, and `Error` as a suffix.
- Removed `TableSingleType`, generic table should be used.
### Fixes
- Fix typo in flexbox/Cell.SetMinHeight. #18 @tvanriel
- Fixed `flex-box-with-table` example to use new sorting functions.
- Refactored table go files.
### Deprecations
- `Cell.SetMinHeigth` is now deprecated in favour of SetMinHeight. #18 @tvanriel
### Features
- Added `x-scrolling` to the table, and refactored some mechanics to accommodate for this change.
- Added cell `ContentGenerator` functions that can be used to generate cell content with dimension context.

## [v1.4.1](https://github.com/76creates/stickers/compare/v1.4.0...v1.4.1) (2024-10-21)
### Features
- Added `OrderByAsc` and `OrderByDesc` methods to `Table` #12 @drmille2
### Deprecations
- `OrderByColumn` is deprecated by `OrderByAsc` and `OrderByDesc` methods. #12 @drmille2
### Updates
- Updated Go to version `1.23`
### Dependencies
- Updated `github.com/charmbracelet/lipgloss` to `v0.13.0`
- Updated `github.com/charmbracelet/bubbletea` to `v1.1.1`
- Updated `github.com/gocarina/gocsv` to `78e41c74b4b1`

## [v1.4.0](https://github.com/76creates/stickers/compare/v1.3.0...v1.4.0) (2024-09-11)
### ⚠ BREAKING CHANGES
- Moved `flexbox` and `table` into separate packages, `github.com/76creates/stickers/flexbox` and `github.com/76creates/stickers/table` respectively. #10 @jon4hz
### Fixes
- Minor lexical fixes
- Fixed repo tags to match go semver format.
### Dependencies
- Updated `github.com/charmbracelet/lipgloss` to `v0.6.0'
### Features
- Added `SetStylePassing` to _Table_ that will pass down the style all the way, from box to cell. No granularity for now.
- Added `HorizontalFlexBox`. #10 @jon4hz
### Updates
- Refactored `FlexBox.GetRow`, `FlexBox.Row`, `FlexBox.MustGetRow`, `FlexBoxRow.Cell`, `FlexBoxRow.GetCellWithID`, `FlexBoxRow.MustGetCellWithIndex`.<br>They are replaced with `FlexBoxRow.GetCell`, `FlexBoxRow.GetCellCopy`, `FlexBox.GetRow`, `FlexBox.GetRowCopy`,`FlexBox.GetRowCellCopy`.<br>Get* now returns pointer and triggers _recalculation_, while one can use Copy* function to get pointer to copied structs which can be used to lookup values without triggering _recalculation_.
- `AddCells` now take cells as a variadic argument. #10 @jon4hz

## [v1.3.0](https://github.com/76creates/stickers/compare/v1.2.0...v1.3.0) (2022-12-28)
### Fixes
* Fixed cursor not moving to the last visible row when filtering
* Fixed margins and borders not being rendered correctly #8
* Additionally, fixed margin and border issues on box
* Allow style override on _Table_ @joejag
### Features
* Converted default cell inheritance of the row style to function `StylePassing` which can be set on the _Box_ and _Row_, if both box and row have style passing enabled, the row will inherit the box style before it passes style to the cells.

## [v1.2.0](https://github.com/76creates/stickers/compare/v1.1.0...v1.2.0) (2022-02-27)
### Features
* Filtering is now available for `Table` and `TableSingleType` using new methods:
    * `UnsetFilter` remove filtering
    * `SetFilter` sets the filter on a column by index
    * `GetFilter` gets index of filtered column and the value of the filter
* Added `MustGetCellWithIndex`
* Fixed visible table calculations when filtering
* Added filter info to the status box
* Header rendering of sorting and filtering symbols is improved

## [v1.1.0](https://github.com/76creates/stickers/compare/v1.0.0...v1.1.0) (2022-02-26)
### ⚠ BREAKING CHANGES
* Refactored `Table` to support sorting, some methods have changed most notably revolving around adding rows since now its taking [][]any instead of [][]string, initial `Table` is now closer to `TableSingleType[string]`
* Stickers now uses generics, so go1.18 is mandatory

### Fixes
* Fixed recalculation triggering when *FlexBoxCell or *FlexBoxRow is fetched from the FlexBox
* Small lexical changes and tidying up

### Features
* Sorting is now available for `Table` and `TableSingleType`
* `Table` has been reformatted and now supports **sorting by type**, when `Table` is initialized each colum type is set to `string`, you can now update that using `SetType` method, types supported are located in interface `Ordered`
* Added `TableSingleType` type which locks row type to `string`, this makes it easier for user when adding rows as there is much fewer errors that can occur as when using `Table` where all depends on a type
* Added method `OrderByColumn` which invokes sorting for column `n`, for now you cannot explicitly set sorting direction and it's switching between `asc` and `desc` when you sort same column 
* Added method `GetCursorLocation` which returns `x`,`y` of the current cursor location on the table
* Added error types `TableBadTypeError`, `TableRowLenError`, `TableBadCellTypeError`
* Minor performance enhancements
