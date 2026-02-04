//go:build windows
// +build windows

package collectors

import (
	model "vsysmon/internal/model"
)

type TopTalkerCollector struct{}

func (c *TopTalkerCollector) Name() string { return "top_talkers" }

func (c *TopTalkerCollector) Collect(s *model.Sample) {
	// No-op
}
