package flexbox

import (
	"slices"
	"testing"
)

func TestDistribute(t *testing.T) {
	ratio := func(r int) track { return track{ratio: r} }
	for _, tc := range []struct {
		name   string
		tracks []track
		total  int
		want   []int
	}{
		{"no tracks", nil, 10, nil},
		{"equal ratios, leftover to the first", []track{ratio(1), ratio(1), ratio(1)}, 10, []int{4, 3, 3}},
		{"leftover to the largest ratio", []track{ratio(1), ratio(2), ratio(3)}, 7, []int{1, 2, 4}},
		{"leftover to each of the largest ratios once", []track{ratio(3), ratio(1), ratio(3), ratio(1)}, 11, []int{5, 1, 4, 1}},
		{"shares are exact integers", slices.Repeat([]track{ratio(1)}, 49), 49, slices.Repeat([]int{1}, 49)},
		{"fixed tracks take their size first", []track{{fixed: 4, ratio: 9}, ratio(1), ratio(1)}, 10, []int{4, 3, 3}},
		{"fixed tracks over the total leave nothing", []track{{fixed: 12}, ratio(1)}, 10, []int{12, 0}},
		{"fixed tracks that use the total leave the minimums", []track{{fixed: 10}, {ratio: 1, minimum: 3}, ratio(1)}, 10, []int{10, 3, 0}},
		{"fixed tracks over the total leave the minimums", []track{{fixed: 12}, {ratio: 1, minimum: 3}}, 10, []int{12, 3}},
		{"a minimum over the share is taken from the others", []track{{ratio: 1, minimum: 8}, ratio(1)}, 10, []int{8, 2}},
		{"a minimum under the share changes nothing", []track{{ratio: 1, minimum: 2}, ratio(1)}, 10, []int{5, 5}},
		{"minimums over the total all hold", []track{{ratio: 1, minimum: 8}, {ratio: 1, minimum: 8}}, 10, []int{8, 8}},
		{"a minimum that leaves nothing", []track{{ratio: 1, minimum: 12}, ratio(1)}, 10, []int{12, 0}},
		{"a second minimum shows after the first is taken", []track{{ratio: 4, minimum: 9}, {ratio: 1, minimum: 1}, ratio(1)}, 10, []int{9, 1, 0}},
		{"zero ratios get nothing", []track{ratio(0), ratio(0)}, 5, []int{0, 0}},
		{"a zero ratio next to a ratio gets nothing", []track{ratio(0), ratio(2)}, 5, []int{0, 5}},
		{"negative ratios count as zero", []track{ratio(-3), ratio(1)}, 5, []int{0, 5}},
		{"no total", []track{ratio(1), {fixed: 3}}, 0, []int{0, 3}},
		{"negative total", []track{ratio(1), ratio(2)}, -4, []int{0, 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			distribute(tc.tracks, tc.total)
			var got []int
			for _, track := range tc.tracks {
				got = append(got, track.size)
			}
			if !slices.Equal(got, tc.want) {
				t.Fatalf("sizes %v, want %v", got, tc.want)
			}
		})
	}
}

// TestDistributeIsRepeatable: the result must not depend on what an earlier call left in
// the tracks.
func TestDistributeIsRepeatable(t *testing.T) {
	tracks := []track{{ratio: 1, minimum: 8}, {ratio: 1}, {ratio: 1}}
	distribute(tracks, 10)
	distribute(tracks, 30)
	if got := []int{tracks[0].size, tracks[1].size, tracks[2].size}; !slices.Equal(got, []int{10, 10, 10}) {
		t.Fatalf("sizes %v, want [10 10 10]", got)
	}
}

func TestScale(t *testing.T) {
	for _, tc := range []struct{ ratio, largest, total, want int }{
		{4, 4, 10, 10},
		{3, 4, 10, 7},
		{1, 49, 49, 1},
		{0, 4, 10, 0},
		{2, 0, 10, 0},
		{2, 4, 0, 0},
		{2, 4, -3, 0},
	} {
		if got := scale(tc.ratio, tc.largest, tc.total); got != tc.want {
			t.Errorf("scale(%d, %d, %d) = %d, want %d", tc.ratio, tc.largest, tc.total, got, tc.want)
		}
	}
}
