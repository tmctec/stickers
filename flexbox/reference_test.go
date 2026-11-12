package flexbox

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// The reference renderers below compose a grid from scratch with lipgloss joins and
// styles, with no blocks and no caches. The tests render random grids, change them at
// random and check that the boxes return the same text every time. That covers both the
// direct composition of unstyled rows, columns and boxes and the reuse of what did not
// change.

// referenceFlexBox renders box by composing its cells with lipgloss.
func referenceFlexBox(box *FlexBox) string {
	box.ForceRecalculate()
	var rows []string
	for _, row := range box.rows {
		if row.height <= 0 {
			continue
		}
		style := row.style.own
		if box.style.passing {
			style = style.Inherit(box.style.own)
		}
		var cells []string
		for _, cell := range row.cells {
			cells = append(cells, referenceCell(cell, style, row.style.passing))
		}
		width := row.width - style.GetHorizontalFrameSize()
		joined := referenceClip(lipgloss.JoinHorizontal(lipgloss.Top, cells...), width)
		rows = append(rows, style.Height(row.height-style.GetVerticalMargins()).MaxHeight(row.height).Render(joined))
	}
	return referenceFrame(box.style.own, box.width, box.height, lipgloss.JoinVertical(box.rowAlign, rows...))
}

// referenceHorizontal renders box by composing its cells with lipgloss.
func referenceHorizontal(box *HorizontalFlexBox) string {
	box.ForceRecalculate()
	var columns []string
	for _, column := range box.columns {
		style := column.style.own
		if box.style.passing {
			style = style.Inherit(box.style.own)
		}
		var cells []string
		for _, cell := range column.cells {
			if cell.height > 0 {
				cells = append(cells, referenceCell(cell, style, column.style.passing))
			}
		}
		columns = append(columns, referenceFrame(style, column.width, column.height, lipgloss.JoinVertical(lipgloss.Left, cells...)))
	}
	width := box.width - box.style.own.GetHorizontalFrameSize()
	placed := lipgloss.PlaceHorizontal(width, box.columnAlign, lipgloss.JoinHorizontal(lipgloss.Top, columns...))
	return referenceFrame(box.style.own, box.width, box.height, placed)
}

// referenceFrame renders content inside style at width by height. A lipgloss width and
// height count the border and the padding in and leave the margins out.
func referenceFrame(style lipgloss.Style, width, height int, content string) string {
	return renderStyled(style.
		Width(width-style.GetHorizontalMargins()).MaxWidth(width).
		Height(height-style.GetVerticalMargins()).MaxHeight(height), content)
}

// referenceCell renders cell with lipgloss alone.
func referenceCell(cell *Cell, parent lipgloss.Style, passing bool) string {
	style := cell.style.own
	if passing {
		style = style.Inherit(parent)
	}
	if cell.height <= 0 {
		return strings.Repeat(" ", max(cell.width, 0))
	}
	if cell.width <= 0 {
		return strings.Repeat("\n", cell.height-1)
	}
	top := style.GetMarginTop() + style.GetBorderTopSize()
	bottom := style.GetBorderBottomSize() + style.GetMarginBottom()
	width := cell.width - style.GetHorizontalBorderSize() - style.GetHorizontalMargins()
	height := cell.height - top - bottom
	width, height = width-style.GetHorizontalPadding(), height-style.GetVerticalPadding()
	var content string
	if cell.contentGenerator != nil {
		content = cell.contentGenerator(max(width, 0), max(height, 0))
	}
	fitted := fitContent(content, width, height)
	rendered := renderStyled(style.
		Width(cell.width-style.GetHorizontalMargins()).MaxWidth(cell.width).
		Height(cell.height-style.GetVerticalMargins()), fitted)
	lines := trimInside(strings.Split(rendered, "\n"), cell.height, top, bottom)
	// lipgloss leaves a line short when a wide character does not fit before the cut;
	// a cell is always its full width
	for i, line := range lines {
		lines[i] = line + strings.Repeat(" ", max(cell.width-ansi.StringWidth(line), 0))
	}
	return strings.Join(lines, "\n")
}

// referenceClip cuts every line of s at width columns.
func referenceClip(s string, width int) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if ansi.StringWidth(line) > width {
			lines[i] = ansi.Truncate(line, max(width, 0), "")
		}
	}
	return strings.Join(lines, "\n")
}

var (
	referenceContents = []string{
		"", "x", "lorem ipsum dolor sit amet", "one\ntwo\nthree\nfour\nfive", "trailing   \nspaces   ",
		"tab\tseparated", "wide 日本語 text", "a-very-long-word-that-has-to-be-broken-somewhere", "  indented",
	}
	referenceCellStyles = []lipgloss.Style{
		lipgloss.NewStyle(),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder()),
		lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#ff0000")),
		lipgloss.NewStyle().Background(lipgloss.Color("#123456")).Foreground(lipgloss.Color("#ffffff")),
		lipgloss.NewStyle().Padding(1, 2),
		lipgloss.NewStyle().Margin(0, 1).Bold(true),
		lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center),
		lipgloss.NewStyle().Align(lipgloss.Right).Underline(true),
	}
	// container styles have no vertical padding, so a short row or column always has
	// room for its frame
	referenceContainerStyles = []lipgloss.Style{
		lipgloss.NewStyle(),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder()),
		lipgloss.NewStyle().Background(lipgloss.Color("#222222")),
		lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#00ff00")),
		lipgloss.NewStyle().Align(lipgloss.Center).Italic(true),
	}
	referenceAligns = []lipgloss.Position{lipgloss.Left, lipgloss.Center, lipgloss.Right}
)

func pick[T any](rng *rand.Rand, values []T) T {
	return values[rng.Intn(len(values))]
}

// randomCell makes a cell with a random ratio, content and, three times out of four, a
// style.
func randomCell(rng *rand.Rand) *Cell {
	cell := NewCell(1+rng.Intn(4), 1+rng.Intn(3))
	if rng.Intn(4) > 0 {
		cell.SetStyle(pick(rng, referenceCellStyles))
	}
	if rng.Intn(5) > 0 {
		cell.SetContent(pick(rng, referenceContents))
	}
	if rng.Intn(6) == 0 {
		cell.SetContentGenerator(func(w, h int) string { return strconv.Itoa(w) + "x" + strconv.Itoa(h) })
	}
	return cell
}

// mutateCell changes one random property of cell.
func mutateCell(rng *rand.Rand, cell *Cell, alongFixed func(int) *Cell, alongMin func(int) *Cell) {
	switch rng.Intn(5) {
	case 0:
		cell.SetContent(pick(rng, referenceContents))
	case 1:
		cell.SetStyle(pick(rng, referenceCellStyles))
	case 2:
		alongFixed(rng.Intn(3) * (2 + rng.Intn(10)))
	case 3:
		alongMin(rng.Intn(3) * (1 + rng.Intn(6)))
	case 4:
		n := rng.Intn(1000)
		cell.SetContentGenerator(func(_, _ int) string { return "tick " + strconv.Itoa(n) })
	}
}

func assertSameRender(t *testing.T, step string, got, want string) {
	t.Helper()
	if got == want {
		return
	}
	gotLines, wantLines := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < len(gotLines) && i < len(wantLines); i++ {
		if gotLines[i] != wantLines[i] {
			t.Fatalf("%s: line %d of %d rendered\n%q\nreference has %d lines and\n%q", step, i, len(gotLines), gotLines[i], len(wantLines), wantLines[i])
		}
	}
	t.Fatalf("%s: rendered %d lines, reference %d", step, len(gotLines), len(wantLines))
}

func TestFlexBoxRendersLikeLipglossComposition(t *testing.T) {
	for seed := int64(0); seed < 80; seed++ {
		rng := rand.New(rand.NewSource(seed))
		box := New(40+rng.Intn(80), 30+rng.Intn(20)).SetRowAlign(pick(rng, referenceAligns))
		if rng.Intn(3) == 0 {
			box.SetStyle(pick(rng, referenceContainerStyles)).StylePassing(rng.Intn(2) == 0)
		}
		for r, n := 0, 1+rng.Intn(4); r < n; r++ {
			row := box.NewRow()
			for c, m := 0, 1+rng.Intn(5); c < m; c++ {
				row.AddCells(randomCell(rng))
			}
			if rng.Intn(3) == 0 {
				row.SetStyle(pick(rng, referenceContainerStyles)).StylePassing(rng.Intn(2) == 0)
			}
			box.AddRows([]*Row{row})
		}
		step := "seed " + strconv.FormatInt(seed, 10)
		assertSameRender(t, step, box.Render(), referenceFlexBox(box))

		for i := 0; i < 12; i++ {
			row := box.rows[rng.Intn(len(box.rows))]
			cell := row.cells[rng.Intn(len(row.cells))]
			switch rng.Intn(8) {
			case 0:
				box.SetWidth(40 + rng.Intn(80)).SetHeight(30 + rng.Intn(20))
			case 1:
				row.SetStyle(pick(rng, referenceContainerStyles)).StylePassing(rng.Intn(2) == 0)
			case 2:
				box.SetStyle(pick(rng, referenceContainerStyles)).StylePassing(rng.Intn(2) == 0)
			case 3:
				row.AddCells(randomCell(rng))
			case 4:
				box.SetRowAlign(pick(rng, referenceAligns))
			default:
				mutateCell(rng, cell, cell.SetFixedWidth, cell.SetMinWidth)
			}
			assertSameRender(t, step+" change "+strconv.Itoa(i), box.Render(), referenceFlexBox(box))
			assertSameRender(t, step+" repeat "+strconv.Itoa(i), box.Render(), referenceFlexBox(box))
		}
	}
}

func TestHorizontalFlexBoxRendersLikeLipglossComposition(t *testing.T) {
	for seed := int64(0); seed < 80; seed++ {
		rng := rand.New(rand.NewSource(seed))
		box := NewHorizontal(60+rng.Intn(60), 30+rng.Intn(20)).SetColumnAlign(pick(rng, referenceAligns))
		if rng.Intn(3) == 0 {
			box.SetStyle(pick(rng, referenceContainerStyles)).StylePassing(rng.Intn(2) == 0)
		}
		for c, n := 0, 1+rng.Intn(4); c < n; c++ {
			column := box.NewColumn()
			for r, m := 0, 1+rng.Intn(4); r < m; r++ {
				column.AddCells(randomCell(rng))
			}
			if rng.Intn(3) == 0 {
				column.SetStyle(pick(rng, referenceContainerStyles)).StylePassing(rng.Intn(2) == 0)
			}
			box.AddColumns([]*Column{column})
		}
		step := "seed " + strconv.FormatInt(seed, 10)
		assertSameRender(t, step, box.Render(), referenceHorizontal(box))

		for i := 0; i < 12; i++ {
			column := box.columns[rng.Intn(len(box.columns))]
			cell := column.cells[rng.Intn(len(column.cells))]
			switch rng.Intn(8) {
			case 0:
				box.SetWidth(60 + rng.Intn(60)).SetHeight(30 + rng.Intn(20))
			case 1:
				column.SetStyle(pick(rng, referenceContainerStyles)).StylePassing(rng.Intn(2) == 0)
			case 2:
				box.SetStyle(pick(rng, referenceContainerStyles)).StylePassing(rng.Intn(2) == 0)
			case 3:
				box.SetColumnAlign(pick(rng, referenceAligns))
			case 4:
				box.LockColumnWidth(rng.Intn(2) * (4 + rng.Intn(8)))
			default:
				// small sizes, so the cells of a column never outgrow it
				mutateCell(rng, cell,
					func(v int) *Cell { return cell.SetFixedHeight(min(v, 4)) },
					func(v int) *Cell { return cell.SetMinHeight(min(v, 3)) })
			}
			assertSameRender(t, step+" change "+strconv.Itoa(i), box.Render(), referenceHorizontal(box))
			assertSameRender(t, step+" repeat "+strconv.Itoa(i), box.Render(), referenceHorizontal(box))
		}
	}
}
