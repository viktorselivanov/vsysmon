package terminal

import (
	"fmt"

	model "vsysmon/internal/model"
)

type RenderCPU struct{}

func (r *RenderCPU) Name() string { return "cpu" }

func (r *RenderCPU) Render(s *model.Snapshot) {
	fmt.Println("CPU USAGE (%)")
	fmt.Println("-----------------------------")
	fmt.Printf("%-10s %8.2f\n", "User", s.CPUUser)
	fmt.Printf("%-10s %8.2f\n", "System", s.CPUSys)
	fmt.Printf("%-10s %8.2f\n\n", "Idle", s.CPUIdle)
}
