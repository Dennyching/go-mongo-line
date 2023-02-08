/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/practice/golang-line/config"
	"github.com/practice/golang-line/models"
	"github.com/practice/golang-line/services"
	"github.com/spf13/cobra"
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

	userCollection *mongo.Collection

	messageCollection *mongo.Collection

	bot *linebot.Client

	botId string
)

// ? Init function that will run before the "main" function
func init() {

	// ? Load the .env variables
	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("Could not load environment variables", err)
	}

	botId = config.LineBotUserId

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

	router := server.Group("/api")

	router.GET("/healthchecker", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Receive message from line
	router.POST("/webhook", ReceiveMessage)

	// semd message to user by linebot
	router.POST("/sendmessage", SendMessage)

	// get user's message by userId
	router.GET("/message/:userId", GetMessageRequest)

	// get user's message by userId
	router.GET("/userlist/", GetAllUer)

	log.Fatal(server.Run(":" + config.Port))
	return nil

}

func ReceiveMessage(ctx *gin.Context) {
	// Read the incoming message from LINE
	events, err := bot.ParseRequest(ctx.Request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
	}
	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			senderID := event.Source.UserID
			fmt.Println("Sender ID:", senderID)
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				NewMessage := &models.CreateMessageRequest{
					ID:        message.ID,
					Text:      message.Text,
					Sender:    senderID,
					Receiver:  botId,
					CreatedAt: time.Now(),
				}
				services.CreateMessageFunc(NewMessage, messageCollection, ctx)
				_, err := services.FindUserById(senderID, userCollection, ctx)
				if err != nil {
					services.InsertUser(senderID, userCollection, ctx)
				}
			}
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "received"})
}

func SendMessage(ctx *gin.Context) {
	var message models.CreateMessageRequest

	if err := ctx.ShouldBindJSON(&message); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	line_message := linebot.NewTextMessage(message.Text)
	response, err := bot.PushMessage(message.Sender, line_message).Do()
	if err != nil {
		log.Print(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
	}

	message.ID = response.RequestID
	message.Receiver = botId

	message.CreatedAt = time.Now()

	services.CreateMessageFunc(&message, messageCollection, ctx)

	if _, err := services.FindUserById(message.Receiver, userCollection, ctx); err != nil {
		services.InsertUser(message.Receiver, userCollection, ctx)
	}

}

func GetAllUer(ctx *gin.Context) {
	res, err := services.FindAllUser(userCollection, ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": res})

}

func GetMessageRequest(ctx *gin.Context) {
	userId := ctx.Param("userId")
	res, err := services.FindMessageByUserId(userId, messageCollection, ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": res})

}
