package environment

import (
	"testing"
)

func TestVcloud(t *testing.T) {
	v := Vcloud{}
	if v.Name() != "Vcloud" {
		t.Fatal("Vcloud Name not correct")
	}
}
