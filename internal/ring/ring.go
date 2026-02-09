package ring

import (
	"sort"
	"strings"
	"sync"

	model "vsysmon/internal/model"
)

type Aggregator struct {
	mu      sync.RWMutex
	samples []*model.Sample
	m       int
}

type Ring struct {
	mu     sync.RWMutex
	buf    []model.Sample
	filled bool
}

func New(size int) *Ring {
	return &Ring{
		buf: make([]model.Sample, size),
	}
}

func NewAggregator(m int) *Aggregator {
	return &Aggregator{
		samples: make([]*model.Sample, 0, m),
		m:       m,
	}
}

func (a *Aggregator) Push(s *model.Sample) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.samples) >= a.m {
		a.samples = a.samples[1:] // удаляем старый
	}
	a.samples = append(a.samples, s)
}

func (r *Ring) Snapshot() []model.Sample {
	r.mu.RLock() // блокировка для чтения
	// RLock() позволяет читать кольцо одновременно с другими чтениями, но блокирует запись.
	defer r.mu.RUnlock()

	if !r.filled { // если кольцо ещё не заполнено полностью, данных недостаточно
		return nil
	}

	out := make([]model.Sample, len(r.buf))
	copy(out, r.buf) // создаём копию, чтобы не дать читателю менять оригинал

	return out
}

// aggregate усредняет данные за последние M секунд.
func (a *Aggregator) Aggregate() *model.Snapshot {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if len(a.samples) == 0 {
		return nil
	}

	var s model.Snapshot
	s.TCPStates = make(map[string]int)

	protoMap := make(map[string]uint64) // суммируем по протоколам
	flowMap := make(map[string]uint64)  // суммируем по flow "src->dst|proto"

	for i := range a.samples {
		x := a.samples[i]
		s.Load += x.Load
		s.CPUUser += x.CPUUser
		s.CPUSys += x.CPUSys
		s.CPUIdle += x.CPUIdle
		s.DiskTPS += x.DiskTPS
		s.DiskKBs += x.DiskKBs

		for k, v := range x.TCPStates {
			s.TCPStates[k] += v
		}
		// объединяем FS, оставляем последние значения
		s.FS = x.FS

		// ProtoTop суммируем
		for _, p := range x.ProtoTop {
			protoMap[p.Proto] += p.Bytes
		}

		// FlowTop суммируем и формируем будущий вид
		for _, f := range x.FlowTop {
			key := f.Src + "->" + f.Dst + "|" + f.Proto
			flowMap[key] += f.BPS
		}
	}

	// усреднение
	n := float64(len(a.samples))
	s.Load /= n
	s.CPUUser /= n
	s.CPUSys /= n
	s.CPUIdle /= n
	s.DiskTPS /= n
	s.DiskKBs /= n

	// формируем ProtoTop с процентами

	totalBytes := uint64(0)
	for _, v := range protoMap {
		totalBytes += v
	}

	for proto, b := range protoMap {
		perc := 0.0
		if totalBytes > 0 {
			perc = float64(b) / float64(totalBytes) * 100
		}
		s.ProtoTop = append(s.ProtoTop, model.ProtoTalker{
			Proto: proto,
			Bytes: b,
			Perc:  perc,
		})
	}

	sort.Slice(s.ProtoTop, func(i, j int) bool {
		return s.ProtoTop[i].Perc > s.ProtoTop[j].Perc // по убыванию %
	})

	// формируем FlowTop
	for k, b := range flowMap {
		parts := strings.Split(k, "|")
		addr := strings.Split(parts[0], "->")
		s.FlowTop = append(s.FlowTop, model.FlowTalker{
			Src:   addr[0],
			Dst:   addr[1],
			Proto: parts[1],
			BPS:   b,
		})
	}

	sort.Slice(s.FlowTop, func(i, j int) bool {
		return s.FlowTop[i].BPS > s.FlowTop[j].BPS // по убыванию BPS
	})

	if len(s.FlowTop) > 10 {
		s.FlowTop = s.FlowTop[:10] // ограничение на вывод не более 10ти
	}

	last := a.samples[len(a.samples)-1]
	s.Listen = last.Listen // берём только из последнего (так как не часто изменяется информация)

	return &s
}
