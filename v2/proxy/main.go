package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// PluginConnection represents a plugin connection
type PluginConnection struct {
	Conn net.Conn
}

// ProxyServer represents the proxy server
type ProxyServer struct {
	plugins      map[string]*PluginConnection
	pluginsMutex sync.Mutex
}

// NewProxyServer creates a new ProxyServer
func NewProxyServer() *ProxyServer {
	return &ProxyServer{
		plugins: make(map[string]*PluginConnection),
	}
}

// HandlePluginConnection handles plugin connections
func (ps *ProxyServer) HandlePluginConnection(conn net.Conn) {
	defer conn.Close()

	// Generate a unique ID for the plugin connection (e.g., using IP and port)
	pluginID := conn.RemoteAddr().String()

	ps.pluginsMutex.Lock()
	ps.plugins[pluginID] = &PluginConnection{Conn: conn}
	ps.pluginsMutex.Unlock()

	log.Printf("Plugin connected: %s", pluginID)

	// Keep the connection open
	io.Copy(io.Discard, conn)

	ps.pluginsMutex.Lock()
	delete(ps.plugins, pluginID)
	ps.pluginsMutex.Unlock()

	log.Printf("Plugin disconnected: %s", pluginID)
}

// HandleHTTPProxy handles HTTP proxy requests
func (ps *ProxyServer) HandleHTTPProxy(w http.ResponseWriter, r *http.Request) {
	ps.pluginsMutex.Lock()
	defer ps.pluginsMutex.Unlock()

	// For simplicity, we'll just use the first available plugin
	for pluginID, pluginConn := range ps.plugins {
		log.Printf("Forwarding request to plugin: %s", pluginID)

		// Forward the HTTP request to the plugin
		err := r.Write(pluginConn.Conn)
		if err != nil {
			log.Printf("Error forwarding request to plugin: %s", err)
			http.Error(w, "Error forwarding request to plugin", http.StatusInternalServerError)
			return
		}

		// Read the response from the plugin and write it back to the client
		_, err = io.Copy(w, pluginConn.Conn)
		if err != nil {
			log.Printf("Error reading response from plugin: %s", err)
			return
		}
		return
	}

	http.Error(w, "No plugin available", http.StatusServiceUnavailable)
}

// HandleWebSocketProxy handles WebSocket proxy requests
func (ps *ProxyServer) HandleWebSocketProxy(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	ps.pluginsMutex.Lock()
	defer ps.pluginsMutex.Unlock()

	// For simplicity, we'll just use the first available plugin
	for pluginID, pluginConn := range ps.plugins {
		log.Printf("Forwarding WebSocket connection to plugin: %s", pluginID)

		// Forward the WebSocket connection to the plugin
		go func() {
			defer conn.Close()
			// defer pluginConn.Conn.Close()

			// Copy data from WebSocket to plugin connection
			go func() {
				for {
					_, msg, err := conn.ReadMessage()
					if err != nil {
						log.Printf("Error reading WebSocket message: %v", err)
						return
					}
					_, err = pluginConn.Conn.Write(msg)
					if err != nil {
						log.Printf("Error writing to plugin connection: %v", err)
						return
					}
				}
			}()

			// Copy data from plugin connection to WebSocket
			buf := make([]byte, 1024)
			for {
				n, err := pluginConn.Conn.Read(buf)
				if err != nil {
					log.Printf("Error reading from plugin connection: %v", err)
					return
				}
				err = conn.WriteMessage(websocket.BinaryMessage, buf[:n])
				if err != nil {
					log.Printf("Error writing WebSocket message: %v", err)
					return
				}
			}
		}()
		return
	}

	http.Error(w, "No plugin available", http.StatusServiceUnavailable)
}

func main() {
	proxyServer := NewProxyServer()

	// Listen for plugin connections
	go func() {
		pluginListener, err := net.Listen("tcp", ":9000")
		if err != nil {
			log.Fatalf("Error starting plugin listener: %v", err)
		}
		defer pluginListener.Close()

		log.Println("Listening for plugin connections on :9000")

		for {
			conn, err := pluginListener.Accept()
			if err != nil {
				log.Printf("Error accepting plugin connection: %v", err)
				continue
			}

			go proxyServer.HandlePluginConnection(conn)
		}
	}()

	// Listen for HTTP proxy requests
	http.HandleFunc("/http", proxyServer.HandleHTTPProxy)

	// Listen for WebSocket proxy requests
	http.HandleFunc("/ws", proxyServer.HandleWebSocketProxy)

	log.Println("Listening for proxy requests on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting proxy server: %v", err)
	}
}
