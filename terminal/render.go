package terminal

import (
	"fmt"
	"vsysmon/model"
)

// интерфейс для секцподсистемы вывода метрик.
type MetricRenderer interface {
	Render(*model.Snapshot)
	Name() string
}

func Render(s *model.Snapshot, sections []MetricRenderer) {
	Clear()
	fmt.Println("======== SYSTEM SNAPSHOT ========")
	fmt.Println()

	// Основной цикл рендеринга
	for _, r := range sections {
		r.Render(s)
		fmt.Println()
	}
}
