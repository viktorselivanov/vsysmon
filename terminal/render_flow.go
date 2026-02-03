package terminal

import (
	"fmt"
	"vsysmon/model"
)

type RenderFlowTop struct{}

func (r *RenderFlowTop) Name() string { return "flotop" }

func (r *RenderFlowTop) Render(s *model.Snapshot) {
	fmt.Println("TOP TALKERS — BY FLOW (BPS)")
	fmt.Println("---------------------------------------------------------------------")
	fmt.Printf("%-22s -> %-22s %-6s %12s\n", "SRC", "DST", "PROTO", "BPS")
	for i, f := range s.FlowTop {
		if i >= 10 {
			break
		}
		fmt.Printf("%-22.22s -> %-22.22s %-6s %10d\n",
			f.Src, f.Dst, f.Proto, f.BPS)
	}
	fmt.Println()
}
