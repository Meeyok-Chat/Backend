package database

import (
	"context"
	"log"
	"time"

	"github.com/Meeyok-Chat/backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type friendshipRepo struct {
	database *mongo.Collection
}

type FriendshipRepo interface {
	CreateFriendship(userID1, userID2 string) (models.Friendship, error)
	UpdateFriendshipStatus(friendshipID string, status string) (models.Friendship, error)
}

func NewFriendshipRepo(database *mongo.Collection) FriendshipRepo {
	return &friendshipRepo{
		database: database,
	}
}

func (r *friendshipRepo) CreateFriendship(userID1, userID2 string) (models.Friendship, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	friendship := &models.Friendship{
		ID:        primitive.NewObjectID(),
		UserID1:   userID1,
		UserID2:   userID2,
		Status:    models.FriendshipPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := r.database.InsertOne(ctx, friendship)
	if err != nil {
		return models.Friendship{}, err
	}

	return *friendship, nil
}

func (r *friendshipRepo) UpdateFriendshipStatus(friendshipID string, status string) (models.Friendship, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(friendshipID)
	if err != nil {
		log.Fatalf("Invalid ObjectID: %v", err)
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	var updatedFriendship models.Friendship
	err = r.database.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedFriendship)
	if err != nil {
		return models.Friendship{}, err
	}

	return updatedFriendship, nil
}
