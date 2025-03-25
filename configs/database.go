package configs

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoClient struct {
	Client     *mongo.Client
	User       *mongo.Collection
	Chat       *mongo.Collection
	Friendship *mongo.Collection
	Post       *mongo.Collection
}

func NewMongoClient() (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI(GetEnv("MONGODB_URI")),
	)
	if err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}

	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("ping mongodb error: %w", err)
	}
	fmt.Println("ping mongo success")
	return &MongoClient{
		Client:     mongoClient,
		User:       mongoClient.Database("Golang").Collection("users"),
		Chat:       mongoClient.Database("Golang").Collection("chats"),
		Friendship: mongoClient.Database("Golang").Collection("friendships"),
		Post:       mongoClient.Database("Golang").Collection("posts"),
	}, nil
}
