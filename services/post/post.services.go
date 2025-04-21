package post

import (
	"github.com/Meeyok-Chat/backend/dtos"
	"github.com/Meeyok-Chat/backend/models"
	"github.com/Meeyok-Chat/backend/repository/database"
)

type postService struct {
	postRepo database.PostRepo
	userRepo database.UserRepo
}

type PostService interface {
	GetPosts() ([]models.Post, error)
	GetPostByID(id string) (*models.Post, error)
	CreatePost(userId string, post dtos.CreatePostRequest) (*models.Post, error)
	AddPostToUser(userID string, postID string) error
	UpdatePost(id string, post models.Post) error
	DeletePost(id string) error
}

func NewPostService(postRepo database.PostRepo, userRepo database.UserRepo) PostService {
	return &postService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

func (s *postService) GetPosts() ([]models.Post, error) {
	return s.postRepo.GetPosts()
}

func (s *postService) GetPostByID(id string) (*models.Post, error) {
	return s.postRepo.GetPostByID(id)
}

func (s *postService) CreatePost(userId string, postDto dtos.CreatePostRequest) (*models.Post, error) {
	post := &models.Post{
		UserID:  userId,
		Title:   postDto.Title,
		Content: postDto.Content,
	}
	return s.postRepo.CreatePost(*post)
}

func (s *postService) AddPostToUser(userID string, postID string) error {
	return s.userRepo.AddPostToUser(userID, postID)
}

func (s *postService) UpdatePost(id string, post models.Post) error {
	return s.postRepo.UpdatePost(id, post)
}

func (s *postService) DeletePost(id string) error {
	return s.postRepo.DeletePost(id)
}
