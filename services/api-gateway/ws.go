package main

import (
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/proto/driver"
	"ride-sharing/shared/util"
)

// 创建一个全局的连接manager
var (
	connManager = messaging.NewConnectionManager()
)

func handleRiderWebSocket(w http.ResponseWriter, r *http.Request) {
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

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		log.Printf("Receive message: %s", message)
	}

}

func handleDriverWebSocket(w http.ResponseWriter, r *http.Request) {
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

	ctx := r.Context()

	driverService, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	// Closing connections
	defer func() {
		connManager.Remove(userID)

		driverService.Client.UnregisterDriver(ctx, &driver.RegisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		})

		driverService.Close()

		log.Println("Driver unregistered: ", userID)
	}()

	type Driver struct {
		Id           string `json:"id"`
		Name         string `json:"name"`
		ProfileImage string `json:"profilePicture"`
		CarPlate     string `json:"carPlate"`
		PackageSlug  string `json:"packageSlug"`
	}

	msg := contracts.WSMessage{
		Type: contracts.DriverCmdRegister,
		Data: Driver{
			Id:           userID,
			Name:         "yuxin",
			ProfileImage: util.GetRandomAvatar(1),
			CarPlate:     "22592",
			PackageSlug:  packageSlug,
		},
	}

	driverData, err := driverService.Client.RegisterDriver(ctx, &driver.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	})
	if err != nil {
		log.Printf("Error registering driver: %v", err)
		return
	}

	if err := connManager.SendMessage(userID, msg); err != nil {
		log.Printf("Error sending message: %v", err)
		return
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
