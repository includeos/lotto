package pretty

import (
	"fmt"
	"strings"

	"github.com/logrusorgru/aurora"
)

const width = 80

var separator = strings.Repeat("=", width)

type PrettyTest struct {
	Name string
}

func NewPrettyTest(name string) PrettyTest {
	return PrettyTest{Name: name}
}

func (t PrettyTest) PrintHeader() {
	fmt.Println(separator)
}

func (t PrettyTest) Print(x interface{}) {
	fmt.Println(x)
}

func (t PrettyTest) PrintResult(result bool) {
	var output string
	if result {
		output = aurora.BgGreen(" PASS ").String()
	} else {
		output = aurora.BgRed(" FAIL ").String()
	}
	fmt.Printf("Test result: [ %s ]\n", output)
}

func (t PrettyTest) EndTest() {
	fmt.Printf("%s\n", separator)
}
