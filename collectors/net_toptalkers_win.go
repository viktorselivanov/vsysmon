//go:build windows
// +build windows

package collectors

import (
	model "vsysmon/model"
)

/*
Top Talkers:
 1) по протоколам (TCP/UDP/ICMP) — % от общего трафика
 2) по потокам src:port -> dst:port — bytes/sec
*/

type TopTalkerCollector struct{}

func (c *TopTalkerCollector) Name() string { return "top_talkers" }

func (c *TopTalkerCollector) Collect(s *model.Sample) {
	// No-op
}
