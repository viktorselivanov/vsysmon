package terminal

import (
	"fmt"

	model "vsysmon/internal/model"
)

type RenderTCP struct{}

func (r *RenderTCP) Name() string { return "load" }

func (r *RenderTCP) Render(s *model.Snapshot) {
	fmt.Println("TCP STATES")
	fmt.Println("-----------------------------")
	fmt.Printf("%-15s %8s\n", "STATE", "COUNT")

	for k, v := range s.TCPStates {
		fmt.Printf("  %s: %d\n", k, v)
	}
	fmt.Println()
}
