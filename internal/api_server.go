package main

import "net/http"

type MessageServer struct{}

func (m MessageServer) PostGroups(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (m MessageServer) PostMessages(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (m MessageServer) GetMessagesId(w http.ResponseWriter, r *http.Request, id float32) {
	//TODO implement me
	panic("implement me")
}

func (m MessageServer) GetMessagesIdReplies(w http.ResponseWriter, r *http.Request, id float32) {
	//TODO implement me
	panic("implement me")
}

func (m MessageServer) PostMessagesIdReplies(w http.ResponseWriter, r *http.Request, id float32) {
	//TODO implement me
	panic("implement me")
}

func (m MessageServer) PostUsers(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (m MessageServer) GetUsersUsernameMailbox(w http.ResponseWriter, r *http.Request, username string) {
	//TODO implement me
	panic("implement me")
}
