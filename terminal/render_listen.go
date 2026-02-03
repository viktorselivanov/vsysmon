package terminal

import (
	"fmt"
	"vsysmon/model"
)

type RenderListen struct{}

func (r *RenderListen) Name() string { return "listen" }

func (r *RenderListen) Render(s *model.Snapshot) {
	fmt.Println("LISTENING SOCKETS")
	fmt.Println("-----------------------------------------------------------")
	fmt.Printf("%-6s %-6s %-6s %-10s %s\n", "P", "PORT", "PID", "USER", "CMD")

	for _, l := range s.Listen {
		fmt.Printf("%-6s %-6d %-6d %-10.10s %s\n",
			l.Protocol, l.Port, l.PID, l.User, l.Command)
	}

	fmt.Println()
}
