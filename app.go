package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// Initialize Redis client
var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379", // Use your Redis server address
})

type UserBalance struct {
	UserID  int64   `json:"user_id"`
	Balance float64 `json:"balance"`
}
type Response struct {
	Message string  `json:"message"`
	Balance float64 `json:"balance"`
}

// balance := make(map[int64]UserBalance{})
var BalanceData = make(map[int64]UserBalance)

func validateUserBalance(userID int64, amount float64) (err error) {
	if BalanceData[userID].UserID == 0 {
		//user does not exist
		err = fmt.Errorf("UserID does not exist")
		return
	}

	if BalanceData[userID].Balance == 0 {
		err = fmt.Errorf("balance is empty, please top up")
		return
	}

	if BalanceData[userID].Balance < amount {
		err = fmt.Errorf("does not have enough balance")
		return
	}

	return
}

// Mock function to update the database
func updateDatabase(userID int64, amount float64) error {
	// Simulate database update
	// Here you would have actual code to update your database
	userBalance := BalanceData[userID]
	userBalance.Balance = userBalance.Balance - amount
	BalanceData[userID] = userBalance
	fmt.Printf("Database updated for user %d with new balance %f\n", userID, amount)
	return nil
}

// DisbursementRequest struct to handle disbursement request payload
type DisbursementRequest struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

func InitBalanceData() {
	BalanceData[1] = UserBalance{UserID: 1, Balance: 1000}
	BalanceData[2] = UserBalance{UserID: 2, Balance: 2000}
	BalanceData[3] = UserBalance{UserID: 3, Balance: 3000}
}

// DisburseBalance handles the disbursement of the user's balance
func DisburseBalance(w http.ResponseWriter, r *http.Request) {
	var req DisbursementRequest
	var resp Response
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	fmt.Println("latest balance: ", BalanceData[req.UserID])
	//user validation
	lockKey := fmt.Sprintf("db:update:lock:%d", req.UserID)
	lockValue := strconv.FormatInt(time.Now().UnixNano(), 10)
	errVal := validateUserBalance(req.UserID, req.Amount)
	if errVal != nil {
		// if the balance already 0, no need to process the request even further.
		resp.Message = errVal.Error()
		resp.Balance = BalanceData[req.UserID].Balance
		// Encode the error response as JSON
		errJSON, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusTooManyRequests)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(errJSON)
		return
	}
	// Acquire lock
	//when use setNX we will prevent all update on this balance
	//so we will prevent an update with a not latest data.
	locked, err := redisClient.SetNX(ctx, lockKey, lockValue, 100*time.Second).Result()
	if err != nil {
		http.Error(w, "Failed to acquire lock", http.StatusInternalServerError)
		return
	}
	if !locked {
		resp.Message = "Could not acquire lock"
		resp.Balance = BalanceData[req.UserID].Balance
		// Encode the error response as JSON
		errJSON, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusTooManyRequests)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(errJSON)
		return
		// http.Error(w, "Could not acquire lock", http.StatusTooManyRequests)
		// return
	}
	defer redisClient.Del(ctx, lockKey) // Release lock after operation
	//time.sleep was used to check if the NX lock is working
	//you can dismiss it later.
	// time.Sleep(5 * time.Second)
	// Perform atomic decrement operation

	// Update the database
	if err := updateDatabase(req.UserID, req.Amount); err != nil {
		// Revert the decrement in case of database update failure
		resp.Message = err.Error()
		resp.Balance = BalanceData[req.UserID].Balance
		// Encode the error response as JSON
		errJSON, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(errJSON)
		return
	}

	resp.Message = "Disbursement successful"
	resp.Balance = BalanceData[req.UserID].Balance

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	InitBalanceData()
	http.HandleFunc("/disburse", DisburseBalance)
	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
