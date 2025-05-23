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
)

type userRepo struct {
	database *mongo.Collection
}

type UserRepo interface {
	GetUsers() ([]models.User, error)
	GetUserByID(id string) (models.User, error)
	GetUsersByIDs(userIDs []string) ([]models.User, error)
	GetUserByEmail(email string) (models.User, error)
	GetUserByUsername(username string) (models.User, error)

	CreateUser(user models.User) error

	AddChatToUser(userID string, chatID string) error
	AddPostToUser(userID string, postID string) error

	UpdateUser(user models.User) error
	UpdateUsername(userID string, newUsername string) error

	DeleteUser(id primitive.ObjectID) error
}

func NewUserRepo(database *mongo.Collection) UserRepo {
	return &userRepo{
		database: database,
	}
}

func (r *userRepo) GetUsers() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	cursor, err := r.database.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	results := []models.User{}
	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *userRepo) GetUserByID(id string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatalf("Invalid ObjectID: %v", err)
	}

	u := models.User{}
	filter := bson.M{"_id": objID}
	err = r.database.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		return models.User{}, err
	}
	return u, nil
}

func (r *userRepo) GetUsersByIDs(userIDs []string) ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectIDs := []primitive.ObjectID{}
	for _, id := range userIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		objectIDs = append(objectIDs, objID)
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	cursor, err := r.database.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepo) GetUserByEmail(email string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	u := models.User{}

	filter := bson.M{"email": email}
	err := r.database.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		return models.User{}, err
	}
	return u, nil
}

func (r *userRepo) GetUserByUsername(username string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	u := models.User{}

	filter := bson.M{"username": username}
	err := r.database.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		return models.User{}, err
	}
	return u, nil
}

func (r *userRepo) CreateUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if a user with the same username already exists
	filter := bson.M{"email": user.Email}
	existingUser := models.User{}
	err := r.database.FindOne(ctx, filter).Decode(&existingUser)
	if err == nil {
		return fmt.Errorf("email '%s' is already taken", user.Email)
	}

	insertData := bson.M{
		"_id":       primitive.NewObjectID(),
		"email":     user.Email,
		"username":  user.Username,
		"role":      user.Role,
		"chats":     []string{},
		"friends":   []string{},
		"updatedat": time.Now(),
	}
	log.Println(insertData)
	_, err = r.database.InsertOne(ctx, insertData)
	if err != nil {
		return err
	}

	var insertedUser bson.M
	err = r.database.FindOne(ctx, bson.M{"_id": insertData["_id"]}).Decode(&insertedUser)
	if err != nil {
		log.Println("FindOne error:", err)
	} else {
		log.Printf("After insert: %+v\n", insertedUser)
	}

	return nil
}

func (r *userRepo) AddChatToUser(userID string, chatID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Fatalf("Invalid ObjectID: %v", err)
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$addToSet": bson.M{"chats": chatID},
		"$set":      bson.M{"updatedAt": time.Now()},
	}

	result, err := r.database.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("no user found or chat already exists")
	}

	return nil
}

func (r *userRepo) AddPostToUser(userID string, postID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	userFilter := bson.M{"_id": objID}
	update := bson.M{"$push": bson.M{"posts": postID}}
	_, err = r.database.UpdateOne(ctx, userFilter, update)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepo) UpdateUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user.UpdatedAt = time.Now()
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}
	result, err := r.database.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf("no user found to update")
	}
	return nil
}

func (r *userRepo) UpdateUsername(userID string, newUsername string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Fatalf("Invalid ObjectID: %v", err)
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"username":  newUsername,
			"updatedAt": time.Now(),
		},
	}

	_, err = r.database.UpdateOne(ctx, filter, update)
	return err
}

func (r *userRepo) DeleteUser(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	result, err := r.database.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("no user found to delete")
	}
	return nil
}
