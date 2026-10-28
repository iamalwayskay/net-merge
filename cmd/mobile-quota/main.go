package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"net-merge/internal/mobilequota"
)

const (
	connectionName = "Kaylum's S23+ Network"
	warnBytes      = 450_000_000_000
	limitBytes     = 500_000_000_000
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fatal(err)
	}
	config := mobilequota.Config{
		Interface:  "bnep0",
		StatePath:  filepath.Join(home, ".local", "state", "net-merge", "mobile-quota.json"),
		LimitBytes: limitBytes,
		WarnBytes:  warnBytes,
		Autoconnect: func(on bool) error {
			value := "no"
			if on {
				value = "yes"
			}
			return command("nmcli", "connection", "modify", connectionName, "connection.autoconnect", value)
		},
		Disconnect: func() error {
			return command("nmcli", "--wait", "15", "connection", "down", connectionName)
		},
		Notify: func(message string) error {
			return command("notify-send", "--urgency=critical", "Vodafone data warning", message+" this month; mobile data pauses at 500 GB.")
		},
		BlockedNotify: func(message string) error {
			return command("notify-send", "--urgency=critical", "Vodafone data paused", message+" this month. Plusnet and TalkTalk remain connected.")
		},
	}
	if err := run(config, os.Stdout); err != nil {
		fatal(err)
	}
}

func run(config mobilequota.Config, out io.Writer) error {
	result, err := mobilequota.Sample(config)
	if err != nil {
		return err
	}
	month, err := time.Parse("2006-01", result.Month)
	if err != nil {
		return err
	}
	status := ""
	if result.Blocked {
		status = " — paused"
	}
	_, err = fmt.Fprintf(out, "Vodafone %s: %.2f GB / %.0f GB%s\n",
		month.Format("January"), float64(result.UsedBytes)/1e9, float64(config.LimitBytes)/1e9, status)
	return err
}

func command(name string, args ...string) error {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "net-merge-mobile-quota:", err)
	os.Exit(1)
}
