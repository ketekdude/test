package main

import (
	"encoding/json"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

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
