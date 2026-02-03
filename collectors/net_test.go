//go:build linux
// +build linux

package collectors

import (
	"testing"

	model "vsysmon/model"
)

func TestTCPStateCollector_Collect(t *testing.T) {
	// Сохраняем оригинальную функцию
	origRead := readTCPStates
	defer func() { readTCPStates = origRead }()

	// Мокаем readTCPStates
	readTCPStates = func() map[string]int {
		return map[string]int{
			"ESTABLISHED": 2,
			"LISTEN":      1,
			"CLOSE_WAIT":  1,
		}
	}

	collector := &TCPStateCollector{}
	s := &model.Sample{}

	collector.Collect(s)

	if len(s.TCPStates) != 3 {
		t.Errorf("expected 3 TCP states, got %d", len(s.TCPStates))
	}

	if s.TCPStates["ESTABLISHED"] != 2 {
		t.Errorf("expected ESTABLISHED=2, got %d", s.TCPStates["ESTABLISHED"])
	}

	if s.TCPStates["LISTEN"] != 1 {
		t.Errorf("expected LISTEN=1, got %d", s.TCPStates["LISTEN"])
	}

	if s.TCPStates["CLOSE_WAIT"] != 1 {
		t.Errorf("expected CLOSE_WAIT=1, got %d", s.TCPStates["CLOSE_WAIT"])
	}
}
