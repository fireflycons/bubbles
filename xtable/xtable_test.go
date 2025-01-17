package xtable

import (
	"hash/fnv"
	"os"
	"runtime"
	"strconv"
	"testing"
	"unsafe"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/stretchr/testify/require"
)

func TestFromValues(t *testing.T) {
	input := "foo1,bar1\nfoo2,bar2\nfoo3,bar3"
	table := New(WithColumns([]Column{{Title: "Foo"}, {Title: "Bar"}}))
	table.FromValues(input, ",")

	if len(table.rows) != 3 {
		t.Fatalf("expect table to have 3 rows but it has %d", len(table.rows))
	}

	expect := []Row{
		{
			Data: []string{"foo1", "bar1"},
		},
		{
			Data: []string{"foo2", "bar2"},
		},
		{
			Data: []string{"foo3", "bar3"},
		},
	}
	if !deepEqual(table.rows, expect) {
		t.Fatal("table rows is not equals to the input")
	}
}

func TestFromValuesWithTabSeparator(t *testing.T) {
	input := "foo1.\tbar1\nfoo,bar,baz\tbar,2"
	table := New(WithColumns([]Column{{Title: "Foo"}, {Title: "Bar"}}))
	table.FromValues(input, "\t")

	if len(table.rows) != 2 {
		t.Fatalf("expect table to have 2 rows but it has %d", len(table.rows))
	}

	expect := []Row{
		{
			Data: []string{"foo1.", "bar1"},
		},
		{
			Data: []string{"foo,bar,baz", "bar,2"},
		},
	}
	if !deepEqual(table.rows, expect) {
		t.Fatal("table rows is not equals to the input")
	}
}

func deepEqual(a, b []Row) bool {
	if len(a) != len(b) {
		return false
	}
	for i, r := range a {
		for j, f := range r.Data {
			if f != b[i].Data[j] {
				return false
			}
		}
	}
	return true
}

var cols = []Column{
	{Title: "col1", Width: 10},
	{Title: "col2", Width: 10},
	{Title: "col3", Width: 10},
}

func TestRenderRow(t *testing.T) {
	tests := []struct {
		name     string
		table    *Model
		expected string
	}{
		{
			name: "simple row",
			table: &Model{
				rows:   []Row{{Data: []string{"Foooooo", "Baaaaar", "Baaaaaz"}}},
				cols:   cols,
				styles: Styles{Cell: lipgloss.NewStyle()},
			},
			expected: "Foooooo   Baaaaar   Baaaaaz   ",
		},
		{
			name: "simple row with truncations",
			table: &Model{
				rows:   []Row{{Data: []string{"Foooooooooo", "Baaaaaaaaar", "Quuuuuuuuux"}}},
				cols:   cols,
				styles: Styles{Cell: lipgloss.NewStyle()},
			},
			expected: "Foooooooo…Baaaaaaaa…Quuuuuuuu…",
		},
		{
			name: "simple row avoiding truncations",
			table: &Model{
				rows:   []Row{{Data: []string{"Fooooooooo", "Baaaaaaaar", "Quuuuuuuux"}}},
				cols:   cols,
				styles: Styles{Cell: lipgloss.NewStyle()},
			},
			expected: "FoooooooooBaaaaaaaarQuuuuuuuux",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			row := tc.table.renderRow(0)
			if row != tc.expected {
				t.Fatalf("\n\nWant: \n%s\n\nGot:  \n%s\n", tc.expected, row)
			}
		})
	}
}

func TestTableAlignment(t *testing.T) {
	// Some issue with golden tests and EOL characters on Windows.
	// Oddly works locally on Windows, but not in GH actions
	skipIfGithubOnWindows(t)

	t.Run("No border", func(t *testing.T) {
		biscuits := New(
			WithHeight(5),
			WithColumns([]Column{
				{Title: "Name", Width: 25},
				{Title: "Country of Origin", Width: 16},
				{Title: "Dunk-able", Width: 12},
			}),
			WithRows([]Row{
				{
					Data: []string{"Chocolate Digestives", "UK", "Yes"},
				},
				{
					Data: []string{"Tim Tams", "Australia", "No"},
				},
				{
					Data: []string{"Hobnobs", "UK", "Yes"},
				},
			}),
		)
		got := ansi.Strip(biscuits.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("With border", func(t *testing.T) {
		baseStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

		s := DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)

		biscuits := New(
			WithHeight(5),
			WithColumns([]Column{
				{Title: "Name", Width: 25},
				{Title: "Country of Origin", Width: 16},
				{Title: "Dunk-able", Width: 12},
			}),
			WithRows([]Row{
				{
					Data: []string{"Chocolate Digestives", "UK", "Yes"},
				},
				{
					Data: []string{"Tim Tams", "Australia", "No"},
				},
				{
					Data: []string{"Hobnobs", "UK", "Yes"},
				},
			}),
			WithStyles(s),
		)
		got := ansi.Strip(baseStyle.Render(biscuits.View()))
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("No border row numbers", func(t *testing.T) {
		biscuits := New(
			WithRowNumbers(),
			WithHeight(5),
			WithColumns([]Column{
				{Title: "Name", Width: 25},
				{Title: "Country of Origin", Width: 16},
				{Title: "Dunk-able", Width: 12},
			}),
			WithRows([]Row{
				{
					Data: []string{"Chocolate Digestives", "UK", "Yes"},
				},
				{
					Data: []string{"Tim Tams", "Australia", "No"},
				},
				{
					Data: []string{"Hobnobs", "UK", "Yes"},
				},
			}),
		)
		got := ansi.Strip(biscuits.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("With border row numbers", func(t *testing.T) {
		baseStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

		s := DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)

		biscuits := New(
			WithRowNumbers(),
			WithHeight(5),
			WithColumns([]Column{
				{Title: "Name", Width: 25},
				{Title: "Country of Origin", Width: 16},
				{Title: "Dunk-able", Width: 12},
			}),
			WithRows([]Row{
				{
					Data: []string{"Chocolate Digestives", "UK", "Yes"},
				},
				{
					Data: []string{"Tim Tams", "Australia", "No"},
				},
				{
					Data: []string{"Hobnobs", "UK", "Yes"},
				},
			}),
			WithStyles(s),
		)
		got := ansi.Strip(baseStyle.Render(biscuits.View()))
		golden.RequireEqual(t, []byte(got))
	})
}

func TestPad(t *testing.T) {

	tests := []struct {
		name         string
		rows         []Row
		stringResult string
		intInput     int
		intResult    string
	}{
		{
			name:         "rows < 10",
			rows:         make([]Row, 8),
			stringResult: "#",
			intInput:     1,
			intResult:    "1",
		},
		{
			name:         "rows < 100 (a)",
			rows:         make([]Row, 34),
			stringResult: " #",
			intInput:     1,
			intResult:    " 1",
		},
		{
			name:         "rows < 100 (b)",
			rows:         make([]Row, 34),
			stringResult: " #",
			intInput:     99,
			intResult:    "99",
		},
		{
			name:         "rows < 1000",
			rows:         make([]Row, 100),
			stringResult: "  #",
			intInput:     1,
			intResult:    "  1",
		},
		{
			name:         "rows < 1000 (a)",
			rows:         make([]Row, 100),
			stringResult: "  #",
			intInput:     1,
			intResult:    "  1",
		},
		{
			name:         "rows < 1000 (b)",
			rows:         make([]Row, 100),
			stringResult: "  #",
			intInput:     42,
			intResult:    " 42",
		},
		{
			name:         "rows < 1000 (c)",
			rows:         make([]Row, 100),
			stringResult: "  #",
			intInput:     345,
			intResult:    "345",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var sRes, iRes string
			colWidth := rowNumberColWidth(test.rows)

			sRes = pad(colWidth, "#")

			require.Equal(t, test.stringResult, sRes)

			iRes = pad(colWidth, test.intInput)

			require.Equal(t, test.intResult, iRes)
		})
	}
}

func TestSortBy(t *testing.T) {

	thetable := New(
		WithHeight(5),
		WithColumns([]Column{
			{Title: "Strings", Width: 10},
			{Title: "Ints", Width: 10},
			{Title: "Floats", Width: 10},
		}),
		WithRows([]Row{
			{
				Data: []string{"abcdEfgh", "42", "0.72"},
			},
			{
				Data: []string{"qwerTYui", "123", "4.35"},
			},
			{
				Data: []string{"zxcvBNmj", "-4", "34.3"},
			},
			{
				Data: []string{"plmnPOiu", "4543534", "-23456.3"},
			},
		}),
	)

	tests := []struct {
		name      string
		col       int
		direction SortOrder
		hint      interface{}
		expected  []Row
	}{
		{
			name:      "col 0 asc",
			col:       0,
			direction: SortAscending,
			hint:      SortString,
			expected: []Row{
				{
					Data: []string{"abcdEfgh", "42", "0.72"},
				},
				{
					Data: []string{"plmnPOiu", "4543534", "-23456.3"},
				},
				{
					Data: []string{"qwerTYui", "123", "4.35"},
				},
				{
					Data: []string{"zxcvBNmj", "-4", "34.3"},
				},
			},
		},
		{
			name:      "col 0 desc",
			col:       0,
			direction: SortDescending,
			hint:      SortString,
			expected: []Row{
				{
					Data: []string{"zxcvBNmj", "-4", "34.3"},
				},
				{
					Data: []string{"qwerTYui", "123", "4.35"},
				},
				{
					Data: []string{"plmnPOiu", "4543534", "-23456.3"},
				},
				{
					Data: []string{"abcdEfgh", "42", "0.72"},
				},
			},
		},
		{
			name:      "col 1 asc",
			col:       1,
			direction: SortAscending,
			hint:      SortNumeric,
			expected: []Row{
				{
					Data: []string{"zxcvBNmj", "-4", "34.3"},
				},
				{
					Data: []string{"abcdEfgh", "42", "0.72"},
				},
				{
					Data: []string{"qwerTYui", "123", "4.35"},
				},
				{
					Data: []string{"plmnPOiu", "4543534", "-23456.3"},
				},
			},
		},
		{
			name:      "col 1 desc",
			col:       1,
			direction: SortDescending,
			hint:      SortNumeric,
			expected: []Row{
				{
					Data: []string{"plmnPOiu", "4543534", "-23456.3"},
				},
				{
					Data: []string{"qwerTYui", "123", "4.35"},
				},
				{
					Data: []string{"abcdEfgh", "42", "0.72"},
				},
				{
					Data: []string{"zxcvBNmj", "-4", "34.3"},
				},
			},
		},
		{
			name:      "col 2 asc",
			col:       2,
			direction: SortAscending,
			hint:      SortNumeric,
			expected: []Row{
				{
					Data: []string{"plmnPOiu", "4543534", "-23456.3"},
				},
				{
					Data: []string{"abcdEfgh", "42", "0.72"},
				},
				{
					Data: []string{"qwerTYui", "123", "4.35"},
				},
				{
					Data: []string{"zxcvBNmj", "-4", "34.3"},
				},
			},
		},
		{
			name:      "col 2 desc",
			col:       2,
			direction: SortDescending,
			hint:      SortNumeric,
			expected: []Row{
				{
					Data: []string{"zxcvBNmj", "-4", "34.3"},
				},
				{
					Data: []string{"qwerTYui", "123", "4.35"},
				},
				{
					Data: []string{"abcdEfgh", "42", "0.72"},
				},
				{
					Data: []string{"plmnPOiu", "4543534", "-23456.3"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			thetable.SortBy(test.col, test.direction, test.hint)
			require.Equal(t, test.expected, thetable.rows)
		})

	}
}

func TestFind(t *testing.T) {
	biscuits := New(
		WithHeight(5),
		WithColumns([]Column{
			{Title: "Name", Width: 25},
			{Title: "Country of Origin", Width: 16},
			{Title: "Dunk-able", Width: 12},
		}),
		WithRows([]Row{
			{
				Data: []string{"Chocolate Digestives", "UK", "Yes"},
			},
			{
				Data: []string{"Tim Tams", "Australia", "No"},
			},
			{
				Data: []string{"Hobnobs", "UK", "Yes"},
			},
			{
				Data: []string{"Peanut Butter Cookie", "USA", "Yes"},
			},
		}),
	)

	tests := []struct {
		name   string
		found  bool
		row    int
		term   string
		repeat bool
	}{
		{
			name:   "find first",
			found:  true,
			row:    2,
			term:   "Yes",
			repeat: false,
		},
		{
			name:   "find again",
			found:  true,
			row:    3,
			term:   "Yes",
			repeat: true,
		},
	}

	lastFind := 0
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			found := biscuits.Find(test.term, func() int {
				if test.repeat {
					return lastFind
				}
				return 0
			}())
			lastFind = biscuits.cursor
			require.True(t, found)
			require.Equal(t, biscuits.cursor, test.row)
		})
	}
}

type rowData struct {
	Name       string
	PacketSize int
	hash       uint64
}

func (r rowData) GetHashCode() uint64 {
	return r.hash
}

func intToBytes(i int) []byte {
	// Determine the size of the int type
	size := unsafe.Sizeof(i)

	// Create a slice of bytes referencing the memory of the integer
	return unsafe.Slice((*byte)(unsafe.Pointer(&i)), size)
}

func newRowData(s string, i int) rowData {
	// Precompute hash for speed
	hasher := fnv.New64a()
	hasher.Write([]byte(s))
	hasher.Write(intToBytes(i))
	hashVal := hasher.Sum64()

	return rowData{
		Name:       s,
		PacketSize: i,
		hash:       hashVal,
	}
}

func TestGetRowByHash(t *testing.T) {
	biscuits := New(
		WithHeight(5),
		WithColumns([]Column{
			{Title: "Name", Width: 25},
			{Title: "PacketSize", Width: 4},
		}),
		WithRows([]Row{
			{
				Data:     []string{"Chocolate Digestives", "12"},
				Metadata: newRowData("Chocolate Digestives", 12),
			},
			{
				Data:     []string{"Tim Tams", "8"},
				Metadata: newRowData("Tim Tams", 8),
			},
			{
				Data:     []string{"Hobnobs", "10"},
				Metadata: newRowData("Hobnobs", 10),
			},
			{
				Data:     []string{"Peanut Butter Cookie", "8"},
				Metadata: newRowData("Peanut Butter Cookie", 8),
			},
		}),
	)

	tests := []struct {
		row  int
		hash uint64
	}{
		{
			row:  0,
			hash: 0x40099b03ee546818,
		},
		{
			row:  1,
			hash: 0x73c93c25d05d3b20,
		},
		{
			row:  2,
			hash: 0x328fbeb31d4437e0,
		},
		{
			row:  3,
			hash: 0xca7bd7aa9e6497d4,
		},
		{
			row:  -1,
			hash: 0,
		},
	}

	for _, test := range tests {
		require.Equal(t, test.row, biscuits.GetRowByHash(test.hash))
	}
}

func TestWithStructData(t *testing.T) {

	data := []rowData{
		newRowData("Chocolate Digestives", 12),
		newRowData("Tim Tams", 8),
		newRowData("Hobnobs", 10),
		newRowData("Peanut Butter Cookie", 8),
	}

	table := New(WithStructData(data))

	require.Equal(t, 2, len(table.cols))
	require.Equal(t, 4, len(table.rows))
	require.Equal(t, "Name", table.cols[0].Title)
	require.Equal(t, "PacketSize", table.cols[1].Title)

	for i, r := range data {
		require.Equal(t, r.Name, table.rows[i].Data[0])
		require.Equal(t, strconv.Itoa(r.PacketSize), table.rows[i].Data[1])
	}
}

type taggedRowData struct {
	Name       string `xtable:"Biscuit"`
	PacketSize int
	hash       uint64
}

func (r taggedRowData) GetHashCode() uint64 {
	return r.hash
}

func newTaggedRowData(s string, i int) taggedRowData {
	// Precompute hash for speed
	hasher := fnv.New64a()
	hasher.Write([]byte(s))
	hasher.Write(intToBytes(i))
	hashVal := hasher.Sum64()

	return taggedRowData{
		Name:       s,
		PacketSize: i,
		hash:       hashVal,
	}
}

func TestWithTaggedStructData(t *testing.T) {

	data := []taggedRowData{
		newTaggedRowData("Chocolate Digestives", 12),
		newTaggedRowData("Tim Tams", 8),
		newTaggedRowData("Hobnobs", 10),
		newTaggedRowData("Peanut Butter Cookie", 8),
	}

	table := New(WithStructData(data))

	require.Equal(t, 2, len(table.cols))
	require.Equal(t, 4, len(table.rows))
	require.Equal(t, "Biscuit", table.cols[0].Title)
	require.Equal(t, "PacketSize", table.cols[1].Title)

	for i, r := range data {
		require.Equal(t, r.Name, table.rows[i].Data[0])
		require.Equal(t, strconv.Itoa(r.PacketSize), table.rows[i].Data[1])
	}
}

func TestRemoveRowsByIndex(t *testing.T) {
	data := []rowData{
		newRowData("Chocolate Digestives", 12),
		newRowData("Tim Tams", 8),
		newRowData("Hobnobs", 10),
		newRowData("Peanut Butter Cookie", 8),
	}

	table := New(WithStructData(data))

	require.True(t, table.RemoveRowByIndex(10))
	require.Equal(t, 4, len(table.rows))

	require.True(t, table.RemoveRowByIndex(0))
	require.Equal(t, 3, len(table.rows))
	require.Equal(t, "Tim Tams", table.rows[0].Data[0])
}

func TestRemoveSelectedRow(t *testing.T) {
	data := []rowData{
		newRowData("Chocolate Digestives", 12),
		newRowData("Tim Tams", 8),
		newRowData("Hobnobs", 10),
		newRowData("Peanut Butter Cookie", 8),
	}

	table := New(WithStructData(data))
	table.SetCursor(3)
	require.True(t, table.RemoveSelectedRow())
	require.Equal(t, 3, len(table.rows))

	// Cursor now on last remaining row
	require.Equal(t, 2, table.Cursor())
}

func TestRemoveRowByHash(t *testing.T) {
	data := []rowData{
		newRowData("Chocolate Digestives", 12),
		newRowData("Tim Tams", 8),
		newRowData("Hobnobs", 10),
		newRowData("Peanut Butter Cookie", 8),
	}

	table := New(WithStructData(data))
	table.SetCursor(3)
	require.True(t, table.RemoveRowByHash(0x40099b03ee546818))
	require.Equal(t, 3, len(table.rows))
	require.Equal(t, "Tim Tams", table.rows[0].Data[0])
}
func TestRemoveRow(t *testing.T) {
	chocolateDigestives := newRowData("Chocolate Digestives", 12)
	data := []rowData{
		chocolateDigestives,
		newRowData("Tim Tams", 8),
		newRowData("Hobnobs", 10),
		newRowData("Peanut Butter Cookie", 8),
	}

	table := New(WithStructData(data))
	table.SetCursor(3)
	require.True(t, table.RemoveRow(chocolateDigestives))
	require.Equal(t, 3, len(table.rows))
	require.Equal(t, "Tim Tams", table.rows[0].Data[0])
}

func skipIfGithubOnWindows(t *testing.T) {
	if _, github := os.LookupEnv("GITHUB_ACTION"); github && runtime.GOOS == "windows" {
		t.Skip("Skipping for github incompatibility")
	}
}
