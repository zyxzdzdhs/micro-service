package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "ride-sharing/shared/proto/driver"
)

type grpcHandler struct {
	pb.UnimplementedDriverServiceServer
	Service *Service
}

func NewGrpcHandler(server *grpc.Server, service *Service) *grpcHandler {
	handler := &grpcHandler{
		Service: service,
	}

	pb.RegisterDriverServiceServer(server, handler)
	return handler
}

func (h *grpcHandler) RegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method RegisterDriver not implemented")
}
func (h *grpcHandler) UnregisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UnregisterDriver not implemented")
}
