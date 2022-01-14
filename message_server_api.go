package main

import (
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
)

type MessageServer struct{}

func (m MessageServer) PostGroups(w http.ResponseWriter, r *http.Request) {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
}

func (m MessageServer) PostMessages(w http.ResponseWriter, r *http.Request) {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
}

func (m MessageServer) GetMessagesId(w http.ResponseWriter, r *http.Request, id float32) {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
}

func (m MessageServer) GetMessagesIdReplies(w http.ResponseWriter, r *http.Request, id float32) {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
}

func (m MessageServer) PostMessagesIdReplies(w http.ResponseWriter, r *http.Request, id float32) {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
}

func (m MessageServer) PostUsers(w http.ResponseWriter, r *http.Request) {
	var newUser PostUsersJSONBody
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if newUser.Username == "" {
		log.Println("Post users invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		//TODO: better error handling, messages etc
		log.Println("Post users json decoder error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Println("Post users initializing redis socket")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Try to write the key to redis, if it already exists bounce the request
	val, err := conn.Do("SETNX", newUser.Username, 1)
	if err != nil {
		log.Println("Error setting redis key")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if val.(int64) == 0 {
		log.Println("Post users duplicate value")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(newUser)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}

func (m MessageServer) GetUsersUsernameMailbox(w http.ResponseWriter, r *http.Request, username string) {
	conn, err := redis.Dial("tcp", "redis:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
}
