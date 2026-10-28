// Package mobilequota persists byte usage for one network interface and disables that connection
// before a monthly mobile-data threshold is crossed.
package mobilequota

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Interface     string
	SysClassNet   string
	StatePath     string
	LimitBytes    uint64
	WarnBytes     uint64
	Now           func() time.Time
	Disconnect    func() error
	Autoconnect   func(bool) error
	Notify        func(string) error
	BlockedNotify func(string) error
}

type Result struct {
	Month     string
	UsedBytes uint64
	Warned    bool
	Blocked   bool
}

type state struct {
	Month     string `json:"month"`
	LastRX    uint64 `json:"last_rx"`
	LastTX    uint64 `json:"last_tx"`
	UsedBytes uint64 `json:"used_bytes"`
	Warned    bool   `json:"warned"`
	Blocked   bool   `json:"blocked"`
}

func Sample(c Config) (Result, error) {
	if c.Now == nil {
		c.Now = time.Now
	}
	if c.SysClassNet == "" {
		c.SysClassNet = "/sys/class/net"
	}
	if c.Interface == "" || c.StatePath == "" || c.LimitBytes == 0 {
		return Result{}, errors.New("interface, state path and limit are required")
	}

	s, existed, err := load(c.StatePath)
	if err != nil {
		return Result{}, err
	}
	month := c.Now().Format("2006-01")
	rx, tx, present, err := counters(c.SysClassNet, c.Interface)
	if err != nil {
		return Result{}, err
	}

	if existed && s.Month != month {
		s = state{Month: month, LastRX: rx, LastTX: tx}
		if c.Autoconnect != nil {
			if err := c.Autoconnect(true); err != nil {
				return Result{}, err
			}
		}
	} else {
		if !existed {
			s.Month = month
		}
		if present {
			s.UsedBytes += delta(rx, s.LastRX) + delta(tx, s.LastTX)
			s.LastRX, s.LastTX = rx, tx
		}
	}

	if !s.Blocked && s.UsedBytes >= c.LimitBytes {
		if c.Autoconnect != nil {
			if err := c.Autoconnect(false); err != nil {
				return Result{}, err
			}
		}
		if c.Disconnect != nil {
			if err := c.Disconnect(); err != nil {
				return Result{}, err
			}
		}
		if c.BlockedNotify != nil {
			if err := c.BlockedNotify(fmt.Sprintf("Vodafone paused at %.1f GB", float64(s.UsedBytes)/1e9)); err != nil {
				return Result{}, err
			}
		}
		s.Blocked = true
	}
	if !s.Warned && s.UsedBytes >= c.WarnBytes {
		if c.Notify != nil {
			if err := c.Notify(fmt.Sprintf("Vodafone usage is %.1f GB", float64(s.UsedBytes)/1e9)); err != nil {
				return Result{}, err
			}
		}
		s.Warned = true
	}
	if err := save(c.StatePath, s); err != nil {
		return Result{}, err
	}
	return Result{Month: s.Month, UsedBytes: s.UsedBytes, Warned: s.Warned, Blocked: s.Blocked}, nil
}

func delta(current, previous uint64) uint64 {
	if current >= previous {
		return current - previous
	}
	return current
}

func counters(root, iface string) (uint64, uint64, bool, error) {
	read := func(name string) (uint64, error) {
		b, err := os.ReadFile(filepath.Join(root, iface, "statistics", name))
		if err != nil {
			return 0, err
		}
		return strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
	}
	rx, err := read("rx_bytes")
	if os.IsNotExist(err) {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, err
	}
	tx, err := read("tx_bytes")
	if err != nil {
		return 0, 0, false, err
	}
	return rx, tx, true, nil
}

func load(path string) (state, bool, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return state{}, false, nil
	}
	if err != nil {
		return state{}, false, err
	}
	var s state
	if err := json.Unmarshal(b, &s); err != nil {
		return state{}, false, fmt.Errorf("read quota state: %w", err)
	}
	return s, true, nil
}

func save(path string, s state) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".new"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
