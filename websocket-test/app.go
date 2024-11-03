package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	ctx            = context.Background()
	redisClient    *redis.Client
	websocketConns = make(map[string]*websocket.Conn) // Map of user IDs to WebSocket connections
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Body      string `json:"body"`
	Timestamp string `json:"timestamp"`
}

type Notification struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis server address
	})

	// WebSocket route
	r := mux.NewRouter()
	// Define routes with specific methods
	r.HandleFunc("/chat", handleWebSocket).Methods(http.MethodGet)

	// Start listening for messages from Redis Pub/Sub
	go listenToRedisChannel()

	// Start the HTTP server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
