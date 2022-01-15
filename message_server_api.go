package main

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
	"time"
)

type MessageServer struct {
	//Conn redis.Conn
}

//TODO: Create one persistent connection to redis for the MessageServer instead of opening a new one on each request
//TODO: Better error handling and meaningful response headers/bodies
//TODO: Golang code golf

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
		log.Printf("Post groups json decoder error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if newGroup.Groupname == "" || len(newGroup.Usernames) == 0 {
		log.Printf("Post groups invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key := fmt.Sprintf("group:%s", newGroup.Groupname)
	val, err := conn.Do("EXISTS", key)
	if err != nil {
		log.Printf("Post groups redis connection error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if val.(int64) != 0 {
		log.Printf("Post groups duplicate key")
		w.WriteHeader(http.StatusConflict)
		err = json.NewEncoder(w).Encode(newGroup)
		return
	}
	for _, name := range newGroup.Usernames {
		val, err = conn.Do("SADD", key, name)
		if err != nil {
			log.Printf("Post groups redis error:%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newGroup)
}

func (m MessageServer) PostMessages(w http.ResponseWriter, r *http.Request) {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	var newMessage PostMessagesJSONRequestBody
	json.NewDecoder(r.Body).Decode(&newMessage)
	if err != nil {
		log.Printf("Post users json decoder error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if newMessage.Sender == "" || newMessage.Recipient == "" || newMessage.Subject == "" {
		log.Printf("Post messages invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id, err := conn.Do("HINCRBY", "idCount", "nextId", 1)
	if err != nil {
		log.Printf("Post messages redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	key := fmt.Sprintf("message:%d", id.(int64))
	_, err = conn.Do("HSET", key, "re", 0, "sender", newMessage.Sender, "recipient", newMessage.Recipient,
		"subject", newMessage.Subject, "body", newMessage.Body, "sentAt", time.Now().String())
	if err != nil {
		log.Printf("Post messages redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fullMessage := Message{
		Id:        id.(int64),
		Sender:    newMessage.Sender,
		Recipient: newMessage.Recipient,
		Subject:   newMessage.Subject,
		Body:      newMessage.Body,
		SentAt:    sendTime.String(),
	}
	if err != nil {
		log.Printf("Post messages redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(fullMessage)
}

func (m MessageServer) GetMessagesId(w http.ResponseWriter, r *http.Request, id int64) {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
}

func (m MessageServer) GetMessagesIdReplies(w http.ResponseWriter, r *http.Request, id int64) {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
}

func (m MessageServer) PostMessagesIdReplies(w http.ResponseWriter, r *http.Request, id int64) {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
}

func (m MessageServer) PostUsers(w http.ResponseWriter, r *http.Request) {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Printf("Post users error initializing redis socket")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	var newUser PostUsersJSONRequestBody
	err = json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		log.Printf("Post users json decoder error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if newUser.Username == "" {
		log.Printf("Post users invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	val, err := conn.Do("SISMEMBER", "users", newUser.Username)
	if err != nil {
		log.Printf("Post users redis error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if val.(int64) == 1 {
		log.Printf("Post users duplicate value")
		w.WriteHeader(http.StatusConflict)
		err = json.NewEncoder(w).Encode(newUser)
		return
	}
	_, err = conn.Do("SADD", "users", newUser.Username)
	if err != nil {
		log.Printf("Post users redis error:%s\n", err)
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
