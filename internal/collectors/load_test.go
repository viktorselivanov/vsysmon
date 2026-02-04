//go:build linux
// +build linux

package collectors

import (
	"testing"

	model "vsysmon/internal/model"
)

func TestLoadCollector_Collect(t *testing.T) {
	// Сохраняем оригинальную функцию, чтобы вернуть после теста
	origRead := readLoad
	defer func() { readLoad = origRead }()

	// Мокаем readLoad для теста
	fakeLoad := 1.23
	readLoad = func() float64 {
		return fakeLoad
	}

	collector := &LoadCollector{}
	s := &model.Sample{}

	collector.Collect(s)

	if s.Load != fakeLoad {
		t.Errorf("expected Load %.2f, got %.2f", fakeLoad, s.Load)
	}
}
