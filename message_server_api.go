package main

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
)

type MessageServer struct {
	//Conn redis.Conn
}

//TODO: Create one persistent connection to redis for the MessageServer instead of opening a new one on each request
//TODO: Better error handling and meaningful response headers/bodies
//TODO: Code golf

func (m MessageServer) PostGroups(w http.ResponseWriter, r *http.Request) {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	var newGroup PostGroupsJSONRequestBody
	json.NewDecoder(r.Body).Decode(&newGroup)
	if err != nil {
		//TODO: better error handling, messages etc
		log.Println("Post groups json decoder error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if newGroup.Groupname == "" {
		log.Println("Post groups invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key := fmt.Sprintf("group:%s", newGroup.Groupname)
	val, err := conn.Do("EXISTS", key)
	if err != nil {
		log.Println("Post groups redis connection error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if val.(int64) > 0 {
		log.Println("Post groups duplicate key")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(newGroup)
		return
	}
	for _, name := range newGroup.Usernames {
		val, err = conn.Do("SADD", key, name)
		if err != nil {
			log.Println("Post groups redis error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusCreated)
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
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Println("Post users error initializing redis socket")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	var newUser PostUsersJSONBody
	err = json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		log.Println("Post users json decoder error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if newUser.Username == "" {
		log.Println("Post users invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	val, err := conn.Do("SISMEMBER", "users", newUser.Username)
	if err != nil {
		log.Println("Post users redis error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if val.(int64) == 0 {
		log.Println("Post users duplicate value")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(newUser)
		return
	}
	_, err = conn.Do("SADD", "users", newUser.Username)
	if err != nil {
		log.Println("Post users redis error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}

func (m MessageServer) GetUsersUsernameMailbox(w http.ResponseWriter, r *http.Request, username string) {
	//TODO: Create one persistent connection to redis for the MessageServer instead of opening a new one on each request
	conn, err := redis.Dial("tcp", "redis:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
}
