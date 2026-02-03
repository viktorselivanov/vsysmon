//go:build linux
// +build linux

package collectors

import (
	"testing"

	model "vsysmon/model"
)

func TestTopTalkerCollector_Collect(t *testing.T) {
	origRead := readNetFileFn
	defer func() { readNetFileFn = origRead }()

	// Мок функции readNetFileFn: увеличиваем значения на второй итерации
	call := 0
	readNetFileFn = func(path, proto string, protoMap map[string]uint64, flowMap map[string]uint64) {
		// Первая итерация — все нули
		if call == 0 {
			switch proto {
			case "TCP":
				protoMap["TCP"] = 0
				flowMap["TCP|1.2.3.4:1234->5.6.7.8:80"] = 0
			case "UDP":
				protoMap["UDP"] = 0
				flowMap["UDP|10.0.0.1:1111->10.0.0.2:2222"] = 0
			}
		} else {
			// Вторая итерация — увеличиваем значения
			switch proto {
			case "TCP":
				protoMap["TCP"] = 1000
				flowMap["TCP|1.2.3.4:1234->5.6.7.8:80"] = 1000
			case "UDP":
				protoMap["UDP"] = 500
				flowMap["UDP|10.0.0.1:1111->10.0.0.2:2222"] = 500
			}
		}
	}

	collector := &TopTalkerCollector{
		prevProto: make(map[string]uint64),
		prevFlows: make(map[string]uint64),
	}
	s := &model.Sample{}

	// Первая итерация - init
	collector.Collect(s)
	if s.ProtoTop != nil || s.FlowTop != nil {
		t.Errorf("expected nil on first call, got %+v %+v", s.ProtoTop, s.FlowTop)
	}

	// Вторая итерация - дельты
	call++ // отмечаем вторую итерацию
	s2 := &model.Sample{}
	collector.Collect(s2)

	// Проверка протоколов
	if len(s2.ProtoTop) != 2 {
		t.Fatalf("expected 2 protoTop entries, got %d", len(s2.ProtoTop))
	}

	for _, p := range s2.ProtoTop {
		switch p.Proto {
		case "TCP":
			if p.Bytes != 1000 {
				t.Errorf("expected TCP=1000, got %d", p.Bytes)
			}
		case "UDP":
			if p.Bytes != 500 {
				t.Errorf("expected UDP=500, got %d", p.Bytes)
			}
		default:
			t.Errorf("unexpected proto %s", p.Proto)
		}
	}

	// Проверка потоков
	if len(s2.FlowTop) != 2 {
		t.Fatalf("expected 2 flowTop entries, got %d", len(s2.FlowTop))
	}

	foundFlow1, foundFlow2 := false, false
	for _, f := range s2.FlowTop {
		switch f.Src {
		case "1.2.3.4:1234":
			foundFlow1 = true
			if f.BPS != 1000 || f.Proto != "TCP" {
				t.Errorf("unexpected flow %+v", f)
			}
		case "10.0.0.1:1111":
			foundFlow2 = true
			if f.BPS != 500 || f.Proto != "UDP" {
				t.Errorf("unexpected flow %+v", f)
			}
		}
	}
	if !foundFlow1 || !foundFlow2 {
		t.Errorf("missing expected flows: %+v", s2.FlowTop)
	}
}
