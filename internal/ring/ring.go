package ring

import (
	"sync"

	model "vsysmon/internal/model"
)

var (
	ring   []model.Sample // кольцевой буфер для хранения последних M сэмплов
	idx    int            // индекс, куда класть следующий сэмпл
	filled bool           // флаг, что кольцо уже полностью заполнено хотя бы один раз
	rmu    sync.RWMutex   // RWMutex для безопасного доступа к кольцу из нескольких горутин

	last model.Snapshot // последний агрегированный снимок
	lmu  sync.RWMutex   // RWMutex для безопасного доступа к last
)

func Init(m int) {
	ring = make([]model.Sample, m)
}

func Push(s *model.Sample) {
	rmu.Lock()                  // блокируем доступ на запись
	ring[idx] = *s              // пишем сэмпл в текущую позицию
	idx = (idx + 1) % len(ring) // двигаем индекс по модулю длины кольца
	if idx == 0 {               // если мы прошли полный круг, значит кольцо заполнено
		filled = true
	}
	rmu.Unlock() // разблокируем
}

func Snapshot() []model.Sample {
	rmu.RLock() // блокировка для чтения
	// RLock() позволяет читать кольцо одновременно с другими чтениями, но блокирует запись.
	defer rmu.RUnlock()

	if !filled { // если кольцо ещё не заполнено полностью, данных недостаточно
		return nil
	}
	out := make([]model.Sample, len(ring))
	copy(out, ring) // создаём копию, чтобы не дать читателю менять оригинал
	return out
}

// Эта функция нужна для того, чтобы клиенты gRPC могли безопасно получить последний snapshot, не опасаясь гонок.
func SaveSnapshot(s *model.Snapshot) {
	lmu.Lock()   // блокируем доступ на запись
	last = *s    // сохраняем snapshot
	lmu.Unlock() // разблокируем
}

func LastSnapshot() model.Snapshot {
	lmu.RLock()         // блокируем доступ на чтение
	defer lmu.RUnlock() // Отложенное снятие блокировки после выхода из функции
	return last         // Возвращаем последний snapshot
}
