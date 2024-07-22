package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
)

type ZoneInfo struct {
	ZoneId string `json:"zone_id"`
}

var zoneConn = make(map[string]net.Conn)

// http => tcp proxy
// tcp => http   zoneAgent
func main() {
	// tcp server
	tcpListen, err := net.Listen("tcp", ":8081")
	if err != nil {
		panic(err)
	}
	defer tcpListen.Close()

	go washout(tcpListen)

	// http 转 tcp
	http.HandleFunc("/:instanceID/http", handHttpHandler)
	// websocket 转 tcp
	http.HandleFunc("/:instanceID/ws", handleWSHandler)
	// 启动 HTTP 服务器
	fmt.Println("Starting server on port 8082...")
	err = http.ListenAndServe(":8082", nil)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

func washout(tcpListen net.Listener) error {
	for {
		conn, err := tcpListen.Accept()
		if err != nil {
			panic(err)
		}
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			panic(err)
		}
		var zoneInfo ZoneInfo
		err = json.Unmarshal(buf[:n], &zoneInfo)
		if err != nil {
			panic(err)
		}
		conn.Write([]byte(zoneInfo.ZoneId))
		zoneConn[zoneInfo.ZoneId] = conn
	}
}

func handHttpHandler(w http.ResponseWriter, r *http.Request) {
	instanceId := r.URL.Query().Get("instanceID")
	// 读取请求体
	body, err := r.GetBody()
	if err != nil {
		panic(err)
	}
	bytes := make([]byte, 1024)
	// 写入 tcp 连接
	zoneConn[instanceId].Write()

	for {
		n, err := zoneConn[instanceId].Read(bytes)
		if err != nil {
			panic(err)
		}
	}
}
func handleWSHandler(w http.ResponseWriter, r *http.Request) {
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
