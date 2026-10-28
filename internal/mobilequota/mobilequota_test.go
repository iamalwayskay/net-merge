package mobilequota

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSampleDisconnectsAtLimitAndPersistsBlock(t *testing.T) {
	root := t.TempDir()
	writeCounters(t, root, 350, 151)
	var disconnected int
	var autoconnect []bool
	var blockedNotices []string
	cfg := Config{
		Interface:   "bnep0",
		SysClassNet: root,
		StatePath:   filepath.Join(root, "state.json"),
		LimitBytes:  500,
		WarnBytes:   400,
		Now:         func() time.Time { return time.Date(2026, 10, 28, 20, 0, 0, 0, time.UTC) },
		Disconnect:  func() error { disconnected++; return nil },
		Autoconnect: func(on bool) error { autoconnect = append(autoconnect, on); return nil },
		BlockedNotify: func(message string) error {
			blockedNotices = append(blockedNotices, message)
			return nil
		},
	}

	result, err := Sample(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if result.UsedBytes != 501 || !result.Blocked {
		t.Fatalf("result = %+v, want 501 bytes used and blocked", result)
	}
	if disconnected != 1 || len(autoconnect) != 1 || autoconnect[0] {
		t.Fatalf("disconnects=%d autoconnect=%v, want one disconnect and autoconnect off", disconnected, autoconnect)
	}
	if len(blockedNotices) != 1 || blockedNotices[0] != "Vodafone paused at 0.0 GB" {
		t.Fatalf("blocked notices = %v, want one cap notice", blockedNotices)
	}

	result, err = Sample(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Blocked || disconnected != 1 || len(autoconnect) != 1 || len(blockedNotices) != 1 {
		t.Fatalf("second sample repeated block actions: result=%+v disconnects=%d autoconnect=%v notices=%v", result, disconnected, autoconnect, blockedNotices)
	}
}

func TestSampleResetsAtNewCalendarMonthAndReenablesConnection(t *testing.T) {
	root := t.TempDir()
	writeCounters(t, root, 200, 100)
	now := time.Date(2026, 10, 31, 23, 59, 0, 0, time.UTC)
	var autoconnect []bool
	cfg := Config{
		Interface:   "bnep0",
		SysClassNet: root,
		StatePath:   filepath.Join(root, "state.json"),
		LimitBytes:  500,
		WarnBytes:   400,
		Now:         func() time.Time { return now },
		Autoconnect: func(on bool) error { autoconnect = append(autoconnect, on); return nil },
	}
	if _, err := Sample(cfg); err != nil {
		t.Fatal(err)
	}

	now = time.Date(2026, 11, 1, 0, 1, 0, 0, time.UTC)
	result, err := Sample(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if result.Month != "2026-11" || result.UsedBytes != 0 || result.Blocked {
		t.Fatalf("new month result = %+v, want a fresh unblocked month", result)
	}
	if len(autoconnect) != 1 || !autoconnect[0] {
		t.Fatalf("autoconnect calls = %v, want connection re-enabled", autoconnect)
	}

	writeCounters(t, root, 250, 125)
	result, err = Sample(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if result.UsedBytes != 75 {
		t.Fatalf("new-month usage = %d, want 75", result.UsedBytes)
	}
}

func TestSampleWarnsOnceBeforeTheLimit(t *testing.T) {
	root := t.TempDir()
	writeCounters(t, root, 300, 101)
	var notices []string
	cfg := Config{
		Interface:   "bnep0",
		SysClassNet: root,
		StatePath:   filepath.Join(root, "state.json"),
		LimitBytes:  500,
		WarnBytes:   400,
		Now:         func() time.Time { return time.Date(2026, 10, 28, 20, 0, 0, 0, time.UTC) },
		Notify:      func(message string) error { notices = append(notices, message); return nil },
	}

	result, err := Sample(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Warned || len(notices) != 1 || notices[0] != "Vodafone usage is 0.0 GB" {
		t.Fatalf("result=%+v notices=%v, want one warning", result, notices)
	}
	if _, err := Sample(cfg); err != nil {
		t.Fatal(err)
	}
	if len(notices) != 1 {
		t.Fatalf("warning repeated: %v", notices)
	}
}

func writeCounters(t *testing.T, root string, rx, tx uint64) {
	t.Helper()
	dir := filepath.Join(root, "bnep0", "statistics")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rx_bytes"), []byte(u64(rx)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tx_bytes"), []byte(u64(tx)), 0o644); err != nil {
		t.Fatal(err)
	}
}

func u64(n uint64) string {
	if n == 0 {
		return "0\n"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:]) + "\n"
}
