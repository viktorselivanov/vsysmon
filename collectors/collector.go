package collectors

import (
	"time"
	model "vsysmon/model"
	"vsysmon/ring"
)

type MetricCollector interface {
	Collect(*model.Sample)
}

type In = <-chan *model.Sample // только чтение
type Out = In
type Bi = chan *model.Sample //двунаправленный

type Stage func(in In, done <-chan struct{}) Out // читает из входного канала слушает done возвращает выходной канал

// ExecutePipeline соединяет несколько Stage
func ExecutePipeline(in In, done <-chan struct{}, stages ...Stage) Out {
	out := in
	for _, stage := range stages {
		out = stage(out, done)
	}
	return out
}

// Stager просто пропускает значения, проверяя done
func Stager(in In, done <-chan struct{}) Out {
	out := make(Bi)
	go func() {
		defer close(out)
		for {
			select {
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-done:
					for range in {
						_ = struct{}{} //корректно закрываем все стадии
					}
					return
				}
			case <-done:
				for range in {
					_ = struct{}{} // тоже самое
				}
				return
			}
		}
	}()
	return out
}

// CollectorStage создаёт Stage из конкретного MetricCollector
func CollectorStage(c MetricCollector) Stage {
	return func(in In, done <-chan struct{}) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for s := range in {
				select {
				case <-done:
					return
				default:
					c.Collect(s) // запись в сэмпл
					out <- s
				}
			}
		}()
		return out
	}
}

// collectorTicker создаёт сэмпл и отправляет в пайплайн кажду секунду
func collectorTicker(done <-chan struct{}) In {
	out := make(Bi)
	go func() {
		defer close(out)
		t := time.NewTicker(time.Second)
		defer t.Stop()

		for {
			select {
			case <-done:
				return
			case <-t.C:
				s := &model.Sample{}
				out <- s
			}
		}
	}()
	return out
}

// StartCollector запускает весь пайплайн
func StartCollector(done <-chan struct{}, collectors []MetricCollector) {
	stages := make([]Stage, 0, len(collectors))
	for _, c := range collectors {
		stages = append(stages, CollectorStage(c))
	}

	src := collectorTicker(done)
	out := ExecutePipeline(src, done, stages...)

	// читаем сэплы записываем в кольцевой буфер
	go func() {
		for s := range out {
			ring.RingPush(*s)
		}
	}()
}
