package demo7

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// bottomRow renders the demo at 70x23 with the given border index and returns the last
// row, which gets 4 lines at that height: P and T fill it, Q and S get 3, R gets 2.
func bottomRow(t *testing.T, borderType int) string {
	t.Helper()
	m := New()
	m.borderType = borderType
	m.updateStyles()
	m.Update(tea.WindowSizeMsg{Width: 70, Height: 23})
	lines := strings.Split(m.View().Content, "\n")
	if len(lines) != 23 {
		t.Fatalf("rendered %d lines, want 23", len(lines))
	}
	return strings.Join(lines[len(lines)-4:], "\n")
}

func TestBorderedShortCellsShowLabelOnly(t *testing.T) {
	row := bottomRow(t, 2) // normal border
	for _, want := range []string{"P", "[1:4]", "Q", "S", "T"} {
		if !strings.Contains(row, want) {
			t.Errorf("missing %q in\n%s", want, row)
		}
	}
	for _, unwanted := range []string{"[1:3]", "[1:2]", "R"} {
		if strings.Contains(row, unwanted) {
			t.Errorf("%q does not fit but was rendered in\n%s", unwanted, row)
		}
	}
	if got := strings.Count(row, "└"); got != 5 {
		t.Errorf("%d closed frames, want 5 in\n%s", got, row)
	}
}

func TestFilledCellsAllShowLabelAndRatio(t *testing.T) {
	row := bottomRow(t, 0)
	for _, want := range []string{"P", "[1:4]", "Q", "[1:3]", "R", "[1:2]", "S", "T"} {
		if !strings.Contains(row, want) {
			t.Errorf("missing %q in\n%s", want, row)
		}
	}
}

func TestBorderCycleOrder(t *testing.T) {
	want := []string{"Filled Background", "Hidden Border", "Normal Border", "Rounded Border", "Thick Border", "Double Border"}
	if strings.Join(borderNames, ",") != strings.Join(want, ",") {
		t.Fatalf("border order %v, want %v", borderNames, want)
	}
	if len(borderTypes) != len(borderNames) {
		t.Fatalf("%d border types for %d names", len(borderTypes), len(borderNames))
	}
	if m := New(); borderNames[m.borderType] != "Normal Border" {
		t.Fatalf("demo starts on %q, want Normal Border", borderNames[m.borderType])
	}
}
