package main

import (
	"context"
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
		s.logError("gRPC listen failed on " + s.grpcAddress + ": " + err.Error())
		return
	}
	defer lis.Close()

	server := grpc.NewServer()
	akipproto.RegisterAkipServer(server, &akipGRPCServer{service: s})

	go func() {
		<-ctx.Done()
		server.Stop()
	}()

	s.logInfo("gRPC server is running on " + s.grpcAddress)
	if err := server.Serve(lis); err != nil {
		s.logWarn("gRPC server stopped: " + err.Error())
	}
}
