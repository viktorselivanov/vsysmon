package report

import (
	"fmt"
	"log"
	"net"
	"sync"

	pb "vsysmon/proto"

	"google.golang.org/grpc"
)

type grpcServer struct {
	pb.UnimplementedStatsServiceServer
	mu      sync.Mutex                                     // защищает доступ
	clients map[pb.StatsService_StreamStatsServer]struct{} // активные подписчики
}

var grpcOut = make(chan *pb.Snapshot, 16) // буферизированный канал

func StartGRPC(port int) {
	sp := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", sp) // создаём TCP-листенер передаём сформированный порт без указания ip
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", sp, err)
	}

	s := &grpcServer{
		clients: make(map[pb.StatsService_StreamStatsServer]struct{}), // создаём сервер и инициализируем мапу подписчиков
	}

	g := grpc.NewServer()               // создаём gRPC-сервер
	pb.RegisterStatsServiceServer(g, s) // регистрируем

	go broadcaster(s) // броадкастер в отдельной го рутине

	if err := g.Serve(lis); err != nil { // запускаем gRPC,
		fmt.Printf("grpc serve stopped: %v", err)
	}
}

func (s *grpcServer) StreamStats(_ *pb.Empty, stream pb.StatsService_StreamStatsServer) error { // поток метрик
	s.mu.Lock()
	s.clients[stream] = struct{}{} //  не потокобезопасна по этому оборачивам в мьютекс
	s.mu.Unlock()

	<-stream.Context().Done() // ждём, пока клиент отключится

	s.mu.Lock()
	delete(s.clients, stream) // удаляем поток из мапы
	s.mu.Unlock()

	return nil
}

func broadcaster(s *grpcServer) { // рассылает snapshot всем подписчикам
	for snap := range grpcOut {
		s.mu.Lock()
		for c := range s.clients {
			if err := c.Send(snap); err != nil { // отправляем данные подписчику
				// клиент отключился, удаляем из списка
				delete(s.clients, c)
			}
		}
		s.mu.Unlock()
	}
}
