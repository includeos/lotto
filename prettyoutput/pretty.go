package pretty

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

const width = 80

var separator = strings.Repeat("=", width)

type PrettyTest struct {
	Name  string
	table *tablewriter.Table
}

func NewPrettyTest(name string) PrettyTest {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	return PrettyTest{Name: name, table: table}
}

func (t PrettyTest) PrintHeader() {
	title := fmt.Sprintf(" test: %s ", t.Name)
	lenTitle := len(title)
	fillEachSide := (width - lenTitle) / 2
	fill := strings.Repeat("=", fillEachSide)
	fmt.Printf("%s%s%s\n", fill, title, fill)
}

func (t PrettyTest) PrintTable(data [][]string) {
	t.table.ClearRows()
	t.table.AppendBulk(data)
	t.table.Render()
}

func (t PrettyTest) EndTest() {
	fmt.Printf("%s\n", separator)
}
