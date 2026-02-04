package terminal

import (
	"fmt"

	model "vsysmon/internal/model"
)

type RenderFS struct{}

func (r *RenderFS) Name() string { return "fs" }

func (r *RenderFS) Render(s *model.Snapshot) {
	fmt.Println("FILESYSTEMS")
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("%-30s %-15s %10s %8s %10s %8s\n",
		"FS", "MOUNT", "USED(MB)", "%", "INODE", "%")

	for _, fs := range s.FS {
		fmt.Printf("%-30s %-15s %10d %8.1f %10d %8.1f\n",
			fs.Filesystem,
			fs.MountPoint,
			fs.UsedMB,
			fs.UsedPerc,
			fs.UsedInode,
			fs.InodePerc,
		)
	}
	fmt.Println()
}
