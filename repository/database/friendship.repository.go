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
	IsFriends(userID1, userID2 string) (bool, error)
	FindPendingFriendshipBetween(userID1, userID2 string) (models.Friendship, error)
	GetFriendshipsByStatus(userID, status string) ([]models.Friendship, error)
	CreateFriendship(userID1, userID2 string) (models.Friendship, error)
	UpdateFriendshipStatus(friendshipID string, status string) (models.Friendship, error)
}

func NewFriendshipRepo(database *mongo.Collection) FriendshipRepo {
	return &friendshipRepo{
		database: database,
	}
}

func (s *friendshipRepo) IsFriends(userID1, userID2 string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"userId1": userID1, "userId2": userID2, "status": "accepted"},
			{"userId1": userID2, "userId2": userID1, "status": "accepted"},
		},
	}

	var friendship struct{}
	err := s.database.FindOne(ctx, filter).Decode(&friendship)

	if err == mongo.ErrNoDocuments {
		return false, nil // Not friends
	} else if err != nil {
		return false, err // Other database errors
	}

	return true, nil // They are friends
}

func (r *friendshipRepo) FindPendingFriendshipBetween(userID1, userID2 string) (models.Friendship, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			// {"userId1": userID1, "userId2": userID2},
			{"userId1": userID2, "userId2": userID1},
		},
		"status": models.FriendshipPending,
	}

	var friendship models.Friendship
	err := r.database.FindOne(ctx, filter).Decode(&friendship)
	if err != nil {
		return models.Friendship{}, err
	}
	return friendship, nil
}

func (r *friendshipRepo) GetFriendshipsByStatus(userID, status string) ([]models.Friendship, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"userId1": userID},
					{"userId2": userID},
				},
			},
			{"status": status},
		},
	}

	cursor, err := r.database.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var friendships []models.Friendship
	if err = cursor.All(ctx, &friendships); err != nil {
		return nil, err
	}

	return friendships, nil
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
