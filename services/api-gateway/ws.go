package main

import (
	"encoding/json"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/proto/driver"
)

// 创建一个全局的连接manager
var (
	connManager = messaging.NewConnectionManager()
)

func handleRiderWebSocket(w http.ResponseWriter, r *http.Request, rmq *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	defer conn.Close()

	// 获取用户ID
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Printf("The user id is empty")
		return
	}

	// 用户ID获取到之后就将该用户的连接加入到Manager的连接池子里面
	connManager.Add(userID, conn)
	defer connManager.Remove(userID)

	// 初始化queue consumers
	queues := []string{
		messaging.NotifyDriverNotFoundQueue,
		messaging.NotifyDriverAssignedQueue,
	}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rmq, connManager, q)

		if err := consumer.Start(); err != nil {
			log.Printf("failed to start consumer for queue: %s: err: %v", q, err)
		}
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		log.Printf("Receive message: %s", message)
	}

}

func handleDriverWebSocket(w http.ResponseWriter, r *http.Request, rmq *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	defer conn.Close()

	// 获取用户ID
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Printf("The user id is empty")
		return
	}

	// 获取司机注册的车辆类型
	packageSlug := r.URL.Query().Get("packageSlug")
	if packageSlug == "" {
		log.Printf("The package slug is empty")
		return
	}

	// 用户ID获取到之后就将该用户的连接加入到Manager的连接池子里面
	connManager.Add(userID, conn)
	connManager.Remove(userID)

	ctx := r.Context()

	driverService, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	// Closing connections
	defer func() {
		driverService.Client.UnregisterDriver(ctx, &driver.RegisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		})

		driverService.Close()

		log.Println("Driver unregistered: ", userID)
	}()

	driverData, err := driverService.Client.RegisterDriver(ctx, &driver.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	})
	if err != nil {
		log.Printf("Error registering driver: %v", err)
		return
	}

	msg := contracts.WSMessage{
		Type: contracts.DriverCmdRegister,
		Data: driverData.Driver,
	}

	if err := connManager.SendMessage(userID, msg); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	// 初始化queue consumers
	queues := []string{
		messaging.DriverCmdTripRequestQueue,
	}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rmq, connManager, q)

		if err := consumer.Start(); err != nil {
			log.Printf("failed to start consumer for queue: %s: err: %v", q, err)
		}
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		type driverMessage struct {
			Type string `json:"type"`
			Data []byte `json:"data"`
		}

		var driverMsg driverMessage
		if err := json.Unmarshal(message, &driverMsg); err != nil {
			log.Printf("Error unmarshall driver message %v", err)
			continue
		}

		// 处理不同的消息类型
		switch driverMsg.Type {
		case contracts.DriverCmdLocation:
			// 后面可以按需去补充添加
			continue
		case contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline:
			// 前端把司机的选择通过WS传到后端，做响应
			if err := rmq.PublishMessage(ctx, driverMsg.Type, contracts.AmqpMessage{
				OwnerID: userID,
				Data:    driverMsg.Data,
			}); err != nil {
				log.Printf("Error publishing message to RabbitMQ: %v", err)
			}
		default:
			log.Printf("Unknown message type: %s", driverMsg.Type)
		}
	}
}
