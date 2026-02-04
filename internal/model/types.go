package model

type ListenSocket struct {
	Command  string
	PID      uint32
	User     string
	Protocol string
	Port     uint64
}

type ProtoTalker struct {
	Proto string
	Bytes uint64
	Perc  float64
}

type FlowTalker struct {
	Src   string
	Dst   string
	Proto string
	BPS   uint64
}

type FSStat struct {
	Filesystem string
	MountPoint string
	UsedMB     uint64
	UsedPerc   float64
	UsedInode  uint64
	InodePerc  float64
}

type Sample struct {
	Load float64

	CPUUser float64
	CPUSys  float64
	CPUIdle float64

	DiskTPS float64
	DiskKBs float64

	TCPStates map[string]int

	FS []FSStat

	ProtoTop []ProtoTalker
	FlowTop  []FlowTalker

	Listen []ListenSocket
}

type Snapshot struct {
	Load float64

	CPUUser float64
	CPUSys  float64
	CPUIdle float64

	DiskTPS float64
	DiskKBs float64

	TCPStates map[string]int

	FS []FSStat

	ProtoTop []ProtoTalker
	FlowTop  []FlowTalker

	Listen []ListenSocket
}
