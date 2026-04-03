package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)
type Order struct {
	ID          string `json:"id"`
	ProductID   string `json:"product_id"`
	Quantity    int    `json:"quantity"`
}

var ctx = context.Background()
var rdb *redis.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if order.ProductID == "" || order.Quantity <= 0 {
    	http.Error(w, "ProductID and Quantity are required", http.StatusBadRequest)
    return
	}
	order.ID = uuid.NewString()
	
	orderJSON, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "Failed to marshal order", http.StatusInternalServerError)
		return
	}

	err = rdb.Publish(ctx, "orders", orderJSON).Err()
	if err != nil {
		http.Error(w, "Failed to send order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}


func main() {
	http.HandleFunc("/orders", createOrderHandler)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
        if err := rdb.Ping(ctx).Err(); err != nil {
            log.Printf("Readiness check failed: %v", err)
            http.Error(w, "Redis not reachable", http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Ready"))
    })

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}