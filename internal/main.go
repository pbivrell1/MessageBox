package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	portNum := flag.Int("p", 8080, "-p <port number> - Define a port to listen on (default:8080)")

	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "You've connected to Gabe's MessageBox")
	})

	log.Println("Server listening on port:", *portNum)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *portNum), nil))
}
