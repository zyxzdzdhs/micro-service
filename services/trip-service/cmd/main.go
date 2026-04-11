package main

import (
	"context"
	grpcserver "google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/trip-service/internal/domain"
	h "ride-sharing/services/trip-service/internal/infrastructure/http"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"syscall"
	"time"
)

var GrpcAddr = ":9093"

func main() {
	memRepo := repository.NewInmemRepository()
	svc := service.NewService(memRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建携程等待系统调用的终止信号，转发到sigCh中并调用取消信号
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatal("failed to listen: %v", err)
	}

	grpcServer := grpcserver.NewServer()

	log.Printf("Starting grpc server Trip service on port %s", lis.Addr().String())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to server: %v", err)
			cancel()
		}
	}()

	// 当调用取消后，GRPC服务器会优雅关闭
	<-ctx.Done()
	log.Printf("Shutting down the server...")
	grpcServer.GracefulStop()
}
