package services

import (
	"github.com/gin-gonic/gin"
	"github.com/practice/golang-line/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateMessageFunc(message *models.CreateMessageRequest, messageCollection *mongo.Collection, ctx *gin.Context) (models.DBMessage, error) {

	res, err := messageCollection.InsertOne(ctx, message)

	emptyMessage := models.DBMessage{}

	if err != nil {
		return emptyMessage, err
	}

	var newMessage models.DBMessage
	query := bson.M{"_id": res.InsertedID}
	if err = messageCollection.FindOne(ctx, query).Decode(&newMessage); err != nil {
		return emptyMessage, err
	}

	return newMessage, nil
}

func FindMessageByUserId(userId string, messageCollection *mongo.Collection, ctx *gin.Context) ([]*models.DBMessage, error) {

	opt := options.FindOptions{}
	opt.SetSort(bson.M{"ID": -1})

	query := bson.M{"$or": []bson.M{
		{"Sender": userId},
		{"Receiver": userId},
	}}
	cursor, err := messageCollection.Find(ctx, query, &opt)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var messages []*models.DBMessage

	for cursor.Next(ctx) {
		message := &models.DBMessage{}
		err := cursor.Decode(message)

		if err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return []*models.DBMessage{}, nil
	}

	return messages, nil
}
