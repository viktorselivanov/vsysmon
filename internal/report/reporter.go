package report

import (
	"time"
	"vsysmon/internal/config"
	"vsysmon/internal/model"
	"vsysmon/internal/ring"
	"vsysmon/internal/terminal"
	pb "vsysmon/proto"
)

func Reporter(cfg config.Config, verbose bool, n, m int, in <-chan *model.Sample) {
	if !verbose {
		return
	}

	ticker := time.NewTicker(time.Duration(n) * time.Second)
	defer ticker.Stop()

	agg := ring.NewAggregator(m)

	for {
		select {
		case sample := <-in:
			if sample == nil {
				continue
			}
			agg.Push(sample) // кладём сэмпл в агрегатор

		case <-ticker.C:
			snap := agg.Aggregate()
			if snap != nil {
				terminal.Render(snap, terminal.BuildSections(cfg))
			}
		}
	}
}

// snapshotToProto преобразует Snapshot в protobuf объект.
func snapshotToProto(s *model.Snapshot) *pb.Snapshot {
	// Файловые системы
	fsProto := make([]*pb.FSStat, 0, len(s.FS))
	for _, fs := range s.FS {
		fsProto = append(fsProto, &pb.FSStat{
			Filesystem: fs.Filesystem,
			MountPoint: fs.MountPoint,
			UsedMb:     fs.UsedMB,
			UsedPerc:   fs.UsedPerc,
			UsedInode:  fs.UsedInode,
			InodePerc:  fs.InodePerc,
		})
	}

	// TCP состояния
	tcpMap := make(map[string]int64)
	for k, v := range s.TCPStates {
		tcpMap[k] = int64(v)
	}

	// Top Talkers — протоколы
	protoTop := make([]*pb.ProtoTalker, 0, len(s.ProtoTop))
	for _, p := range s.ProtoTop {
		protoTop = append(protoTop, &pb.ProtoTalker{
			Proto: p.Proto,
			Bytes: p.Bytes,
			Perc:  p.Perc,
		})
	}

	// Top Talkers — потоки
	flowTop := make([]*pb.FlowTalker, 0, len(s.FlowTop))
	for _, f := range s.FlowTop {
		flowTop = append(flowTop, &pb.FlowTalker{
			Src:   f.Src,
			Dst:   f.Dst,
			Proto: f.Proto,
			Bps:   f.BPS,
		})
	}

	// Listening sockets
	listenProto := make([]*pb.ListenSocket, 0, len(s.Listen))
	for _, l := range s.Listen {
		listenProto = append(listenProto, &pb.ListenSocket{
			Protocol: l.Protocol,
			Port:     l.Port,
			User:     l.User,
			Pid:      l.PID,
			Command:  l.Command,
		})
	}
	// Простые показатели без ковертации так же включены
	return &pb.Snapshot{
		Load:      s.Load,
		CpuUser:   s.CPUUser,
		CpuSys:    s.CPUSys,
		CpuIdle:   s.CPUIdle,
		DiskTps:   s.DiskTPS,
		DiskKbs:   s.DiskKBs,
		TcpStates: tcpMap,
		Fs:        fsProto,
		ProtoTop:  protoTop,
		FlowTop:   flowTop,
		Listen:    listenProto,
	}
}
