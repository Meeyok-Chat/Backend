package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Meeyok-Chat/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type chatRepo struct {
	chatDb *mongo.Collection
	userDb *mongo.Collection
}

type ChatRepo interface {
	// New
	NewChat(id primitive.ObjectID) models.Chat
	NewMessage(msg string, from string) models.Message

	// Get
	GetChats() ([]models.Chat, error)
	GetChatByID(id string) (models.Chat, error)
	GetGroupChats(userID string) ([]models.Chat, error)
	GetFriendChats(userID string) ([]models.Chat, error)
	GetNonFriendChats(userID string) ([]models.Chat, error)

	// Create
	CreateChat(chat models.Chat) (models.Chat, error)

	// Manage
	AddUsersToChat(chatID string, users []string) error
	AppendMessage(chatID string, message models.Message) error
	UploadChat(chat models.Chat) error

	// Update
	UpdateChat(chat models.Chat) error

	// Delete
	DeleteChat(id string) error
}

func NewChatRepo(chatDb *mongo.Collection, userDb *mongo.Collection) ChatRepo {
	return &chatRepo{
		chatDb: chatDb,
		userDb: userDb,
	}
}

func (r *chatRepo) NewChat(id primitive.ObjectID) models.Chat {
	chat := models.Chat{
		ID:        id,
		Messages:  []models.Message{},
		UpdatedAt: time.Now(),
	}

	return chat
}

func (r *chatRepo) NewMessage(msg string, from string) models.Message {
	message := models.Message{
		ID:        primitive.NewObjectID(),
		From:      from,
		Message:   msg,
		CreatedAt: time.Now(),
	}
	return message
}

func (r *chatRepo) GetChats() ([]models.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	cursor, err := r.chatDb.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	results := []models.Chat{}
	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *chatRepo) GetChatByID(id string) (models.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatalf("Invalid ObjectID: %v", err)
	}

	chat := models.Chat{}
	filter := bson.M{"_id": objID}
	err = r.chatDb.FindOne(ctx, filter).Decode(&chat)
	if err != nil {
		return models.Chat{}, err
	}
	return chat, nil
}

func (r *chatRepo) GetGroupChats(userID string) ([]models.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var chats []models.Chat
	filter := bson.M{"users": userID, "type": models.GroupChatType}
	cursor, err := r.chatDb.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &chats); err != nil {
		return nil, err
	}
	return chats, nil
}

func (r *chatRepo) GetFriendChats(userID string) ([]models.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Fatalf("Invalid ObjectID: %v", err)
	}

	var user models.User
	err = r.userDb.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	var chats []models.Chat
	filter := bson.M{
		"users": bson.M{"$all": user.Friends},
		"type":  models.IndividualChatType,
	}

	cursor, err := r.chatDb.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &chats); err != nil {
		return nil, err
	}
	return chats, nil
}

func (r *chatRepo) GetNonFriendChats(userID string) ([]models.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Fatalf("Invalid ObjectID: %v", err)
	}

	var user models.User
	err = r.userDb.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	var chats []models.Chat
	filter := bson.M{
		"users": bson.M{"$in": []string{userID}, "$nin": user.Friends},
		"type":  models.IndividualChatType,
	}

	cursor, err := r.chatDb.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &chats); err != nil {
		return nil, err
	}
	return chats, nil
}

func (r *chatRepo) CreateChat(chat models.Chat) (models.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chat.ID = primitive.NewObjectID()
	chat.Messages = make([]models.Message, 0)
	chat.UpdatedAt = time.Now()
	_, err := r.chatDb.InsertOne(ctx, chat)
	if err != nil {
		return models.Chat{}, err
	}
	return chat, nil
}

func (r *chatRepo) AddUsersToChat(chatID string, users []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		log.Fatalf("Invalid ObjectID: %v", err)
	}

	// Check chat type
	var chat models.Chat
	err = r.chatDb.FindOne(ctx, bson.M{"_id": objID}).Decode(&chat)
	if err != nil {
		return fmt.Errorf("chat not found: %v", err)
	}

	if chat.Type == models.IndividualChatType {
		return fmt.Errorf("cannot add users to Individual chat")
	}

	// Add new users to group chat
	filter := bson.M{"_id": objID}
	update := bson.M{
		"$addToSet": bson.M{"users": bson.M{"$each": users}},
		"$set":      bson.M{"updatedAt": time.Now()},
	}

	result, err := r.chatDb.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf("chat not found")
	}
	return nil
}

func (r *chatRepo) AppendMessage(chatID string, message models.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		log.Fatalf("Invalid ObjectID: %v", err)
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$push": bson.M{
			"messages": message,
		},
	}
	updateOptions := options.Update().SetUpsert(true)
	_, err = r.chatDb.UpdateOne(ctx, filter, update, updateOptions)
	// log.Println("Uploading chat ", chatID)
	if err != nil {
		return err
	}
	return nil
}

func (r *chatRepo) UploadChat(chat models.Chat) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": chat.ID}
	update := bson.M{
		"$push": bson.M{
			"messages": bson.M{
				"$each": chat.Messages,
			},
		},
	}
	updateOptions := options.Update().SetUpsert(true)
	_, err := r.chatDb.UpdateOne(ctx, filter, update, updateOptions)
	log.Println("Uploading chat ", chat.ID.Hex())
	if err != nil {
		return err
	}
	return nil
}

func (r *chatRepo) UpdateChat(chat models.Chat) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chat.UpdatedAt = time.Now()
	filter := bson.M{"_id": chat.ID}
	update := bson.M{"$set": chat}
	_, err := r.chatDb.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (r *chatRepo) DeleteChat(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatalf("Invalid ObjectID: %v", err)
	}

	filter := bson.M{"_id": objID}
	result, err := r.chatDb.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("no chat found to delete")
	}
	return nil
}
