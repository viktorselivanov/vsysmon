//go:build linux
// +build linux

package collectors

import (
	"bufio"
	"os"
	"strings"
	model "vsysmon/model"
)

type TCPStateCollector struct{}

func (c *TCPStateCollector) Name() string { return "tcp_states" }

func (c *TCPStateCollector) Collect(s *model.Sample) {
	s.TCPStates = readTCPStates()
}

var readTCPStates = func() map[string]int {
	m := make(map[string]int) // мапа для подсчёта состояний

	read := func(path string) {
		f, err := os.Open(path)
		if err != nil {
			return
		}
		defer f.Close()

		sc := bufio.NewScanner(f) // читаем построчно
		sc.Scan()
		for sc.Scan() {
			fs := strings.Fields(sc.Text()) // разбиваем на поля
			if len(fs) < 4 {
				continue
			}
			st := fs[3]
			m[tcpState(st)]++
		}
	}

	read("/proc/net/tcp")
	read("/proc/net/tcp6") // учитываем IPv6
	return m
}

func tcpState(h string) string { // переводи hex кода состояния в читабельный вид
	switch h {
	case "01":
		return "ESTABLISHED"
	case "02":
		return "SYN_SENT"
	case "03":
		return "SYN_RECV"
	case "04":
		return "FIN_WAIT1"
	case "05":
		return "FIN_WAIT2"
	case "06":
		return "TIME_WAIT"
	case "07":
		return "CLOSE"
	case "08":
		return "CLOSE_WAIT"
	case "09":
		return "LAST_ACK"
	case "0A":
		return "LISTEN"
	case "0B":
		return "CLOSING"
	default:
		return "UNK"
	}
}
