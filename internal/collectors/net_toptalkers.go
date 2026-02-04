//go:build linux
// +build linux

package collectors

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	model "vsysmon/internal/model"
)

var readNetFileFn = readNetFile

// TopTalkerCollector хранит состояние между вызовами.
type TopTalkerCollector struct {
	prevProto map[string]uint64 // предыдущие значения байт по протоколам
	prevFlows map[string]uint64 // предыдущие значения байт по потокам
	init      bool              // флаг первой инициализации
	mu        sync.Mutex        // защита состояния коллектора
}

func (c *TopTalkerCollector) Name() string { return "top_talkers" }

func (c *TopTalkerCollector) Collect(s *model.Sample) {
	c.mu.Lock()
	defer c.mu.Unlock()

	curProto := map[string]uint64{}
	curFlows := map[string]uint64{}

	readNetFileFn("/proc/net/tcp", "TCP", curProto, curFlows)
	readNetFileFn("/proc/net/udp", "UDP", curProto, curFlows)
	readNetFileFn("/proc/net/icmp", "ICMP", curProto, curFlows)

	// первая инициализация — только сохраняем значения
	if !c.init {
		c.prevProto = curProto
		c.prevFlows = curFlows
		c.init = true
		s.ProtoTop = nil
		s.FlowTop = nil
		return
	}

	// вычисляем дельты по протоколам
	diffProto := map[string]uint64{}
	total := uint64(0)
	for p, v := range curProto {
		prev := c.prevProto[p]
		var d uint64
		if v >= prev {
			d = v - prev
		} else {
			d = 0
		}
		diffProto[p] = d
		total += d
	}
	protoTop := make([]model.ProtoTalker, 0, len(diffProto))
	for p, v := range diffProto {
		perc := 0.0
		if total > 0 {
			perc = float64(v) / float64(total) * 100
		}
		protoTop = append(protoTop, model.ProtoTalker{
			Proto: p,
			Bytes: v,
			Perc:  perc,
		})
	}

	sort.Slice(protoTop, func(i, j int) bool { return protoTop[i].Perc > protoTop[j].Perc })

	// формируем топ по потокам
	flowTop := make([]model.FlowTalker, 0, len(curFlows))
	for k, v := range curFlows {
		prev := c.prevFlows[k]

		var bps uint64
		if v >= prev {
			bps = v - prev
		} else {
			bps = 0
		}

		if bps == 0 {
			continue
		}

		src, dst, proto := parseFlowKey(k)
		flowTop = append(flowTop, model.FlowTalker{
			Src:   src,
			Dst:   dst,
			Proto: proto,
			BPS:   bps,
		})
	}

	sort.Slice(flowTop, func(i, j int) bool { return flowTop[i].BPS > flowTop[j].BPS })

	// сохраняем текущее состояние для следующей итерации
	c.prevProto = curProto
	c.prevFlows = curFlows

	s.ProtoTop = protoTop
	s.FlowTop = flowTop
}

func readNetFile(path, proto string, protoMap, flowMap map[string]uint64) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close() // гарантируем закрытие

	sc := bufio.NewScanner(f) // читаем построчно
	first := true
	for sc.Scan() {
		if first { // пропуск заголовка
			first = false
			continue
		}
		fields := strings.Fields(sc.Text()) // разбиваем поля
		if len(fields) < 10 {
			continue
		}

		local := fields[1]
		remote := fields[2]
		txrx, err := strconv.ParseUint(fields[9], 10, 64) // собираем объём информации в байтах
		if err != nil {
			continue // пропускаем некорректное значение
		}
		src := parseAddr(local)  // извлекаем адрес источника
		dst := parseAddr(remote) // извлекаем адрес назначения

		protoMap[proto] += txrx

		key := fmt.Sprintf("%s|%s->%s", proto, src, dst) // формируем ключ
		flowMap[key] += txrx
	}
}

func parseAddr(hex string) string {
	parts := strings.Split(hex, ":") // разделяем на чести по ip и port
	if len(parts) != 2 {
		return hex
	}

	ipHex := parts[0]
	portHex := parts[1]

	ip := make([]string, 4) // переводим hex
	for i := 0; i < 4; i++ {
		b, err := strconv.ParseInt(ipHex[i*2:i*2+2], 16, 64)
		if err != nil {
			b = 0 // безопасно заменяем на 0
		}

		ip[3-i] = strconv.Itoa(int(b))
	}

	port, err := strconv.ParseInt(portHex, 16, 64) // переводим hex
	if err != nil {
		port = 0 // безопасно заменяем на 0
	}

	return fmt.Sprintf("%s.%s.%s.%s:%d",
		ip[0], ip[1], ip[2], ip[3], port)
}

func parseFlowKey(k string) (src, dst, proto string) { // разбораем ключ для FlowTalker
	p1 := strings.Split(k, "|")
	if len(p1) != 2 {
		return "", "", ""
	}
	proto = p1[0]
	p2 := strings.Split(p1[1], "->")
	if len(p2) != 2 {
		return "", "", proto
	}
	return p2[0], p2[1], proto
}
