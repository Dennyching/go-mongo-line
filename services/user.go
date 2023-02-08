package services

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/practice/golang-line/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func FindUserById(userId string, userCollection *mongo.Collection, ctx *gin.Context) (*models.User, error) {
	var user *models.User

	query := bson.M{"ID": strings.ToLower(userId)}
	err := userCollection.FindOne(ctx, query).Decode(&user)

	if err != nil {
		return &models.User{}, err
	}

	return user, nil
}

func InsertUser(userId string, userCollection *mongo.Collection, ctx *gin.Context) (*models.User, error) {
	user := &models.User{
		ID: userId,
	}
	res, err := userCollection.InsertOne(ctx, &user)
	if err != nil {
		return nil, err
	}
	// Create a unique index for the ID field
	opt := options.Index()
	opt.SetUnique(true)
	index := mongo.IndexModel{Keys: bson.M{"ID": 1}, Options: opt}

	if _, err := userCollection.Indexes().CreateOne(ctx, index); err != nil {
		return nil, errors.New("could not create index for user")
	}
	var newUser *models.User
	query := bson.M{"_id": res.InsertedID}

	err = userCollection.FindOne(ctx, query).Decode(&newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func FindAllUser(userCollection *mongo.Collection, ctx *gin.Context) ([]*models.User, error) {

	opt := options.FindOptions{}
	opt.SetSort(bson.M{"ID": -1})

	query := bson.M{}
	cursor, err := userCollection.Find(ctx, query, &opt)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var users []*models.User

	for cursor.Next(ctx) {
		user := &models.User{}
		err := cursor.Decode(user)

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return []*models.User{}, nil
	}

	return users, nil
}
