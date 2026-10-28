package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"net-merge/internal/mobilequota"
)

func TestRunPrintsCurrentMonthlyUsageAndCap(t *testing.T) {
	root := t.TempDir()
	stats := filepath.Join(root, "bnep0", "statistics")
	if err := os.MkdirAll(stats, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stats, "rx_bytes"), []byte("1250000000\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stats, "tx_bytes"), []byte("250000000\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	config := mobilequota.Config{
		Interface:   "bnep0",
		SysClassNet: root,
		StatePath:   filepath.Join(root, "state.json"),
		LimitBytes:  500_000_000_000,
		WarnBytes:   450_000_000_000,
		Now:         func() time.Time { return time.Date(2026, 10, 28, 20, 0, 0, 0, time.UTC) },
	}
	var out bytes.Buffer
	if err := run(config, &out); err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), "Vodafone October: 1.50 GB / 500 GB\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
