package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	Name      string             `json:"name,omitempty" bson:"name"`
	Messages  []Message          `json:"messages"`
	Users     []string           `json:"users,omitempty" bson:"users"`
	Type      string             `json:"type,omitempty" bson:"type"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

type Message struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	From      string             `json:"from"`
	Message   string             `json:"message"`
	CreatedAt time.Time          `json:"createAt"`
}

const (
	IndividualChatType = "Individual"
	GroupChatType      = "Group"
)
