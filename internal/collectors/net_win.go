//go:build windows
// +build windows

package collectors

import (
	model "vsysmon/internal/model"
)

type TCPStateCollector struct{}

func (c *TCPStateCollector) Name() string { return "tcp_states" }

func (c *TCPStateCollector) Collect(s *model.Sample) {
	// No-op
}
