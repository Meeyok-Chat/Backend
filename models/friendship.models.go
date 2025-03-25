package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Friendship struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	UserID1   string             `json:"userId1" bson:"userId1"`
	UserID2   string             `json:"userId2" bson:"userId2"`
	Status    string             `json:"status" bson:"status"` // Status 'pending', 'accepted'
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}

const (
	FriendshipPending  = "pending"
	FriendshipAccepted = "accepted"
	FriendshipRejected = "rejected"
)
