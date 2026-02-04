//go:build linux
// +build linux

package collectors

import (
	"testing"

	model "vsysmon/internal/model"
)

func TestDiskCollector_Collect(t *testing.T) {
	// Сохраняем оригинальную функцию, чтобы вернуть после теста
	origRead := readDiskStat
	defer func() { readDiskStat = origRead }()

	// Подготавливаем фейковые значения счётчиков
	stats := []diskStat{
		{reads: 100, writes: 50, sectors: 1000}, // первый вызов (инициализация)
		{reads: 130, writes: 70, sectors: 1600}, // второй вызов (дельта)
	}
	call := 0

	// Мокаем readDiskStat
	readDiskStat = func() diskStat {
		v := stats[call]
		call++
		return v
	}

	collector := &DiskCollector{}
	s := &model.Sample{}

	// Первый Collect — только инициализация, без записи в Sample
	collector.Collect(s)

	if s.DiskTPS != 0 || s.DiskKBs != 0 {
		t.Errorf("expected zero values on first collect, got TPS=%.2f KBs=%.2f",
			s.DiskTPS, s.DiskKBs)
	}

	// Второй Collect — считаем дельту
	collector.Collect(s)

	// проверяем, что подсчёт операций выполняется корректно
	expectedTPS := float64(50)

	// проверяем, что подсчёт секторов выполняется корректно
	expectedKBs := float64(600*512) / 1024

	if s.DiskTPS != expectedTPS {
		t.Errorf("expected TPS %.2f, got %.2f", expectedTPS, s.DiskTPS)
	}

	if s.DiskKBs != expectedKBs {
		t.Errorf("expected KB/s %.2f, got %.2f", expectedKBs, s.DiskKBs)
	}
}
