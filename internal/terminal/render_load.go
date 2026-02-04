package terminal

import (
	"fmt"

	model "vsysmon/internal/model"
)

type LoadRenderer struct{}

func (r *LoadRenderer) Name() string { return rendererLoad }

func (r *LoadRenderer) Render(s *model.Snapshot) {
	fmt.Println("LOAD AVERAGE")
	fmt.Println("-----------------------------")
	fmt.Printf("Load avg: %.2f\n", s.Load)
}
