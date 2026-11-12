package flexbox

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// TestRowClipsFixedCellWiderThanBox: a fixed-width cell wider than the box is cut at the
// box edge, never wrapped onto extra lines.
func TestRowClipsFixedCellWiderThanBox(t *testing.T) {
	box := New(20, 3)
	row := box.NewRow()
	row.AddCells(NewCell(1, 1).SetStyle(bordered()).SetContent("x").SetFixedWidth(30))
	box.AddRows([]*Row{row})

	assertLines(t, strings.Split(box.Render(), "\n"),
		"┌───────────────────",
		"│x                  ",
		"└───────────────────")
}

// TestRowZeroWidthCellAddsNothing: when fixed cells use up the row, the dynamic cells get
// no columns and must neither push the others nor show their content.
func TestRowZeroWidthCellAddsNothing(t *testing.T) {
	box := New(20, 3)
	row := box.NewRow()
	row.AddCells(
		NewCell(1, 1).SetContent("ZZZ"),
		NewCell(1, 1).SetStyle(bordered()).SetContent("x").SetFixedWidth(30),
	)
	box.AddRows([]*Row{row})

	assertLines(t, strings.Split(box.Render(), "\n"),
		"┌───────────────────",
		"│x                  ",
		"└───────────────────")
}

// TestRowNarrowerThanBoxIsAligned: a row narrower than the box keeps its own width so
// SetRowAlign can place it against the full-width rows.
func TestRowNarrowerThanBoxIsAligned(t *testing.T) {
	box := New(10, 2).SetRowAlign(lipgloss.Right)
	full := box.NewRow()
	full.AddCells(NewCell(1, 1).SetContent("##########"))
	narrow := box.NewRow()
	narrow.AddCells(NewCell(1, 1).SetContent("abc").SetFixedWidth(4))
	box.AddRows([]*Row{full, narrow})

	assertLines(t, strings.Split(box.Render(), "\n"),
		"##########",
		"      abc ")
}
