package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"vsysmon/internal/config"
	"vsysmon/internal/model"
	"vsysmon/internal/terminal"
	pb "vsysmon/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func RunClient() error {
	sp := fmt.Sprintf("localhost:%d", *port)

	conn, err := grpc.NewClient( // открываем соединение к gRPC-серверу
		sp,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // создаём обычное TCP соединение без TLS
	)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewStatsServiceClient(conn) // создаём gRPC-клиент

	stream, err := client.StreamStats(context.Background(), &pb.Empty{}) // получаем поток данных
	if err != nil {
		return fmt.Errorf("failed to start stream: %w", err)
	}
	defer func() {
		if err := stream.CloseSend(); err != nil {
			// Можно логировать, если это критично
			fmt.Printf("failed to close gRPC send stream: %v\n", err)
		}
	}()

	sig := make(chan os.Signal, 1) // обрабатываем Ctrl+C при нажатии выдаём сообщение и завершаем процесс
	signal.Notify(sig, os.Interrupt)

	go func() {
		<-sig
		fmt.Println("\nbye")
		os.Exit(0)
	}()

	for { // бесконечный цикл чтения сообщений из стрима
		s, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}
		// используем единый вывод с сервером,
		//  для этого конвертируем прото в снепшот и даём дефолтную конфигурацию что бы не тащить за собою конфиг
		// не по grpc и не файлом.
		//  в случае отключения функций на сервере -> в клиенте будут висеть нулевые значения в выключенных функциях.
		terminal.Render(FromProto(s), terminal.BuildSections(config.DefaultConfig()))
	}
}

// конвертируем протобуф в снепшот.
func FromProto(p *pb.Snapshot) *model.Snapshot {
	var s model.Snapshot

	// Простые показатели без ковертации
	s.Load = p.Load
	s.CPUUser = p.CpuUser
	s.CPUSys = p.CpuSys
	s.CPUIdle = p.CpuIdle
	s.DiskTPS = p.DiskTps
	s.DiskKBs = p.DiskKbs

	// Файловые системы
	s.FS = make([]model.FSStat, len(p.Fs))
	for i, f := range p.Fs {
		s.FS[i] = model.FSStat{
			Filesystem: f.Filesystem,
			MountPoint: f.MountPoint,
			UsedMB:     f.UsedMb,
			UsedPerc:   f.UsedPerc,
			UsedInode:  f.UsedInode,
			InodePerc:  f.InodePerc,
		}
	}

	// TCP состояния
	s.TCPStates = make(map[string]int)
	for k, v := range p.TcpStates {
		s.TCPStates[k] = int(v)
	}

	// Top Talkers — протоколы
	s.ProtoTop = make([]model.ProtoTalker, len(p.ProtoTop))
	for i, t := range p.ProtoTop {
		s.ProtoTop[i] = model.ProtoTalker{
			Proto: t.Proto,
			Bytes: t.Bytes,
			Perc:  t.Perc,
		}
	}

	// Top Talkers — потоки
	s.FlowTop = make([]model.FlowTalker, len(p.FlowTop))
	for i, f := range p.FlowTop {
		s.FlowTop[i] = model.FlowTalker{
			Src:   f.Src,
			Dst:   f.Dst,
			Proto: f.Proto,
			BPS:   f.Bps,
		}
	}

	// Listening sockets
	s.Listen = make([]model.ListenSocket, len(p.Listen))
	for i, l := range p.Listen {
		s.Listen[i] = model.ListenSocket{
			Protocol: l.Protocol,
			Port:     l.Port,
			User:     l.User,
			PID:      l.Pid,
			Command:  l.Command,
		}
	}

	return &s
}
