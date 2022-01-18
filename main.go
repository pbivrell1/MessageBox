package main

import (
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"os"
)

//TODO: would love to make the log locations, redis address, all part of a configuration, ran out of time
func main() {
	s := MessageServer{}
	logFile, err := os.OpenFile("/var/log/messagebox.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening log file: %s", err)
	}
	defer logFile.Close()
	s.DbConn = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", //no auth
		DB:       0,
	})
	s.logger = log.New(logFile, "MessageboxAPI:", log.LstdFlags)
	h := Handler(s)
	s.logger.Println("Server listening on internal container port 3001")
	log.Fatal(http.ListenAndServe(":3001", h))
}
