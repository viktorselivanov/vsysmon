package report

import (
	"sort"
	"strings"
	"time"
	"vsysmon/config"
	model "vsysmon/model"
	pb "vsysmon/proto"
	"vsysmon/ring"
	"vsysmon/terminal"
)

func Reporter(cfg config.Config, verbose bool, N int) {
	t := time.NewTicker(time.Duration(N) * time.Second) //таймер, тикер срабатывает раз в N секунд
	for range t.C {
		samples := ring.RingSnapshot() // возвращает копию всех накопленных Sample, если пусто пропускаем
		if len(samples) == 0 {
			continue
		}

		snap := aggregate(samples) // аггрегируем, формируем снепшот
		ring.SaveSnapshot(snap)    // безопасное сохранение
		if verbose {               // включаем запись в консоль мониторинга на стороне сервера только при отладке с флагом -v

			// используем единый вывод с клиентом и при этом можем выключать вывод отдельных функций в терминал
			terminal.Render(&snap, terminal.BuildSections(cfg)) // отправляем снепшот для обработки, так же формируем пайплайн через cfg для включения/выключения функций
		}
		grpcOut <- snapshotToProto(ring.LastSnapshot()) // безопасное сохранение
	}
}

// aggregate усредняет данные за последние M секунд
func aggregate(samples []model.Sample) model.Snapshot {
	var s model.Snapshot
	s.TCPStates = make(map[string]int)

	protoMap := make(map[string]uint64) // суммируем по протоколам
	flowMap := make(map[string]uint64)  // суммируем по flow "src->dst|proto"

	for _, x := range samples {
		s.Load += x.Load
		s.CPUUser += x.CPUUser
		s.CPUSys += x.CPUSys
		s.CPUIdle += x.CPUIdle
		s.DiskTPS += x.DiskTPS
		s.DiskKBs += x.DiskKBs

		for k, v := range x.TCPStates {
			s.TCPStates[k] += v
		}
		// объединяем FS, оставляем последние значения
		s.FS = x.FS

		// ProtoTop суммируем
		for _, p := range x.ProtoTop {
			protoMap[p.Proto] += p.Bytes
		}

		// FlowTop суммируем и формируем будущий вид
		for _, f := range x.FlowTop {
			key := f.Src + "->" + f.Dst + "|" + f.Proto
			flowMap[key] += f.BPS
		}
	}

	// усреднение
	n := float64(len(samples))
	s.Load /= n
	s.CPUUser /= n
	s.CPUSys /= n
	s.CPUIdle /= n
	s.DiskTPS /= n
	s.DiskKBs /= n

	// формируем ProtoTop с процентами

	totalBytes := uint64(0)
	for _, v := range protoMap {
		totalBytes += v
	}

	for proto, b := range protoMap {
		perc := 0.0
		if totalBytes > 0 {
			perc = float64(b) / float64(totalBytes) * 100
		}
		s.ProtoTop = append(s.ProtoTop, model.ProtoTalker{
			Proto: proto,
			Bytes: b,
			Perc:  perc,
		})
	}

	sort.Slice(s.ProtoTop, func(i, j int) bool {
		return s.ProtoTop[i].Perc > s.ProtoTop[j].Perc // по убыванию %
	})

	// формируем FlowTop
	for k, b := range flowMap {
		parts := strings.Split(k, "|")
		addr := strings.Split(parts[0], "->")
		s.FlowTop = append(s.FlowTop, model.FlowTalker{
			Src:   addr[0],
			Dst:   addr[1],
			Proto: parts[1],
			BPS:   b,
		})
	}

	sort.Slice(s.FlowTop, func(i, j int) bool {
		return s.FlowTop[i].BPS > s.FlowTop[j].BPS // по убыванию BPS
	})

	if len(s.FlowTop) > 10 {
		s.FlowTop = s.FlowTop[:10] // ограничение на вывод не более 10ти
	}

	last := samples[len(samples)-1]
	s.Listen = last.Listen // берём только из последнего (так как не часто изменяется информация)

	return s
}

// snapshotToProto преобразует Snapshot в protobuf объект
func snapshotToProto(s model.Snapshot) *pb.Snapshot {
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
	tcpMap := make(map[string]int32)
	for k, v := range s.TCPStates {
		tcpMap[k] = int32(v)
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
