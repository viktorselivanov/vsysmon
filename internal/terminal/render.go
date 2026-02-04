package terminal

import (
	"fmt"

	model "vsysmon/internal/model"
)

const (
	rendererLoad   = "load"
	rendererFS     = "fs"
	rendererListen = "listen"
	rendererCPU    = "cpu"
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
