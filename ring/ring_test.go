//go:build linux
// +build linux

package ring_test

import (
	"testing"
	"vsysmon/model"
	"vsysmon/ring"
)

func TestRingPushAndSnapshot(t *testing.T) {
	// Инициализируем кольцо на 3 сэмпла
	ring.RingInit(3)

	// Кольцо ещё не заполнено, RingSnapshot должен вернуть nil
	if snapshots := ring.RingSnapshot(); snapshots != nil {
		t.Errorf("expected nil snapshot before filling ring, got %v", snapshots)
	}

	// Добавляем 3 сэмпла
	samples := []model.Sample{
		{Load: 1.0},
		{Load: 2.0},
		{Load: 3.0},
	}

	for _, s := range samples {
		ring.RingPush(s)
	}

	// Теперь кольцо заполнено, RingSnapshot должен вернуть копию
	snapshots := ring.RingSnapshot()
	if len(snapshots) != 3 {
		t.Fatalf("expected 3 samples in snapshot, got %d", len(snapshots))
	}

	for i, s := range samples {
		if snapshots[i].Load != s.Load {
			t.Errorf("expected snapshot[%d].Load=%.1f, got %.1f", i, s.Load, snapshots[i].Load)
		}
	}

	// Добавляем ещё один сэмпл → переполнение кольца
	ring.RingPush(model.Sample{Load: 4.0})

	// Последний снимок должен содержать новый сэмпл на позиции 0
	snapshots = ring.RingSnapshot()
	if snapshots[0].Load != 4.0 {
		t.Errorf("expected snapshot[0].Load=4.0 after wraparound, got %.1f", snapshots[0].Load)
	}
}

func TestSaveAndLastSnapshot(t *testing.T) {
	s := model.Snapshot{Load: 42.0}
	ring.SaveSnapshot(s)

	// Проверяем чтение
	last := ring.LastSnapshot()
	if last.Load != 42.0 {
		t.Errorf("expected last snapshot load=42.0, got %.1f", last.Load)
	}

	// Меняем snapshot и проверяем, что LastSnapshot обновляется
	s2 := model.Snapshot{Load: 100.0}
	ring.SaveSnapshot(s2)
	last = ring.LastSnapshot()
	if last.Load != 100.0 {
		t.Errorf("expected last snapshot load=100.0, got %.1f", last.Load)
	}
}
