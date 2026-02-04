//go:build linux
// +build linux

package collectors

import (
	"os"
	"strconv"
	"strings"

	model "vsysmon/internal/model"
)

type LoadCollector struct{}

func (c *LoadCollector) Name() string { return "load" }

func (c *LoadCollector) Collect(s *model.Sample) {
	s.Load = readLoad()
}

var readLoad = func() float64 {
	b, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0 // возвращаем 0 при ошибке чтения
	}
	f := strings.Fields(string(b)) // разбивает по пробелам
	v, err := strconv.ParseFloat(f[0], 64)
	if err != nil {
		return 0 // возвращаем 0 при ошибке парсинга
	}
	return v
}
