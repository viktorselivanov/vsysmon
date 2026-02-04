//go:build linux
// +build linux

package collectors

import (
	"testing"

	model "vsysmon/internal/model"
)

const (
	protoTCP = "TCP"
	protoUDP = "UDP"
)

func TestTopTalkerCollector_Collect(t *testing.T) {
	origRead := readNetFileFn
	defer func() { readNetFileFn = origRead }()

	// Мок функции readNetFileFn: увеличиваем значения на второй итерации
	call := 0
	readNetFileFn = mockReadNetFile(&call)

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

	checkProtoTop(t, s2.ProtoTop)
	checkFlowTop(t, s2.FlowTop)
}

func mockReadNetFile(call *int) func(_, proto string, protoMap map[string]uint64, flowMap map[string]uint64) {
	return func(_, _ string, protoMap map[string]uint64, flowMap map[string]uint64) {
		if *call == 0 {
			setFlowValues(protoMap, flowMap, 0)
		} else {
			setFlowValues(protoMap, flowMap, 1000)
		}
	}
}

func setFlowValues(protoMap, flowMap map[string]uint64, val uint64) {
	protoMap[protoTCP] = val
	protoMap[protoUDP] = val / 2
	flowMap[protoTCP+"|1.2.3.4:1234->5.6.7.8:80"] = val
	flowMap[protoUDP+"|10.0.0.1:1111->10.0.0.2:2222"] = val / 2
}

func checkProtoTop(t *testing.T, protoTop []model.ProtoTalker) {
	t.Helper()
	for _, p := range protoTop {
		switch p.Proto {
		case protoTCP:
			if p.Bytes != 1000 {
				t.Errorf("expected TCP=1000, got %d", p.Bytes)
			}
		case protoUDP:
			if p.Bytes != 500 {
				t.Errorf("expected UDP=500, got %d", p.Bytes)
			}
		default:
			t.Errorf("unexpected proto %s", p.Proto)
		}
	}
}

func checkFlowTop(t *testing.T, flowTop []model.FlowTalker) {
	t.Helper()
	foundFlow1, foundFlow2 := false, false
	for _, f := range flowTop {
		switch f.Src {
		case "1.2.3.4:1234":
			foundFlow1 = true
			if f.BPS != 1000 || f.Proto != protoTCP {
				t.Errorf("unexpected flow %+v", f)
			}
		case "10.0.0.1:1111":
			foundFlow2 = true
			if f.BPS != 500 || f.Proto != protoUDP {
				t.Errorf("unexpected flow %+v", f)
			}
		}
	}
	if !foundFlow1 || !foundFlow2 {
		t.Errorf("missing expected flows: %+v", flowTop)
	}
}
