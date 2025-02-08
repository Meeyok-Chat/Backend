package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	Email     string             `json:"email,omitempty" bson:"email"`
	Username  string             `json:"username" bson:"username"`
	Role      string             `json:"role,omitempty" bson:"role"`
	Chats     []string           `json:"chats,omitempty" bson:"chats,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedat"`
}
