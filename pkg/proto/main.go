package proto

import (
	"context"
	"log"
	"net"

	"github.com/pioruner/HardWorker.git/pkg/app"
	"google.golang.org/grpc"
)

// Реализация сервиса
type myAkipServer struct {
	UnimplementedAkipServer
}

func (s *myAkipServer) Data(ctx context.Context, req *AkipRequest) (*AkipResponse, error) {
	return &AkipResponse{Level: -1.00}, nil
}

func Grpc() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}
	defer lis.Close()
	grpcServer := grpc.NewServer()
	RegisterAkipServer(grpcServer, &myAkipServer{})
	go func() {
		<-app.Ctx.Done()  // Ожидание отмены контекста
		grpcServer.Stop() // Остановка gRPC сервера
		log.Println("gRPC server stopped")
	}()

	log.Println("gRPC server is running on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
