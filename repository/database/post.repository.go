package database

import (
	"context"
	"time"

	"github.com/Meeyok-Chat/backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type postRepo struct {
	database *mongo.Collection
}

type PostRepo interface {
	GetPosts() ([]models.Post, error)
	GetPostByID(id string) (*models.Post, error)
	CreatePost(post models.Post) (*models.Post, error)
	UpdatePost(id string, post models.Post) error
	DeletePost(id string) error
}

func NewPostRepo(database *mongo.Collection) PostRepo {
	return &postRepo{database: database}
}

func (r *postRepo) GetPosts() ([]models.Post, error) {
	var posts []models.Post
	cursor, err := r.database.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.Background(), &posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepo) GetPostByID(id string) (*models.Post, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var post models.Post
	err = r.database.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&post)
	return &post, err
}

func (r *postRepo) CreatePost(post models.Post) (*models.Post, error) {
	post.ID = primitive.NewObjectID()
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()

	_, err := r.database.InsertOne(context.TODO(), post)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepo) UpdatePost(id string, post models.Post) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update := bson.M{
		"$set": bson.M{
			"content":   post.Content,
			"updatedAt": time.Now(),
		},
	}
	_, err = r.database.UpdateOne(context.TODO(), bson.M{"_id": objID}, update)
	return err
}

func (r *postRepo) DeletePost(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.database.DeleteOne(context.TODO(), bson.M{"_id": objID})
	return err
}
