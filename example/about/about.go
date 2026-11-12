// Package about draws the box the demos show on top of themselves when a is pressed.
package about

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var style = lipgloss.NewStyle().
	Padding(2, 4).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Background(lipgloss.Color("#1a1a2e"))

// Over returns content with text in a box on top of it, at the center of the width by
// height area the content fills. The box is a layer of its own, so the content stays
// visible around it; a box larger than the area is cut at the area's edges. The result
// is height lines of width columns. An area without columns or lines has no room for a
// box, and the content comes back as it is.
func Over(content, text string, width, height int) string {
	if width <= 0 || height <= 0 {
		return content
	}
	box := style.Render(text)
	boxWidth, boxHeight := lipgloss.Size(box)
	layers := lipgloss.NewCompositor(
		lipgloss.NewLayer(content),
		lipgloss.NewLayer(box).X((width-boxWidth)/2).Y((height-boxHeight)/2).Z(1),
	)
	// the canvas ends every line at its last character; a full-width line keeps what
	// is joined to the right of the result in its column
	lines := strings.Split(lipgloss.NewCanvas(width, height).Compose(layers).Render(), "\n")
	for i, line := range lines {
		lines[i] = line + strings.Repeat(" ", max(width-lipgloss.Width(line), 0))
	}
	return strings.Join(lines, "\n")
}
