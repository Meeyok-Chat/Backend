package database

import (
	"context"
	"errors"
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
	chatDb       *mongo.Collection
	backupChatDb *mongo.Collection
}

type ChatRepo interface {
	NewChat(id primitive.ObjectID) models.Chat
	NewProvidePromptData() models.ProvidePromptData
	NewFeedback() models.Feedback
	CreateFeedback(id primitive.ObjectID, msg string, score int) models.Feedback
	NewMessage(role string, msg string, phase string, reasoning string) models.Message
	UploadChat(chat models.Chat) error
	UpdateChatFeedback(id primitive.ObjectID, feedback models.Feedback) error
	UpdateMessageFeedback(id primitive.ObjectID, feedback models.Feedback) error
	BackupChats() error
	GetChats() ([]models.Chat, error)
	GetChatById(id primitive.ObjectID) (models.Chat, error)
	GetChatsByBatchId(batchId string) ([]models.Chat, error)
	CreateChat(chat models.Chat) (models.Chat, error)
	UpdateChat(chat models.Chat) error
	UpdateChatStatus(id primitive.ObjectID, status string, startTime time.Time) error
	UpdatePromptVersion(id primitive.ObjectID, promptVersion int) error
	DeleteChat(id primitive.ObjectID) error
	GetBackupChats() ([]models.BackupChat, error)
	// Summary sheet
	GetSummarysheetByChatId(id primitive.ObjectID) (models.SummarySheet, error)
	UpdateSummarysheet(id primitive.ObjectID, summarySheet models.SummarySheet) error
}

func NewChatRepo(chatDb *mongo.Collection, backupChatDb *mongo.Collection) ChatRepo {
	return &chatRepo{
		chatDb:       chatDb,
		backupChatDb: backupChatDb,
	}
}

func (r *chatRepo) NewChat(id primitive.ObjectID) models.Chat {
	chat := models.Chat{
		ID:                id,
		Messages:          []models.Message{},
		Feedback:          models.Feedback{},
		ProvidePromptData: models.ProvidePromptData{},
		UpdatedAt:         time.Now(),
	}

	chat.Feedback = r.NewFeedback()
	chat.ProvidePromptData = r.NewProvidePromptData()
	return chat
}

func (r *chatRepo) NewProvidePromptData() models.ProvidePromptData {
	providePromptData := models.ProvidePromptData{
		SummaryState:    false,
		Turn:            0,
		TimeLastMessage: time.Now(),
	}
	return providePromptData
}

func (r *chatRepo) NewFeedback() models.Feedback {
	feedback := models.Feedback{
		ID:      primitive.NewObjectID(),
		Status:  false,
		Message: "",
		Score:   0,
	}
	return feedback
}

func (r *chatRepo) CreateFeedback(id primitive.ObjectID, msg string, score int) models.Feedback {
	feedback := models.Feedback{
		ID:      id,
		Status:  true,
		Message: msg,
		Score:   score,
	}
	return feedback
}

func (r *chatRepo) NewMessage(role string, msg string, phase string, reasoing string) models.Message {
	message := models.Message{
		ID:        primitive.NewObjectID(),
		Role:      role,
		Message:   msg,
		Phase:     phase,
		Reasoning: reasoing,
		CreatedAt: time.Now(),
		Feedback:  r.NewFeedback(),
	}
	return message
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
		"$set": bson.M{
			"providepromptdata": chat.ProvidePromptData,
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

func (r *chatRepo) UpdateChatFeedback(id primitive.ObjectID, feedback models.Feedback) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updateFields := bson.M{}
	if feedback.Message != "" {
		updateFields["feedback.message"] = feedback.Message
	}
	if feedback.Score != 0 {
		updateFields["feedback.score"] = int(feedback.Score)
	}

	if len(updateFields) == 0 {
		return errors.New("no fields to update")
	}

	filter := bson.M{"_id": id}
	updateFields["feedback.status"] = true
	update := bson.M{"$set": updateFields}
	_, err := r.chatDb.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("failed to update feedback")
	}
	return nil
}

func (r *chatRepo) UpdateMessageFeedback(id primitive.ObjectID, feedback models.Feedback) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updateFields := bson.M{}
	if feedback.Message != "" {
		updateFields["messages.$[msg].feedback.message"] = feedback.Message
	}
	if feedback.Score != 0 {
		updateFields["messages.$[msg].feedback.score"] = int(feedback.Score)
	}

	if len(updateFields) == 0 {
		return errors.New("no fields to update")
	}

	filter := bson.M{"_id": id}
	updateFields["messages.$[msg].feedback.status"] = true
	update := bson.M{
		"$set": updateFields,
	}

	arrayFilters := bson.A{
		bson.M{"msg.feedback._id": feedback.ID},
	}

	options := options.Update().SetArrayFilters(options.ArrayFilters{Filters: arrayFilters})

	result, err := r.chatDb.UpdateOne(ctx, filter, update, options)
	if err != nil {
		return errors.New("failed to update feedback in database")
	}
	if result.ModifiedCount == 0 {
		return errors.New("no messages to update")
	}
	return nil
}

func (r *chatRepo) BackupChats() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chats, err := r.GetChats()
	if err != nil {
		return err
	}

	_, err = r.backupChatDb.InsertOne(ctx, models.BackupChat{
		ID:        primitive.NewObjectID(),
		Chats:     chats,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return err
	}
	return nil
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

func (r *chatRepo) GetChatById(id primitive.ObjectID) (models.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	u := models.Chat{}

	filter := bson.M{"_id": id}
	err := r.chatDb.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		return models.Chat{}, err
	}
	return u, nil
}

func (r *chatRepo) GetChatsByBatchId(batchId string) ([]models.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"batchid": batchId}
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

func (r *chatRepo) CreateChat(chat models.Chat) (models.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if chat.ID.IsZero() {
		chat.ID = primitive.NewObjectID()
	}
	chat.UpdatedAt = time.Now()
	_, err := r.chatDb.InsertOne(ctx, chat)
	if err != nil {
		return models.Chat{}, err
	}
	return chat, nil
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

func (r *chatRepo) UpdateChatStatus(id primitive.ObjectID, status string, startTime time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updateFields := bson.M{"status": status}
	if status == models.ChatProcessing {
		if startTime.IsZero() {
			updateFields["starttime"] = time.Now()
		} else {
			updateFields["starttime"] = startTime
		}
	}

	if len(updateFields) == 0 {
		return errors.New("no fields to update")
	}

	filter := bson.M{"_id": id}
	update := bson.M{"$set": updateFields}
	_, err := r.chatDb.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (r *chatRepo) UpdatePromptVersion(id primitive.ObjectID, promptVersion int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"promptversion": promptVersion}}
	_, err := r.chatDb.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (r *chatRepo) DeleteChat(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	result, err := r.chatDb.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("no chat found to delete")
	}
	return nil
}

func (r *chatRepo) GetBackupChats() ([]models.BackupChat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	cursor, err := r.backupChatDb.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	results := []models.BackupChat{}
	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *chatRepo) GetSummarysheetByChatId(id primitive.ObjectID) (models.SummarySheet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	u := models.Chat{}

	filter := bson.M{"_id": id}
	err := r.chatDb.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		return models.SummarySheet{}, err
	}
	return u.SummarySheet, nil
}

func (r *chatRepo) UpdateSummarysheet(id primitive.ObjectID, summarySheet models.SummarySheet) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"summarySheet": summarySheet}}
	_, err := r.chatDb.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}
