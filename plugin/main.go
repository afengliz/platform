package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// handHttpHandler
func handHttpHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

// upgrader 将 HTTP 连接升级为 WebSocket 连接
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 默认允许所有来源的连接
	},
}

// handleWSConnection 处理 WebSocket 连接
func handleWSConnection(w http.ResponseWriter, r *http.Request) {
	// 将 HTTP 连接升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error while upgrading connection:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Client connected")

	for {
		// 读取消息
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error while reading message:", err)
			break
		}

		// 回显消息
		err = conn.WriteMessage(messageType, msg)
		if err != nil {
			log.Println("Error while writing message:", err)
			break
		}
	}
}

func main() {
	// 注册处理函数
	http.HandleFunc("/", handHttpHandler)
	http.HandleFunc("/ws", handleWSConnection)
	// 启动 HTTP 服务器
	fmt.Println("Starting server on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
