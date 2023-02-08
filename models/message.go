package models

import "time"

type CreateMessageRequest struct {
	ID        string    `json:"ID" bson:"ID"`
	Text      string    `json:"Text" bson:"Text" binding:"required"`
	Sender    string    `json:"Sender" bson:"Sender" binding:"required"`
	Receiver  string    `json:"Receiver" bson:"Receiver" `
	CreatedAt time.Time `json:"Created_at" bson:"Created_at"`
}

type DBMessage struct {
	ID        string    `json:"ID" bson:"ID" binding:"required"`
	Text      string    `json:"Text" bson:"Text" binding:"required"`
	Sender    string    `json:"Sender" bson:"Sender"`
	Receiver  string    `json:"Receiver" bson:"Receiver" binding:"required"`
	CreatedAt time.Time `json:"Created_at" bson:"Created_at"`
}
