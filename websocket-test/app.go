package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	ctx            = context.Background()
	redisClient    *redis.Client
	websocketConns = make(map[string]*websocket.Conn) // Map of user IDs to WebSocket connections
)

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
	r.HandleFunc("/disburse", handleWebSocket).Methods(http.MethodGet)

	// Start listening for messages from Redis Pub/Sub
	go listenToRedisChannel()

	// Start the HTTP server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WebSocket handler
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from query parameters
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	// Upgrade HTTP request to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	websocketConns[userID] = conn
	log.Printf("User %s connected\n", userID)

	// Check for any queued messages for the user in Redis
	deliverQueuedMessages(userID, conn)
	// Listen for messages from the WebSocket connection
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			delete(websocketConns, userID) // Remove user on disconnect
			return
		}

		// Unmarshal the JSON message
		var message Message
		if err := json.Unmarshal(msg, &message); err != nil {
			log.Println("Error unmarshalling message:", err)
			continue
		}
		message.Timestamp = time.Now().Format("2006-01-02 15:04:05")
		// Log the received message
		fmt.Printf("Received message from %s to %s: %s\n", message.Sender, message.Recipient, message.Body)
		msg, _ = json.Marshal(message)
		// Check if recipient is online
		if _, exists := websocketConns[message.Recipient]; exists {
			// Publish the message to the Redis channel for the recipient
			publishMessageToRedis(message.Recipient, msg)
		} else {
			// Notify the sender that the recipient is offline
			notification := Notification{
				Status:  "error",
				Message: fmt.Sprintf("Recipient %s is offline. Message will be queued.", message.Recipient),
			}
			notificationMsg, _ := json.Marshal(notification)
			err := conn.WriteMessage(websocket.TextMessage, notificationMsg)
			if err != nil {
				log.Printf("Error sending notification to %s: %v\n", message.Sender, err)
				continue
			}
			fmt.Printf("Notification sent to %s: %s\n", message.Sender, notification.Message)
			// Here you can implement message queueing logic (e.g., store in Redis or a database)
			// Queue the message for later delivery in Redis
			queueMessageInRedis(message.Recipient, message)
		}
	}
}

// Publish a message to a Redis channel
func publishMessageToRedis(targetUserID string, message []byte) {
	err := redisClient.Publish(ctx, targetUserID, message).Err()
	if err != nil {
		log.Println("Error publishing message to Redis:", err)
	}
}

// Listen to Redis channels for incoming messages
func listenToRedisChannel() {
	// Subscribe to all relevant channels (e.g., for user messages)
	pubsub := redisClient.PSubscribe(ctx, "*") // Using wildcard to subscribe to all user channels
	defer pubsub.Close()

	for msg := range pubsub.Channel() {
		// The message payload is expected to be in JSON format
		targetUserID := msg.Channel
		if conn, exists := websocketConns[targetUserID]; exists {
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
			if err != nil {
				log.Printf("Error sending message to %s: %v\n", targetUserID, err)
				conn.Close()
				delete(websocketConns, targetUserID) // Remove on failure
			}
		}
	}
}

// Queue a message in Redis for offline users
func queueMessageInRedis(recipient string, message Message) {
	messageBytes, _ := json.Marshal(message)
	err := redisClient.RPush(ctx, "queue:"+recipient, messageBytes).Err()
	if err != nil {
		log.Println("Error queuing message in Redis:", err)
	}
}

// Deliver queued messages from Redis to the user
func deliverQueuedMessages(userID string, conn *websocket.Conn) {
	for {
		messageBytes, err := redisClient.LPop(ctx, "queue:"+userID).Result()
		if err != nil && err != redis.Nil {
			log.Println("Error retrieving queued message from Redis:", err)
			break
		}
		if messageBytes == "" {
			break // No more messages in the queue
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte(messageBytes))
		if err != nil {
			log.Printf("Error sending queued message to %s: %v\n", userID, err)
			conn.Close()
			delete(websocketConns, userID) // Remove on failure
			return
		}
	}
}
