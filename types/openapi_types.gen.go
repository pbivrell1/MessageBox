package types
// Package types provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version (devel) DO NOT EDIT.

// ComposedMessage defines model for ComposedMessage.
type ComposedMessage struct {
	Body      *string     `json:"body,omitempty"`
	Recipient interface{} `json:"recipient"`
	Sender    string      `json:"sender"`
	Subject   string      `json:"subject"`
}

// GroupCreation defines model for GroupCreation.
type GroupCreation struct {
	Groupname string   `json:"groupname"`
	Usernames []string `json:"usernames"`
}

// GroupRecipient A message recipient representing a group of users
type GroupRecipient struct {
	Groupname *string `json:"groupname,omitempty"`
}

// Message defines model for Message.
type Message struct {
	Id        int64       `json:"id"`
	Re        int64       `json:"re,omitempty"`
	Sender    string      `json:"sender"`
	Recipient interface{} `json:"recipient"`
	Subject   string      `json:"subject"`
	Body      *string     `json:"body,omitempty"`
	SentAt    string      `json:"sentAt"`
}

// ReplyMessage defines model for ReplyMessage.
type ReplyMessage struct {
	Body    *string `json:"body,omitempty"`
	Sender  string  `json:"sender"`
	Subject string  `json:"subject"`
}

// UserRecipient A message recipient representing a single user
type UserRecipient struct {
	Username *string `json:"username,omitempty"`
}

// UserRegistration defines model for UserRegistration.
type UserRegistration struct {
	Username string `json:"username"`
}

// PostGroupsJSONBody defines parameters for PostGroups.
type PostGroupsJSONBody GroupCreation

// PostMessagesJSONBody defines parameters for PostMessages.
type PostMessagesJSONBody ComposedMessage

// PostMessagesIdRepliesJSONBody defines parameters for PostMessagesIdReplies.
type PostMessagesIdRepliesJSONBody ReplyMessage

// PostUsersJSONBody defines parameters for PostUsers.
type PostUsersJSONBody UserRegistration

// PostGroupsJSONRequestBody defines body for PostGroups for application/json ContentType.
type PostGroupsJSONRequestBody PostGroupsJSONBody

// PostMessagesJSONRequestBody defines body for PostMessages for application/json ContentType.
type PostMessagesJSONRequestBody PostMessagesJSONBody

// PostMessagesIdRepliesJSONRequestBody defines body for PostMessagesIdReplies for application/json ContentType.
type PostMessagesIdRepliesJSONRequestBody PostMessagesIdRepliesJSONBody

// PostUsersJSONRequestBody defines body for PostUsers for application/json ContentType.
type PostUsersJSONRequestBody PostUsersJSONBody
