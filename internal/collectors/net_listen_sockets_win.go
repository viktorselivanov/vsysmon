//go:build windows
// +build windows

package collectors

import (
	model "vsysmon/internal/model"
)

type ListenSocketCollector struct{}

func (c *ListenSocketCollector) Name() string { return "listen" }

func (c *ListenSocketCollector) Collect(s *model.Sample) {
	// No-op
}
