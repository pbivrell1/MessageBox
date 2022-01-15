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

//TODO: Better error handling and meaningful response headers/bodies
//TODO: add content type to responses in a common middleware!!
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
			//add a warning or possibly error here probably
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
	respMessage := Message{
		Id:        id,
		Sender:    newMessage.Sender,
		Recipient: newMessage.Recipient,
		Subject:   newMessage.Subject,
		Body:      newMessage.Body,
		SentAt:    sendTime,
	}
	jsonMessage, err := json.Marshal(respMessage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error marshalling json:%s", err)
		return
	}
	err = m.DbConn.Set(ctx, key, jsonMessage, 0).Err()
	if err != nil {
		log.Printf("Post messages redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err != nil {
		log.Printf("Post messages redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write(jsonMessage)
}

func (m MessageServer) GetMessagesId(w http.ResponseWriter, r *http.Request, id int64) {
	ctx := context.Background()
	key := fmt.Sprintf("message:%d", id)
	val, err := m.DbConn.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Printf("Get messages/id key not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Get messages/id redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write([]byte(val))
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
