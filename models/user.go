package models

type User struct {
	ID string `json:"ID" bson:"ID" binding:"required"`
}
