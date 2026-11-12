package table_test

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/tmctec/stickers/v2/table"
)

// A table with typed columns, sorted by score, with the cursor on its second row. The
// example prints the table without its colors.
func Example() {
	t := table.NewTable(30, 5, []string{"ID", "Name", "Score"})
	if _, err := t.SetTypes(0, "", 0.0); err != nil {
		panic(err)
	}
	t.SetRatio([]int{1, 2, 2})
	t.MustAddRows([][]any{
		{2, "Linus", 8.0},
		{1, "Ada", 9.5},
		{3, "Grace", 9.25},
	})
	t.OrderByDesc(2).CursorDown()

	for _, line := range strings.Split(ansi.Strip(t.Render()), "\n") {
		fmt.Println(strings.TrimRight(line, " "))
	}
	fmt.Println("row under the cursor:", t.GetCursorRow())
	// Output:
	// ID    Name        Score ▼
	// 1     Ada         9.5
	// 3     Grace       9.25
	// 2     Linus       8
	//                     0:1 / 30:3
	// row under the cursor: [3 Grace 9.25]
}
