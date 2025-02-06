package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	Username  string             `json:"username,omitempty" bson:"username"`
	Role      string             `json:"role,omitempty" bson:"role"`
	Chats     []string           `json:"chats,omitempty" bson:"chats,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt"`
}
