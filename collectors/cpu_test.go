//go:build linux
// +build linux

package collectors

import (
	"testing"

	model "vsysmon/model"
)

// тестируем CPUCollector
func TestCPUCollector_Collect(t *testing.T) {
	// сохраняем оригинальную функцию, чтобы вернуть после теста
	origRead := readCPUStat
	defer func() { readCPUStat = origRead }()

	// создаём фиктивные значения CPU
	fakeStats := []cpuStat{
		{user: 100, nice: 0, system: 50, idle: 850},
		{user: 120, nice: 0, system: 70, idle: 860},
	}

	i := 0
	readCPUStat = func() cpuStat {
		if i >= len(fakeStats) {
			return fakeStats[len(fakeStats)-1]
		}
		s := fakeStats[i]
		i++
		return s
	}

	collector := &CPUCollector{}

	s := &model.Sample{}

	// первый вызов должен только инициализировать prev
	collector.Collect(s)
	if s.CPUUser != 0 || s.CPUSys != 0 || s.CPUIdle != 0 {
		t.Errorf("first collect should not set values, got %+v", s)
	}

	// второй вызов должен вычислить дельту
	collector.Collect(s)

	expectedUser := float64(120-100) / float64(120+70+860-(100+50+850)) * 100
	expectedSys := float64(70-50) / float64(120+70+860-(100+50+850)) * 100
	expectedIdle := float64(860-850) / float64(120+70+860-(100+50+850)) * 100

	if s.CPUUser != expectedUser {
		t.Errorf("CPUUser expected %.2f, got %.2f", expectedUser, s.CPUUser)
	}
	if s.CPUSys != expectedSys {
		t.Errorf("CPUSys expected %.2f, got %.2f", expectedSys, s.CPUSys)
	}
	if s.CPUIdle != expectedIdle {
		t.Errorf("CPUIdle expected %.2f, got %.2f", expectedIdle, s.CPUIdle)
	}
}
