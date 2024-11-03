package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

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
