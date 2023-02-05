/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/practice/golang-line/config"
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
	redisclient *redis.Client
)

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

	router := server.Group("/api")
	router.GET("/healthchecker", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": value})
	})

	log.Fatal(server.Run(":" + config.Port))
	return nil

}
