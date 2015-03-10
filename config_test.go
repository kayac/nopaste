package nopaste_test

import (
	"fmt"
	"testing"

	"github.com/kayac/nopaste"
)

func TestLoadConfig(t *testing.T) {
	c, err := nopaste.LoadConfig("test/example.yaml")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%#v", c)
}
