package flexbox

// track is one entry that shares a length with its siblings: a cell along its row or
// column, or a row or column along its box.
type track struct {
	ratio   int // share of the length the fixed tracks leave
	minimum int // least size of a ratio-sized track
	fixed   int // exact size when greater than 0

	size int // the result of distribute
	// cross is the size across the shared length, which the owner of the tracks fills
	// in: the height of a cell in a row, the width of a cell in a column
	cross int

	pinned bool // took its minimum
	bumped bool // took a leftover unit
}

// distribute sizes the tracks to share total. Fixed tracks take their size. The others
// split what is left in proportion to their ratios; a track whose share falls under its
// minimum takes the minimum and the others split the rest. When the fixed tracks leave
// nothing, the others get their minimum, which is 0 without one.
func distribute(tracks []track, total int) {
	remaining := total
	for i := range tracks {
		t := &tracks[i]
		t.size, t.pinned = 0, false
		if t.fixed > 0 {
			t.size = t.fixed
			remaining -= t.fixed
		}
	}

	for {
		share(tracks, remaining)
		short := -1
		for i := range tracks {
			if t := &tracks[i]; t.flexible() && t.minimum > t.size {
				short = i
				break
			}
		}
		if short < 0 {
			return
		}
		t := &tracks[short]
		t.size, t.pinned = t.minimum, true
		remaining -= t.minimum
	}
}

// flexible reports whether the track is sized by its ratio.
func (t *track) flexible() bool {
	return t.fixed <= 0 && !t.pinned
}

// share splits total between the flexible tracks in proportion to their ratios. Each
// share is rounded down and the leftover goes one unit at a time to the largest ratios,
// the first of equal ratios winning.
func share(tracks []track, total int) {
	combined := 0
	for i := range tracks {
		if t := &tracks[i]; t.flexible() {
			t.size, t.bumped = 0, false
			combined += max(t.ratio, 0)
		}
	}
	if combined == 0 || total <= 0 {
		return
	}

	leftover := total
	for i := range tracks {
		if t := &tracks[i]; t.flexible() && t.ratio > 0 {
			t.size = t.ratio * total / combined
			leftover -= t.size
		}
	}
	for ; leftover > 0; leftover-- {
		largest := -1
		for i := range tracks {
			t := &tracks[i]
			if !t.flexible() || t.bumped || t.ratio <= 0 {
				continue
			}
			if largest < 0 || t.ratio > tracks[largest].ratio {
				largest = i
			}
		}
		if largest < 0 {
			return
		}
		tracks[largest].size++
		tracks[largest].bumped = true
	}
}

// scale returns ratio/largest of total, rounded down. It sizes a cell across its row or
// column, where the cell with the largest ratio takes the whole length.
func scale(ratio, largest, total int) int {
	if ratio <= 0 || largest <= 0 || total <= 0 {
		return 0
	}
	return ratio * total / largest
}
