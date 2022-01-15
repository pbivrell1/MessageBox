package main

import (
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
)

func main() {
	conn := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", //no auth
		DB:       0,
	})
	s := MessageServer{
		DbConn: conn,
	}
	h := Handler(s)

	log.Println("Server listening on internal container port 3001")
	log.Fatal(http.ListenAndServe(":3001", h))
}
