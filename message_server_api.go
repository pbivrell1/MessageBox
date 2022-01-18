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

//TODO: drastically improve the mailbox system. Store each user's mailbox as a hash maybe
//TODO: bounce unhandled scnearios where user doesnt exist / user replies to a message they didn't receive
//TODO: Better error handling and meaningful response headers/bodies
//TODO: Golang code golf
func (m MessageServer) PostGroups(w http.ResponseWriter, r *http.Request) {
	var newGroup PostGroupsJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&newGroup); err != nil {
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
	ctx := context.Background()
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
		if err = m.DbConn.SAdd(ctx, key, name).Err(); err != nil {
			log.Printf("Post groups redis error:%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newGroup)
}

func (m MessageServer) PostMessages(w http.ResponseWriter, r *http.Request) {
	var newMessage PostMessagesJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&newMessage); err != nil {
		log.Printf("Post users json decoder error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if newMessage.Sender == "" || newMessage.Recipient == "" || newMessage.Subject == "" {
		log.Printf("Post messages invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// use an incremented redis value to store the next available unique integer id
	// TODO: do this a better way maybe, works for now
	ctx := context.Background()
	id, err := m.DbConn.HIncrBy(ctx, "idCount", "nextID", 1).Result()
	if err != nil {
		log.Printf("Post messages redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	sendTime := time.Now().Round(0).String()
	// fill this struct then marshal it to store in redis as a json string
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
		log.Printf("Error marshalling json:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	key := fmt.Sprintf("message:%d", id)
	if err = m.DbConn.Set(ctx, key, jsonMessage, 0).Err(); err != nil {
		log.Printf("Post messages redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// if the message successfully posted, add it to the appropriate mailboxes
	recipients, ok := newMessage.Recipient.(map[string]interface{})
	if recipients == nil || ok == false {
		log.Println("Post messages unexpected typecasting error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// check whether the key in the map is username or groupname
	if val, found := recipients["username"]; found == true {
		// if the recipient is an individual user, add the message to the user's mailbox set
		key := fmt.Sprintf("mailbox:%s", val)
		err = m.DbConn.LPush(ctx, key, respMessage.Id).Err()
		if err != nil {
			log.Printf("Post messages redis error:%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if val, found := recipients["groupname"]; found == true {
		// if the message is for a group, get the names of all members of the group then add the mail to their boxes
		key := fmt.Sprintf("group:%s", val)
		groupMembers, err := m.DbConn.SMembers(ctx, key).Result()
		if err != nil {
			log.Printf("Post messages redis error:%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for _, user := range groupMembers {
			key := fmt.Sprintf("mailbox:%s", user)
			err := m.DbConn.LPush(ctx, key, respMessage.Id).Err()
			if err != nil {
				log.Printf("Post messages redis error:%s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(respMessage)
}

func (m MessageServer) GetMessagesId(w http.ResponseWriter, r *http.Request, id int64) {
	key := fmt.Sprintf("message:%d", id)
	ctx := context.Background()
	msg, err := m.DbConn.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			w.WriteHeader(http.StatusNotFound)
			log.Printf("get messages/id non nonexistent message id param")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Post users redis error:%s\n", err)
		}
		return
	}
	var message Message
	if err = json.Unmarshal([]byte(msg), &message); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error unmarshalling json:%s\n", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func (m MessageServer) GetMessagesIdReplies(w http.ResponseWriter, r *http.Request, id int64) {
	key := fmt.Sprintf("message:%d", id)
	ctx := context.Background()
	exists, err := m.DbConn.Exists(ctx, key).Result()
	if err != nil {
		log.Printf("Get message replies redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists == 0 {
		log.Printf("Get message replies received nonexistent key")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	key = fmt.Sprintf("replies:%d", id)
	replies, err := m.DbConn.SMembers(ctx, key).Result()
	if err != nil {
		log.Printf("Get user mailbox redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var allReplies []Message
	var message Message
	for _, reply := range replies {
		err = json.Unmarshal([]byte(reply), &message)
		if err != nil {
			log.Printf("Error unmarshalling json:%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		allReplies = append(allReplies, message)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allReplies)
}

func (m MessageServer) PostMessagesIdReplies(w http.ResponseWriter, r *http.Request, id int64) {
	// make sure this message exists and if not bounce immediately
	key := fmt.Sprintf("message:%d", id)
	ctx := context.Background()
	// get the message we want to reply to from db, if it doesn't exist bounce
	msg, err := m.DbConn.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("post message reply received non nonexistent message id param")
			w.WriteHeader(http.StatusNotFound)
		} else {
			log.Printf("Post users redis error:%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	var newMessage PostMessagesIdRepliesJSONBody
	json.NewDecoder(r.Body).Decode(&newMessage)
	if newMessage.Sender == "" || newMessage.Subject == "" {
		log.Printf("Post messages reply invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// stuff should be *somewhat* validated here, go ahead and unmarshal and get the original sender then build the reply
	var ogMessage Message
	err = json.Unmarshal([]byte(msg), &ogMessage)
	if err != nil {
		log.Printf("Error unmarshalling json:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	replyId, err := m.DbConn.HIncrBy(ctx, "idCount", "nextID", 1).Result()
	if err != nil {
		log.Printf("Post messages reply redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ogRecipient, ok := ogMessage.Recipient.(map[string]interface{})
	if ogRecipient == nil || ok == false {
		log.Println("Post messages reply unexpected typecasting error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	replyRecipient := make(map[string]interface{})
	var recipients []string
	if user, found := ogRecipient["username"]; found == true {
		// if the original recipient is a single user who exists, send this message back to the sender
		exists, err := m.DbConn.SIsMember(ctx, "users", user).Result()
		if err != nil {
			log.Printf("Post messages reply redis error:%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if exists == false {
			log.Println("Post messages reply original sender not found")
			w.WriteHeader(http.StatusGone)
			return
		}
		replyRecipient["username"] = ogMessage.Sender
		recipients = append(recipients, ogMessage.Sender)
	} else if group, found := ogRecipient["groupname"]; found == true {
		// if the original recipient was a group, reply to everyone in the group. Make sure to add the sender if they
		// aren't in the group
		key := fmt.Sprintf("group:%s", group)
		exists, err := m.DbConn.Exists(ctx, key).Result()
		if err != nil {
			log.Printf("Post messages reply redis error:%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if exists == 0 {
			log.Println("Post messages reply group not found")
			w.WriteHeader(http.StatusGone)
			return
		}
		// set the Recipient that will be used for the reply as well as the recipients for mailbox delivery
		replyRecipient["groupname"] = group
		recipients, err = m.DbConn.SMembers(ctx, key).Result()
		// if the original message's sender is not a member of the group, add them to the recipients
		found, err := m.DbConn.SIsMember(ctx, key, ogMessage.Sender).Result()
		if err != nil {
			log.Printf("Post messages redis error:%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if found == false {
			recipients = append(recipients, ogMessage.Sender)
		}
	} else {
		// need to handle this case when the original message is sent, but just in case
		log.Println("recipient json key was unrecognized")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sendTime := time.Now().Round(0).String()
	replyMessage := Message{
		Id:        replyId,
		Re:        ogMessage.Id,
		Sender:    newMessage.Sender,
		Recipient: replyRecipient,
		Subject:   newMessage.Subject,
		Body:      newMessage.Body,
		SentAt:    sendTime,
	}
	// iterate through the recipients and add to their mailboxes
	for _, user := range recipients {
		key := fmt.Sprintf("mailbox:%s", user)
		if err := m.DbConn.LPush(ctx, key, replyId).Err(); err != nil {
			log.Printf("Post messages redis error:%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	// marshal and store it
	jsonMessage, err := json.Marshal(replyMessage)
	if err != nil {
		log.Printf("Error marshalling json:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	key = fmt.Sprintf("message:%d", replyId)
	err = m.DbConn.Set(ctx, key, jsonMessage, 0).Err()
	if err != nil {
		log.Printf("Post messages reply redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// store in redis all the replies to a single message in a set, so we can get them easily later
	key = fmt.Sprintf("replies:%d", id)
	err = m.DbConn.SAdd(ctx, key, jsonMessage).Err()
	if err != nil {
		log.Printf("Post messages reply redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(replyMessage)
}

func (m MessageServer) PostUsers(w http.ResponseWriter, r *http.Request) {
	var newUser PostUsersJSONRequestBody
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		log.Printf("Post users json decoder error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// check that the request contains the required fields and the username is not already taken
	// if not taken, add to redis users set
	if newUser.Username == "" {
		log.Printf("Post users invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx := context.Background()
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}

func (m MessageServer) GetUsersUsernameMailbox(w http.ResponseWriter, r *http.Request, username string) {
	ctx := context.Background()
	// check first that the passed username exists in redis
	exists, err := m.DbConn.SIsMember(ctx, "users", username).Result()
	if err != nil {
		log.Printf("Get user mailbox redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists == false {
		log.Printf("Get user mailbox was passed an unknown username\n")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	key := fmt.Sprintf("mailbox:%s", username)
	listLen, err := m.DbConn.LLen(ctx, key).Result()
	if err != nil {
		log.Printf("Get user mailbox redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if listLen == 0 {
		log.Printf("Get user mailbox request for empty mailbox")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	mailIds, err := m.DbConn.LRange(ctx, key, 0, listLen-1).Result()
	if err != nil {
		log.Printf("Get user mailbox redis error:%s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var userMail []Message
	for _, id := range mailIds {
		var message Message
		key := fmt.Sprintf("message:%s", id)
		msg, err := m.DbConn.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				log.Printf("get mailbox message doesn't exist")
				w.WriteHeader(http.StatusNotFound)
			} else {
				log.Printf("Get user mailbox redis error:%s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		if err != nil {
			log.Printf("Get user mailbox redis error:%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal([]byte(msg), &message)
		if err != nil {
			log.Printf("Error unmarshalling json:%s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		userMail = append(userMail, message)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userMail)
}
