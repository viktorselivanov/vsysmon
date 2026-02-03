//go:build linux
// +build linux

package collectors

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	model "vsysmon/model"
)

type cpuStat struct { // счётчики ядра
	user, nice, system, idle, iowait, irq, softirq, steal uint64
}

type CPUCollector struct {
	prev cpuStat
	init bool
}

func (c *CPUCollector) Name() string { return "cpu" }

func (c *CPUCollector) Collect(s *model.Sample) {
	cur := readCPUStat() // Читаем текущее состояние счётчиков.

	if !c.init { // вызываем только после инициализации
		c.prev = cur
		c.init = true
		return
	}

	d := cpuDelta(c.prev, cur)
	c.prev = cur

	s.CPUUser = d.user
	s.CPUSys = d.system
	s.CPUIdle = d.idle
}

var readCPUStat = func() cpuStat {
	f, _ := os.Open("/proc/stat") // открываем файл ядра
	defer f.Close()

	// читаем построчно
	sc := bufio.NewScanner(f)
	sc.Scan()
	fields := strings.Fields(sc.Text())

	// парсим числа
	val := func(i int) uint64 {
		v, _ := strconv.ParseUint(fields[i], 10, 64)
		return v
	}

	return cpuStat{
		user:    val(1),
		nice:    val(2),
		system:  val(3),
		idle:    val(4),
		iowait:  val(5),
		irq:     val(6),
		softirq: val(7),
		steal:   val(8),
	}
}

type cpuDeltaStat struct {
	user, system, idle float64
}

func cpuDelta(a, b cpuStat) cpuDeltaStat {
	prev := a.user + a.nice + a.system + a.idle + a.iowait + a.irq + a.softirq + a.steal
	cur := b.user + b.nice + b.system + b.idle + b.iowait + b.irq + b.softirq + b.steal

	// защита от деления на ноль.
	dt := float64(cur - prev)
	if dt == 0 {
		return cpuDeltaStat{}
	}

	return cpuDeltaStat{ // расчёт процентов
		user:   float64(b.user-a.user) / dt * 100,
		system: float64(b.system-a.system) / dt * 100,
		idle:   float64(b.idle-a.idle) / dt * 100,
	}
}
