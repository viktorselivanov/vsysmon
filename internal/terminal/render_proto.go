package terminal

import (
	"fmt"

	model "vsysmon/internal/model"
)

type RenderProtoTop struct{}

func (r *RenderProtoTop) Name() string { return "proto" }

func (r *RenderProtoTop) Render(s *model.Snapshot) {
	fmt.Println("TOP TALKERS — BY PROTOCOL")
	fmt.Println("--------------------------------------------")
	fmt.Printf("%-8s %12s %8s\n", "PROTO", "BYTES/s", "%")

	for _, p := range s.ProtoTop {
		fmt.Printf("%-8s %12d %7.1f%%\n", p.Proto, p.Bytes, p.Perc)
	}
	fmt.Println()
}
