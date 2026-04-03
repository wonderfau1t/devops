package main

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

func main() {
	pubsub := rdb.Subscribe(ctx, "orders")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		log.Printf("New message from %s channel: %s", msg.Channel, msg.Payload)		
	}
}