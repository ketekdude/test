package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/cors"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
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
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type BillingHeader struct {
	BillingID            int64           `json:"billing_id"`
	LoanID               int64           `json:"loan_id"`
	UserID               int64           `json:"user_id"`
	LoanAmount           float64         `json:"loan_amount"`
	OutstandingAmount    float64         `json:"outstanding_amount"`
	BillingStatus        int             `json:"billing_status"`
	BillingStatusWording string          `json:"billing_status_wording"`
	IsDelinquent         bool            `json:"is_delinquent"`
	DetailBilling        []BillingDetail `json:"billing_detail"`
}

type BillingDetail struct {
	BillingDetailID            int64     `json:"billing_detail_id"`
	BillingID                  int64     `json:"billing_id"`
	Amount                     float64   `json:"amount"`
	ExpiredDate                time.Time `json:"expired_date"`
	BillingDetailStatus        int       `json:"billing_detail_status"`
	BillingDetailStatusWording string    `json:"billing_detail_status_wording"`
	PaymentID                  int64     `json:"payment_id"`
}

// balance := make(map[int64]UserBalance{})
var BillingDetails = make(map[int64]BillingDetail)

var BillingData = make(map[int64]BillingHeader)

// DisbursementRequest struct to handle disbursement request payload
type MakePaymentRequest struct {
	LoanID int64 `json:"loan_id"`
}

type GetOutstandingRequest struct {
	LoanID int64 `json:"loan_id"`
}

// billingstatus 1=open 2=paid
// billingstatusdetail 1=open 2=paid 3=late
func InitBillingData() {
	Detail1 := []BillingDetail{}
	Detail2 := []BillingDetail{}
	for i := 0; i < 50; i++ {
		duration := fmt.Sprintf("%dh", 24*(i+1)*7)
		dur, err := time.ParseDuration(duration)
		if err != nil {
			panic("fail init data")
		}
		temp := BillingDetail{
			BillingDetailID:            int64(i),
			BillingID:                  1,
			Amount:                     110000,
			ExpiredDate:                time.Now().Add(dur),
			BillingDetailStatus:        1,
			BillingDetailStatusWording: "open",
		}
		Detail1 = append(Detail1, temp)

		temp.ExpiredDate = temp.ExpiredDate.Add(-336 * time.Hour)
		if temp.ExpiredDate.Before(time.Now()) {
			temp.BillingDetailID = int64(i)
			temp.BillingID = 2
			temp.BillingDetailStatus = 3
			temp.BillingDetailStatusWording = "late"
		}
		Detail2 = append(Detail2, temp)
	}

	//loan with 0 late billing payment
	BillingData[1] = BillingHeader{
		BillingID:            1,
		LoanID:               1,
		UserID:               1,
		LoanAmount:           5000000,
		OutstandingAmount:    5500000,
		BillingStatus:        1,
		BillingStatusWording: "Open",
		IsDelinquent:         false,
		DetailBilling:        Detail1,
	}

	//loan with 2 late billing payment
	BillingData[2] = BillingHeader{
		BillingID:            2,
		LoanID:               2,
		UserID:               2,
		LoanAmount:           5000000,
		OutstandingAmount:    5500000,
		BillingStatus:        1,
		BillingStatusWording: "Open",
		IsDelinquent:         true,
		DetailBilling:        Detail2,
	}

}

// DisburseBalance handles the disbursement of the user's balance
func GetLoanBilling(w http.ResponseWriter, r *http.Request) {
	var req GetOutstandingRequest
	var resp Response
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	fmt.Println(req)
	data, ok := BillingData[req.LoanID]
	// If the key exists
	if !ok {
		// Do something
		resp.Message = "Fail, billing for those loan is not exist"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp.Message = "Success"
	resp.Data = data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func TestPrint(w http.ResponseWriter, r *http.Request) {
	var resp Response
	var input map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&input)

	resp.Message = "Success"
	resp.Data = input
	fmt.Println(input)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Mock function to update the database

func GeneratePaymentRequest(loanID int64) error {
	//dummy function
	return nil
}

// Mock function to update the database
func updateDatabase(loanID int64) error {
	// Simulate database update
	// Here you would have actual code to update your database
	data := BillingData[loanID]
	data.OutstandingAmount -= data.DetailBilling[0].Amount
	i := 0
	paid := false
	for !paid {
		if data.DetailBilling[i].BillingDetailStatus != 3 {
			data.DetailBilling[i].BillingDetailStatus = 3
			data.DetailBilling[i].BillingDetailStatusWording = "paid"

			paid = true
		}
		i++
	}
	BillingData[loanID] = data
	fmt.Printf("Database updated for loan %d with new outstanding %f\n", loanID, data.OutstandingAmount)
	return nil
}

// DisburseBalance handles the disbursement of the user's balance
func MakePayment(w http.ResponseWriter, r *http.Request) {
	var req MakePaymentRequest
	var resp Response
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	//set locking to prevent double payment
	lockKey := fmt.Sprintf("db:update:lock:%d", req.LoanID)
	lockValue := strconv.FormatInt(time.Now().UnixNano(), 10)
	locked, err := redisClient.SetNX(ctx, lockKey, lockValue, 100*time.Second).Result()
	if err != nil {
		http.Error(w, "Failed to acquire lock", http.StatusInternalServerError)
		return
	}
	if !locked {
		resp.Message = "payment failed, due to double payment please try again later"
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
	time.Sleep(5 * time.Second)
	//use time sleep here to proof that the locking works
	defer redisClient.Del(ctx, lockKey) // Release lock after operation
	data, ok := BillingData[req.LoanID]
	// If the key exists
	if !ok {
		// Do something
		resp.Message = "Fail, billing for those loan is not exist"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if data.OutstandingAmount == 0 || data.BillingStatus == 2 {
		resp.Message = "loan already paid"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	//we don't need to check the paymount because we consider it is handled by account team
	//we consider the hook request from payment will continue here also
	err = GeneratePaymentRequest(req.LoanID)
	if err != nil {
		resp.Message = "Payment Failed, please try again later"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}
	err = updateDatabase(req.LoanID)
	if err != nil {
		//we can consider to publish nsq to be the retrier here
		resp.Message = "Failed Update Oustanding, please try again later"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp.Message = "Success"
	resp.Data = BillingData[req.LoanID]
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	InitBillingData()
	// http.HandleFunc("/disburse", DisburseBalance)
	// http.HandleFunc("/payment", DisburseBalance)
	r := mux.NewRouter()

	// r.Use(corsMiddleware)

	// Define routes with specific methods
	// r.HandleFunc("/disburse", DisburseBalance).Methods(http.MethodPost)
	r.HandleFunc("/make_payment", MakePayment).Methods(http.MethodPost)
	//this endpoint cover all the billing data for oustanding & isdelinquent
	r.HandleFunc("/get_loan_billing", GetLoanBilling).Methods(http.MethodPost)
	r.HandleFunc("/api/print", TestPrint).Methods(http.MethodPost, http.MethodOptions)

	fmt.Println("Server is listening on :5000")
	// Set up CORS middleware
	corsOptions := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                      // Your Laravel server's IP or domain
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"}, // HTTP methods you want to allow
		AllowedHeaders:   []string{"authorization,content-type"},
		AllowCredentials: true,
	})
	// Apply CORS middleware to the Mux router
	handler := corsOptions.Handler(r)
	// http.ListenAndServe(":8080", r)
	log.Fatal(http.ListenAndServe(":5000", handler))

	// fmt.Println("Server is running on port 8080")
	// log.Fatal(http.ListenAndServe(":8080", nil))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Change "*" to a specific domain for better security
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "content-type, authorization")

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
