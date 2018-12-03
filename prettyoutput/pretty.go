package pretty

import (
	"fmt"
	"strings"

	"github.com/bclicn/color"
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
	fmt.Printf("%s\n"+
		"Test: %s\n"+
		"Test-info: blah blah\n", separator, t.Name)
}

func (t PrettyTest) PrintResult(result bool) {
	var output string
	if result {
		output = color.BGreen("PASS")
	} else {
		output = color.BRed("FAIL")
	}
	fmt.Printf("Test result: [ %s ]\n", output)
}

func (t PrettyTest) EndTest() {
	fmt.Printf("%s\n", separator)
}
