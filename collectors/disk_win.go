//go:build windows
// +build windows

package collectors

import (
	model "vsysmon/model"
)

type diskStat struct {
	reads, writes, sectors uint64
}

type DiskCollector struct {
	prev diskStat
	init bool
}

func (c *DiskCollector) Name() string { return "disk" }

func (c *DiskCollector) Collect(s *model.Sample) {
	// No-op
}
