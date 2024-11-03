package main

import (
	"encoding/json"
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

func handleOnlineList(w http.ResponseWriter, r *http.Request) {
	var resp Response

	resp.Message = "success"
	data := struct {
		OnlineList  interface{}
		TotalOnline int
	}{
		OnlineList:  websocketConns,
		TotalOnline: getOnlineList(),
	}
	resp.Data = data

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
