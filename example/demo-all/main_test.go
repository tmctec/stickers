package main

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// The keys the tests press.
var (
	keySpace    = tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	keyTab      = tea.KeyPressMsg{Code: tea.KeyTab}
	keyShiftTab = tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	keyQ        = tea.KeyPressMsg{Code: 'q', Text: "q"}
	keyCtrlC    = tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
)

func press(t *testing.T, m *model, msg tea.Msg) *model {
	t.Helper()
	next, _ := m.Update(msg)
	m, ok := next.(*model)
	if !ok {
		t.Fatalf("Update returned %T, want *model", next)
	}
	return m
}

// TestSplashThenCycle starts on the splash screen, presses a key, then
// cycles every demo in both directions. Each demo must size and render
// without panicking.
func TestSplashThenCycle(t *testing.T) {
	m := newModel()
	m = press(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	if v := m.View().Content; !strings.Contains(v, "press any key to start") {
		t.Fatalf("splash missing start hint:\n%s", v)
	}

	m = press(t, m, keySpace)
	if !m.started {
		t.Fatal("space did not leave the splash screen")
	}
	if v := m.View().Content; !strings.Contains(v, demoInfos[0].name) {
		t.Fatalf("first demo header missing:\n%s", v)
	}

	for i := 0; i < len(m.demos); i++ {
		m = press(t, m, keyShiftTab)
		m.View()
	}
	if m.current != 0 {
		t.Fatalf("shift+tab x%d ended on %d, want 0", len(m.demos), m.current)
	}
	for i := 0; i < len(m.demos); i++ {
		m = press(t, m, keyTab)
		m.View()
	}
	if m.current != 0 {
		t.Fatalf("tab x%d ended on %d, want 0", len(m.demos), m.current)
	}
}

// TestSplashQuit checks q and ctrl+c quit from the splash screen.
func TestSplashQuit(t *testing.T) {
	for _, k := range []tea.KeyPressMsg{keyQ, keyCtrlC} {
		m := newModel()
		_, cmd := m.Update(k)
		if cmd == nil {
			t.Fatalf("%s on splash returned no command, want tea.Quit", k.String())
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("%s on splash did not quit", k.String())
		}
	}
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

// TestEveryDemoFitsUnderHeader: the header plus the demo never exceed the terminal, the
// header is one line that starts with the demo's name and own keys, carries the shared
// keys whenever it fits, and a resize while running keeps all of that true.
func TestEveryDemoFitsUnderHeader(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}, {160, 40}} {
		m := newModel()
		m = press(t, m, tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		m = press(t, m, keySpace)
		for i := range m.demos {
			checkFits(t, m, size[0], size[1])
			if i < len(m.demos)-1 {
				m = press(t, m, keyTab)
			}
		}
		m = press(t, m, tea.WindowSizeMsg{Width: size[0] - 10, Height: size[1] - 4})
		checkFits(t, m, size[0]-10, size[1]-4)
	}
}

func checkFits(t *testing.T, m *model, width, height int) {
	t.Helper()
	info := demoInfos[m.current]
	lines := strings.Split(m.View().Content, "\n")
	if len(lines) > height {
		t.Errorf("%s at %dx%d renders %d lines", info.name, width, height, len(lines))
	}
	header := strings.TrimSpace(ansiRe.ReplaceAllString(lines[0], ""))
	if w := len([]rune(header)); w > width {
		t.Errorf("%s header is %d columns wide at %d", info.name, w, width)
	}
	lead := "Demo " + info.name
	if info.keys != "" {
		lead += " | " + info.keys
	}
	full := headerText(info)
	switch {
	case len([]rune(full)) <= width && header != full:
		t.Errorf("%s header at %d columns is %q, want %q", info.name, width, header, full)
	case len([]rune(full)) > width && !strings.HasPrefix(header, lead[:min(len(lead), width-1)]):
		t.Errorf("%s header at %d columns lost its own keys: %q", info.name, width, header)
	}
}

// screen returns the lines of the model's view without escape sequences.
func screen(m *model) []string {
	return strings.Split(ansiRe.ReplaceAllString(m.View().Content, ""), "\n")
}

// TestMouseAndCursorAccountForTheHeader: a demo draws under the header line, so a click
// reaches it one line up and its cursor shows one line down.
func TestMouseAndCursorAccountForTheHeader(t *testing.T) {
	m := newModel()
	m = press(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(t, m, keySpace)
	m = press(t, m, keyShiftTab) // the last demo: Mouse and Cursor

	view := m.View()
	if view.MouseMode != tea.MouseModeCellMotion || view.WindowTitle != "Stickers: Mouse and Cursor" {
		t.Fatalf("the demo's mouse mode and title do not reach the program: %v, %q", view.MouseMode, view.WindowTitle)
	}
	cursorShowsLabel := func(label string) {
		t.Helper()
		view := m.View()
		if view.Cursor == nil {
			t.Fatal("the demo's cursor does not reach the program")
		}
		line := []rune(screen(m)[view.Cursor.Y])
		if got := string(line[view.Cursor.X : view.Cursor.X+len(label)]); got != label {
			t.Fatalf("the screen shows %q at the cursor, want %q", got, label)
		}
	}
	cursorShowsLabel("cell 1")

	// click the label of cell 6, wherever it is on screen
	for y, line := range screen(m) {
		if x := strings.Index(line, "cell 6"); x >= 0 {
			m = press(t, m, tea.MouseClickMsg{X: len([]rune(line[:x])), Y: y, Button: tea.MouseLeft})
		}
	}
	cursorShowsLabel("cell 6")
}

// TestClickSelectsATableRowUnderTheHeader: in a table demo the table's header is on
// line 1 of the screen and its first row on line 2.
func TestClickSelectsATableRowUnderTheHeader(t *testing.T) {
	m := newModel()
	m = press(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	m = press(t, m, keySpace)
	for i := 0; i < 3; i++ {
		m = press(t, m, keyTab) // demo 4: Table Simple String
	}
	if !strings.Contains(strings.Join(screen(m), "\n"), "0:0 / 100:") {
		t.Fatalf("the table's footer does not show the cursor at 0:0:\n%s", strings.Join(screen(m), "\n"))
	}
	m = press(t, m, tea.MouseClickMsg{X: 30, Y: 4, Button: tea.MouseLeft})
	if !strings.Contains(strings.Join(screen(m), "\n"), "1:2 / 100:") {
		t.Fatalf("a click on line 4, the table's third row, did not move the cursor to 1:2:\n%s", strings.Join(screen(m), "\n"))
	}
	m = press(t, m, tea.MouseWheelMsg{X: 30, Y: 4, Button: tea.MouseWheelDown})
	if !strings.Contains(strings.Join(screen(m), "\n"), "1:3 / 100:") {
		t.Fatal("the wheel did not move the cursor a row down")
	}
	// a click on the table's header line sorts by that column
	m = press(t, m, tea.MouseClickMsg{X: 30, Y: 1, Button: tea.MouseLeft})
	if !strings.Contains(screen(m)[1], "▲") {
		t.Fatalf("a click on the table's header did not sort: %q", screen(m)[1])
	}
}

// TestAboutBeforeTheFirstWindowSize: the about box of a demo shows, or does not, without
// a panic while the program has no window size yet.
func TestAboutBeforeTheFirstWindowSize(t *testing.T) {
	m := newModel()
	m = press(t, m, keySpace)
	for range m.demos {
		m = press(t, m, tea.KeyPressMsg{Code: 'a', Text: "a"})
		m.View()
		m = press(t, m, tea.KeyPressMsg{Code: 'a', Text: "a"})
		m = press(t, m, keyTab)
	}
	m = press(t, m, tea.WindowSizeMsg{Width: 1, Height: 1})
	m = press(t, m, tea.KeyPressMsg{Code: 'a', Text: "a"})
	m.View()
}
