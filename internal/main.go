package main

import (
	"log"
	"net/http"
)

func main() {
	s := MessageServer{}
	h := Handler(s)

	log.Println("Server listening on internal container port:", *portNum)
	log.Fatal(http.ListenAndServe("3001", h))
}
