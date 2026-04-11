package main

import (
	"log"
	"net/http"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/util"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleRiderWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
	}

	defer conn.Close()

	// 获取用户ID
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Printf("The user id is empty")
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

func handleDriverWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
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

	type Driver struct {
		Id           string `json:"id"`
		Name         string `json:"name"`
		ProfileImage string `json:"profilePicture"`
		CarPlate     string `json:"carPlate"`
		PackageSlug  string `json:"packageSlug"`
	}

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: Driver{
			Id:           userID,
			Name:         "yuxin",
			ProfileImage: util.GetRandomAvatar(),
			CarPlate:     "22592",
			PackageSlug:  packageSlug,
		},
	}

	if err := conn.WriteJSON(msg); err != nil {
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
