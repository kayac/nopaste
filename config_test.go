package nopaste_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/kayac/nopaste"
)

func TestLoadConfig(t *testing.T) {
	base := "http://nopaste.example.com"
	os.Setenv("BASE_URL", base)
	c, err := nopaste.LoadConfig(context.Background(), "test/example.yaml")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%#v", c)
	if c.BaseURL != base {
		t.Errorf("unexpected BaseURL: %s", c.BaseURL)
	}
}
