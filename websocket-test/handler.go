package main

import (
	"log"
	"net/http"
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

	messagingConnector(userID, conn)
}
