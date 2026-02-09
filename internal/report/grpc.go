package report

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
	"vsysmon/internal/model"
	"vsysmon/internal/ring"

	pb "vsysmon/proto"

	"google.golang.org/grpc"
)

type grpcServer struct {
	pb.UnimplementedStatsServiceServer
	mu      sync.Mutex                                    // защищает доступ
	clients map[pb.StatsService_StreamStatsServer]*client // активные подписчики
}

type client struct {
	stream pb.StatsService_StreamStatsServer
	agg    *ring.Aggregator
	n      int // интервал отправки snapshot клиенту (секунды)
	m      int // окно агрегации / сколько последних секунд усреднять
}

func StartGRPC(port int, in <-chan *model.Sample) {
	sp := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", sp) // создаём TCP-листенер передаём сформированный порт без указания ip
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", sp, err)
	}

	s := &grpcServer{
		clients: make(map[pb.StatsService_StreamStatsServer]*client), // создаём сервер и инициализируем мапу подписчиков
	}

	g := grpc.NewServer()               // создаём gRPC-сервер
	pb.RegisterStatsServiceServer(g, s) // регистрируем

	go broadcaster(in, s) // броадкастер в отдельной го рутине

	if err := g.Serve(lis); err != nil { // запускаем gRPC,
		fmt.Printf("grpc serve stopped: %v", err)
	}
}

func (s *grpcServer) StreamStats(req *pb.StreamRequest, stream pb.StatsService_StreamStatsServer) error { // поток метрик

	n := int(req.N)
	if req.N > uint64(^uint(0)>>1) { // максимум int на текущей платформе
		n = int(^uint(0) >> 1)
	}

	m := int(req.M)
	if req.M > uint64(^uint(0)>>1) {
		m = int(^uint(0) >> 1)
	}

	c := &client{
		stream: stream,
		agg:    ring.NewAggregator(int(req.M)),
		n:      n,
		m:      m,
	}

	// дефолты
	if c.n <= 0 {
		c.n = 15
	}
	if c.m <= 0 {
		c.m = 5
	}

	s.mu.Lock()
	s.clients[stream] = c //  не потокобезопасна по этому оборачивам в мьютекс
	s.mu.Unlock()

	go func(c *client) {
		ticker := time.NewTicker(time.Duration(c.n) * time.Second)
		defer ticker.Stop()

		for range ticker.C {

			clientSnap := c.agg.Aggregate() // агрегируем по окну m
			if clientSnap != nil {
				if err := c.stream.Send(snapshotToProto(clientSnap)); err != nil {
					// клиент отключился
					s.mu.Lock()
					delete(s.clients, c.stream)
					s.mu.Unlock()
					return
				}
			}
		}
	}(c)

	<-stream.Context().Done() // ждём, пока клиент отключится

	s.mu.Lock()
	delete(s.clients, stream) // удаляем поток из мапы
	s.mu.Unlock()

	return nil
}

func broadcaster(samples <-chan *model.Sample, s *grpcServer) { // рассылает snapshot всем подписчикам
	for sample := range samples {
		s.mu.Lock()
		for _, c := range s.clients {
			c.agg.Push(sample) // обновляем агрегатор клиента
		}
		s.mu.Unlock()
	}
}
