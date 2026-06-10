package main

import (
	"os/exec"
	"strings"
	"testing"
)

var packages = []struct {
	name string
	path string
}{
	{"protocol", "./protocol"},
	{"types", "./types"},
	{"config", "./config"},
	{"logger", "./logger"},
	{"datastore", "./datastore"},
	{"commands", "./commands"},
}

func TestAll(t *testing.T) {
	total := 0
	for _, pkg := range packages {
		t.Run(pkg.name, func(t *testing.T) {
			out, err := exec.Command("go", "test", "-v", "-count=1", pkg.path).CombinedOutput()
			if err != nil {
				t.Fatalf("FAIL:\n%s", out)
			}

			count := strings.Count(string(out), "--- PASS:")
			total += count
			t.Logf("passed %d tests", count)
		})
	}
	t.Logf("total: %d tests passed", total)
}
