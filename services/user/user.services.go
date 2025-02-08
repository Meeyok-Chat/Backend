package user

import (
	"fmt"

	"github.com/Meeyok-Chat/backend/models"
	"github.com/Meeyok-Chat/backend/repository/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type userService struct {
	userRepo database.UserRepo
}

type UserService interface {
	GetUsers() ([]models.User, error)
	GetUserByID(id string) (models.User, error)
	GetUserByEmail(email string) (models.User, error)
	GetUserByUsername(username string) (models.User, error)

	CreateUser(user models.User) error

	AddChatToUser(userIDs []string, chatID string) error

	UpdateUser(user models.User) error
	UpdateUsername(userID string, newUsername string) error

	DeleteUser(id primitive.ObjectID) error
}

func NewUserService(userRepo database.UserRepo) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (us userService) GetUsers() ([]models.User, error) {
	result, err := us.userRepo.GetUsers()
	if err != nil {
		return []models.User{}, err
	}
	return result, nil
}

func (us userService) GetUserByID(id string) (models.User, error) {
	result, err := us.userRepo.GetUserByID(id)
	if err != nil {
		return models.User{}, err
	}
	return result, nil
}

func (us userService) GetUserByEmail(email string) (models.User, error) {
	result, err := us.userRepo.GetUserByEmail(email)
	if err != nil {
		return models.User{}, err
	}
	return result, nil
}

func (us userService) GetUserByUsername(username string) (models.User, error) {
	result, err := us.userRepo.GetUserByUsername(username)
	if err != nil {
		return models.User{}, err
	}
	return result, nil
}

func (us userService) AddChatToUser(userIDs []string, chatID string) error {
	for _, userID := range userIDs {
		err := us.userRepo.AddChatToUser(userID, chatID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (us userService) CreateUser(user models.User) error {
	err := us.userRepo.CreateUser(user)
	if err != nil {
		return err
	}
	return nil
}

func (us userService) UpdateUser(user models.User) error {
	err := us.userRepo.UpdateUser(user)
	if err != nil {
		return err
	}
	return nil
}

func (us userService) UpdateUsername(userID string, newUsername string) error {
	if newUsername == "" {
		return fmt.Errorf("username cannot be empty")
	}

	return us.userRepo.UpdateUsername(userID, newUsername)
}

func (us userService) DeleteUser(id primitive.ObjectID) error {
	err := us.userRepo.DeleteUser(id)
	if err != nil {
		return err
	}
	return nil
}
