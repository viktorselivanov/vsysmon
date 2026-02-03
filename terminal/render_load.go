package terminal

import (
	"fmt"
	"vsysmon/model"
)

type LoadRenderer struct{}

func (r *LoadRenderer) Name() string { return "load" }

func (r *LoadRenderer) Render(s *model.Snapshot) {
	fmt.Println("LOAD AVERAGE")
	fmt.Println("-----------------------------")
	fmt.Printf("Load avg: %.2f\n", s.Load)
}
