package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"time"
)

type MessageServer struct {
	DbConn *redis.Client
}

//TODO: Create one persistent connection to redis for the MessageServer instead of opening a new one on each request
//TODO: Better error handling and meaningful response headers/bodies
//TODO: Golang code golf

func (m MessageServer) PostGroups(w http.ResponseWriter, r *http.Request) {
	var newGroup PostGroupsJSONRequestBody
	ctx := context.Background()
	err := json.NewDecoder(r.Body).Decode(&newGroup)
	if err != nil {
		//TODO: better error handling, messages etc
		log.Printf("Post groups json decoder error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if newGroup.Groupname == "" || len(newGroup.Usernames) == 0 {
		log.Printf("Post groups invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key := fmt.Sprintf("group:%s", newGroup.Groupname)
	val, err := m.DbConn.Exists(ctx, key).Result()
	if err != nil {
		log.Printf("Post groups redis connection error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if val != 0 {
		log.Println("Post groups duplicate key")
		w.WriteHeader(http.StatusConflict)
		err = json.NewEncoder(w).Encode(newGroup)
		return
	}
	for _, name := range newGroup.Usernames {
		exists, err := m.DbConn.SIsMember(ctx, key, name).Result()
		if err != nil {
			log.Println("Post users redis error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if exists == true {
			//don't add a username if it doesn't correspond to a registered user
			//error here probably
			continue
		}
		err = m.DbConn.SAdd(ctx, key, name).Err()
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
	var newMessage PostMessagesJSONRequestBody
	ctx := context.Background()
	err := json.NewDecoder(r.Body).Decode(&newMessage)
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
	id, err := m.DbConn.HIncrBy(ctx, "idCount", "nextID", 1).Result()
	if err != nil {
		log.Printf("Post messages redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	key := fmt.Sprintf("message:%d", id)
	sendTime := time.Now().String()
	recipients, err := json.Marshal(newMessage.Recipient)
	if err != nil {
		log.Printf("Error marshalling json:%s", err)
	}
	err = m.DbConn.HSet(ctx, key, "re", 0, "sender", newMessage.Sender, "recipient", string(recipients[:]),
		"subject", newMessage.Subject, "body", *newMessage.Body, "sentAt", sendTime).Err()
	if err != nil {
		log.Printf("Post messages redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fullMessage := Message{
		Id:        id,
		Sender:    newMessage.Sender,
		Recipient: newMessage.Recipient,
		Subject:   newMessage.Subject,
		Body:      newMessage.Body,
		SentAt:    sendTime,
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

}

func (m MessageServer) GetMessagesIdReplies(w http.ResponseWriter, r *http.Request, id int64) {

}

func (m MessageServer) PostMessagesIdReplies(w http.ResponseWriter, r *http.Request, id int64) {

}

func (m MessageServer) PostUsers(w http.ResponseWriter, r *http.Request) {
	var newUser PostUsersJSONRequestBody
	ctx := context.Background()
	err := json.NewDecoder(r.Body).Decode(&newUser)
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
	exists, err := m.DbConn.SIsMember(ctx, "users", newUser.Username).Result()
	if err != nil {
		log.Printf("Post users redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists == true {
		log.Println("Post users duplicate value")
		w.WriteHeader(http.StatusConflict)
		err = json.NewEncoder(w).Encode(newUser)
		return
	}
	err = m.DbConn.SAdd(ctx, "users", newUser.Username).Err()
	if err != nil {
		log.Printf("Post users redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}

func (m MessageServer) GetUsersUsernameMailbox(w http.ResponseWriter, r *http.Request, username string) {

}
