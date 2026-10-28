package main

import (
	"os/exec"
	"strings"
	"testing"
)

// The installed command must not contain an automatic release checker. Reintroducing the updater
// would make the tray contact GitHub at sign-in and every few hours without the user asking it to.
func TestCommandHasNoAutomaticUpdateClient(t *testing.T) {
	cmd := exec.Command("go", "list", "-deps", ".")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("list command dependencies: %v", err)
	}
	for dep := range strings.Lines(string(out)) {
		if strings.TrimSpace(dep) == "net-merge/internal/update" {
			t.Fatal("command includes the automatic GitHub release checker")
		}
	}
}
