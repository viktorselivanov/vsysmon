//go:build windows
// +build windows

package collectors

import (
	model "vsysmon/internal/model"
)

type FSCollector struct{}

func (c *FSCollector) Name() string { return "fs" }

func (c *FSCollector) Collect(s *model.Sample) {
	// No-op
}
