//go:build windows
// +build windows

package collectors

import (
	model "vsysmon/internal/model"
)

type LoadCollector struct{}

func (c *LoadCollector) Name() string { return "load" }

func (c *LoadCollector) Collect(s *model.Sample) {
	// No-op
}
