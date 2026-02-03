//go:build linux
// +build linux

package collectors

import (
	"os"
	"strconv"
	"strings"
	model "vsysmon/model"
)

type LoadCollector struct{}

func (c *LoadCollector) Name() string { return "load" }

func (c *LoadCollector) Collect(s *model.Sample) {

	s.Load = readLoad()
}

var readLoad = func() float64 {
	b, _ := os.ReadFile("/proc/loadavg")
	f := strings.Fields(string(b)) // разбивает по пробелам
	v, _ := strconv.ParseFloat(f[0], 64)
	return v
}
