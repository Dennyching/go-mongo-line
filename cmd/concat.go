/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/practice/golang-line/config"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// concatCmd represents the concat command
var (
	// #TODO Have a Flag for GUI Mode and API Server Mode
	MonoCmd = &cobra.Command{
		Use:     "mono",
		Short:   "start",
		Long:    "start",
		Aliases: []string{"a"},
		Example: "app some_file.cfg",
		RunE:    execute_service,
	}
)

// ? Create required variables that we'll re-assign later
var (
	server      *gin.Engine
	ctx         context.Context
	mongoclient *mongo.Client
	redisclient *redis.Client

	userCollection *mongo.Collection

	messageCollection *mongo.Collection

	bot *linebot.Client
)

type CreateMessageRequest struct {
	ID        string    `json:"ID" bson:"ID" binding:"required"`
	Text      string    `json:"Text" bson:"Text" binding:"required"`
	Sender    string    `json:"Sender" bson:"Sender"`
	Receiver  string    `json:"Receiver" bson:"Receiver" binding:"required"`
	CreatedAt time.Time `json:"Created_at" bson:"Created_at"`
}

type DBMessage struct {
	ID        string    `json:"ID" bson:"ID" binding:"required"`
	Text      string    `json:"Text" bson:"Text" binding:"required"`
	Sender    string    `json:"Sender" bson:"Sender"`
	Receiver  string    `json:"Receiver" bson:"Receiver" binding:"required"`
	CreatedAt time.Time `json:"Created_at" bson:"Created_at"`
}

type User struct {
	ID string `json:"ID" bson:"ID" binding:"required"`
}

// ? Init function that will run before the "main" function
func init() {

	// ? Load the .env variables
	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("Could not load environment variables", err)
	}

	// ? Create a context
	ctx = context.TODO()

	// ? Connect to MongoDB
	mongoconn := options.Client().ApplyURI(config.DBUri)
	mongoclient, err := mongo.Connect(ctx, mongoconn)

	if err != nil {
		panic(err)
	}

	if err := mongoclient.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

	fmt.Println("MongoDB successfully connected...")

	// ? Connect to Redis
	redisclient = redis.NewClient(&redis.Options{
		Addr: config.RedisUri,
	})

	if _, err := redisclient.Ping(ctx).Result(); err != nil {
		panic(err)
	}

	err = redisclient.Set(ctx, "test", "Welcome to Golang with Redis and MongoDB", 0).Err()
	if err != nil {
		panic(err)
	}

	fmt.Println("Redis client connected successfully...")

	// Collections
	userCollection = mongoclient.Database("golang_mongodb").Collection("users")

	// 👇 Instantiate the Constructors
	messageCollection = mongoclient.Database("golang_mongodb").Collection("message")

	bot, err = linebot.New(
		config.LineChannelSecret,
		config.LineChannelAccessToken,
	)
	if err != nil {
		log.Fatal(err)
	}

	// ? Create the Gin Engine instance
	server = gin.Default()
}

func execute_service(cmd *cobra.Command, args []string) error {
	config, err := config.LoadConfig(".")

	if err != nil {
		log.Fatal("Could not load config", err)
		return err
	}

	defer mongoclient.Disconnect(ctx)

	value, err := redisclient.Get(ctx, "test").Result()

	if err == redis.Nil {
		fmt.Println("key: test does not exist")
	} else if err != nil {
		return err
	}

	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				senderID := event.Source.UserID
				fmt.Println("Sender ID:", senderID)
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					NewMessage := &CreateMessageRequest{
						ID:        message.ID,
						Text:      message.Text,
						Sender:    senderID,
						Receiver:  config.LineBotUserId,
						CreatedAt: time.Now(),
					}
					log.Printf("User sent a text message: %s", message.Text)
					log.Printf("Id: %s", message.ID)
					CreateMessageFunc(NewMessage, messageCollection)
					_, err := FindUserById(senderID)
					if err != nil {
						InsertUser(senderID)
					}
				}
			}
		}
	})

	router := server.Group("/api")
	router.GET("/healthchecker", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": value})
	})

	router.POST("/sendmessage", SendMessage)

	log.Fatal(server.Run(":" + config.Port))
	return nil

}

func SendMessage(ctx *gin.Context) {
	var message CreateMessageRequest

	if err := ctx.ShouldBindJSON(&message); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	CreateMessageFunc(&message, messageCollection)

	_, err := FindUserById(message.Receiver)
	if err != nil {
		InsertUser(message.Receiver)
	}

	line_message := linebot.NewTextMessage(message.Text)
	if _, err := bot.PushMessage(message.Receiver, line_message).Do(); err != nil {
		log.Print(err)
	}

}

func CreateMessageFunc(post *CreateMessageRequest, postCollection *mongo.Collection) (DBMessage, error) {

	res, err := postCollection.InsertOne(ctx, post)

	emptyPost := DBMessage{}

	if err != nil {
		if er, ok := err.(mongo.WriteException); ok && er.WriteErrors[0].Code == 11000 {
			return emptyPost, errors.New("post with that title already exists")
		}
		return emptyPost, err
	}

	var newPost DBMessage
	query := bson.M{"_id": res.InsertedID}
	if err = postCollection.FindOne(ctx, query).Decode(&newPost); err != nil {
		return emptyPost, err
	}

	return newPost, nil
}

func FindUserById(userId string) (*User, error) {
	var user *User

	query := bson.M{"ID": strings.ToLower(userId)}
	err := userCollection.FindOne(ctx, query).Decode(&user)

	if err != nil {
		return &User{}, err
	}

	return user, nil
}

func InsertUser(userId string) (*User, error) {
	user := &User{
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
		return nil, errors.New("could not create index for email")
	}
	var newUser *User
	query := bson.M{"_id": res.InsertedID}

	err = userCollection.FindOne(ctx, query).Decode(&newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func FindAllUser(page int, limit int) ([]*User, error) {
	if page == 0 {
		page = 1
	}

	if limit == 0 {
		limit = 10
	}

	skip := (page - 1) * limit

	opt := options.FindOptions{}
	opt.SetLimit(int64(limit))
	opt.SetSkip(int64(skip))
	opt.SetSort(bson.M{"ID": -1})

	query := bson.M{}
	cursor, err := userCollection.Find(ctx, query, &opt)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var users []*User

	for cursor.Next(ctx) {
		post := &User{}
		err := cursor.Decode(post)

		if err != nil {
			return nil, err
		}

		users = append(users, post)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return []*User{}, nil
	}

	return users, nil
}

func FindMessageByUserId(page int, limit int, userId string) ([]*User, error) {
	if page == 0 {
		page = 1
	}

	if limit == 0 {
		limit = 10
	}

	skip := (page - 1) * limit

	opt := options.FindOptions{}
	opt.SetLimit(int64(limit))
	opt.SetSkip(int64(skip))
	opt.SetSort(bson.M{"ID": -1})

	query := bson.M{"$or": []bson.M{
		{"Sender": strings.ToLower(userId)},
		{"Receiver": strings.ToLower(userId)},
	}}
	cursor, err := messageCollection.Find(ctx, query, &opt)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var users []*User

	for cursor.Next(ctx) {
		post := &User{}
		err := cursor.Decode(post)

		if err != nil {
			return nil, err
		}

		users = append(users, post)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return []*User{}, nil
	}

	return users, nil
}
