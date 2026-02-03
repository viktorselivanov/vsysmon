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

type diskStat struct {
	reads, writes, sectors uint64
}

type DiskCollector struct {
	prev diskStat
	init bool
}

func (c *DiskCollector) Name() string { return "disk" }

func (c *DiskCollector) Collect(s *model.Sample) {
	cur := readDiskStat() // считываем текущее состояние всех дисков

	if !c.init {
		c.prev = cur
		c.init = true
		return
	}

	d := diskDelta(c.prev, cur) // считаем разницу между прошлым и текущим
	c.prev = cur

	s.DiskTPS = d.tps
	s.DiskKBs = d.kbs
}

// собираем общую нагрузку системы
var readDiskStat = func() diskStat {
	f, _ := os.Open("/proc/diskstats")
	defer f.Close()

	var r diskStat
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fs := strings.Fields(sc.Text())
		if len(fs) < 14 {
			continue
		}
		if strings.HasPrefix(fs[2], "sd") || strings.HasPrefix(fs[2], "nvme") {
			rd, _ := strconv.ParseUint(fs[3], 10, 64)
			wr, _ := strconv.ParseUint(fs[7], 10, 64)
			sec, _ := strconv.ParseUint(fs[5], 10, 64)
			wsec, _ := strconv.ParseUint(fs[9], 10, 64)

			r.reads += rd
			r.writes += wr
			r.sectors += sec + wsec
		}
	}
	return r
}

type diskDeltaStat struct {
	tps float64
	kbs float64
}

// рачитываем  скорость
func diskDelta(a, b diskStat) diskDeltaStat {
	dr := b.reads + b.writes - a.reads - a.writes
	ds := b.sectors - a.sectors

	return diskDeltaStat{
		tps: float64(dr),
		kbs: float64(ds*512) / 1024,
	}
}
