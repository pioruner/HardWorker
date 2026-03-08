package main

import (
	"context"
	"log"
	"net"

	akipproto "akip-wails-prototype/proto"
	"google.golang.org/grpc"
)

type akipGRPCServer struct {
	akipproto.UnimplementedAkipServer
	service *AkipService
}

func (s *akipGRPCServer) Data(ctx context.Context, req *akipproto.AkipRequest) (*akipproto.AkipResponse, error) {
	return &akipproto.AkipResponse{Level: s.service.volumeLevel()}, nil
}

func (s *AkipService) grpcLoop(ctx context.Context) {
	lis, err := net.Listen("tcp", s.grpcAddress)
	if err != nil {
		log.Printf("gRPC listen failed on %s: %v", s.grpcAddress, err)
		return
	}
	defer lis.Close()

	server := grpc.NewServer()
	akipproto.RegisterAkipServer(server, &akipGRPCServer{service: s})

	go func() {
		<-ctx.Done()
		server.Stop()
	}()

	log.Printf("gRPC server is running on %s", s.grpcAddress)
	if err := server.Serve(lis); err != nil {
		log.Printf("gRPC server stopped: %v", err)
	}
}
