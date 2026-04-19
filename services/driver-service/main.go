package main

import (
	"context"
	grpcserver "google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"syscall"
)

var GrpcAddr = ":9092"

func main() {
	service := NewService()
	rabbitMqUri := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

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
		log.Fatalf("failed to listen: %v", err)
	}

	// RabbitMQ connection
	rmq, err := messaging.NewRabbitMQ(rabbitMqUri)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer rmq.Close()

	// 开始GRPC SERVER
	grpcServer := grpcserver.NewServer()
	NewGrpcHandler(grpcServer, service)

	log.Printf("Starting grpc server Driver service on port %s", lis.Addr().String())

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
