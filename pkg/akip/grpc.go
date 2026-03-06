package akip

import (
	"context"
	"log"
	"net"

	"github.com/pioruner/HardWorker.git/pkg/app"
	"github.com/pioruner/HardWorker.git/pkg/proto"
	"google.golang.org/grpc"
)

// Реализация сервиса
type myAkipServer struct {
	proto.UnimplementedAkipServer
}

func (s *myAkipServer) Data(ctx context.Context, req *proto.AkipRequest) (*proto.AkipResponse, error) {
	return &proto.AkipResponse{Level: -1.00}, nil
}

func (ak *AkipUI) gRPC() {
	lis, err := net.Listen("tcp", ak.gport)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}
	defer lis.Close()
	grpcServer := grpc.NewServer()
	proto.RegisterAkipServer(grpcServer, &myAkipServer{})
	go func() {
		<-app.Ctx.Done()  // Ожидание отмены контекста
		grpcServer.Stop() // Остановка gRPC сервера
		log.Println("gRPC server stopped")
	}()

	log.Printf("gRPC server is running on :%s", ak.gport)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
