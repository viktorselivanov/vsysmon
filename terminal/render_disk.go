package terminal

import (
	"fmt"
	"vsysmon/model"
)

type RenderDisk struct{}

func (r *RenderDisk) Name() string { return "disk" }

func (r *RenderDisk) Render(s *model.Snapshot) {
	fmt.Println("DISK IO")
	fmt.Println("-----------------------------")
	fmt.Printf("%-10s %10.2f\n", "TPS", s.DiskTPS)
	fmt.Printf("%-10s %10.2f\n\n", "KB/s", s.DiskKBs)
}
