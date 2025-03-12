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
)

var ctx = context.Background()

// Initialize Redis client
var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379", // Use your Redis server address
})

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var Data = []string{}

// DisbursementRequest struct to handle disbursement request payload
type MakePaymentRequest struct {
	LoanID int64 `json:"loan_id"`
}

type TimeoutTestingRequest struct {
	StringText string `json:"string_text"`
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/timeout", TimeoutTesting).Methods(http.MethodPost)
	r.HandleFunc("/get_data", GetData).Methods(http.MethodGet)

	fmt.Println("Server is listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func TimeoutTesting(w http.ResponseWriter, r *http.Request) {
	var req TimeoutTestingRequest
	var resp Response
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	slowFunction(req.StringText)
	// If the key exists

	resp.Message = "Success"
	resp.Data = Data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func GetData(w http.ResponseWriter, r *http.Request) {
	var resp Response

	resp.Message = "Success"
	resp.Data = Data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func slowFunction(data string) string {
	// Simulate a delay in processing
	time.Sleep(5000 * time.Millisecond)
	Data = append(Data, data)
	return "Function completed successfully"
}
