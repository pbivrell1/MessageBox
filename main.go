package main

import (
	"github.com/mgheebs/MessageBox.git/api"
	"github.com/mgheebs/MessageBox.git/server"
	"log"
	"net/http"
	"os"
)

//TODO: would love to make the log locations, redis address, all part of a configuration, ran out of time
func main() {
	logFile, err := os.OpenFile("/var/log/messagebox.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening log file: %s", err)
	}
	defer logFile.Close()
	s := api.MessageServer{}
	s.InitMessageServer(logFile)
	h := server.Handler(s)
	log.Fatal(http.ListenAndServe(":3001", h))
}
