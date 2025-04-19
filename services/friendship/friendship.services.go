package friendship

import (
	"fmt"

	"github.com/Meeyok-Chat/backend/models"
	"github.com/Meeyok-Chat/backend/repository/database"
)

type friendshipService struct {
	userRepo       database.UserRepo
	friendshipRepo database.FriendshipRepo
}

type FriendshipService interface {
	GetFriends(userID, status string) ([]models.User, error)
	AddFriendship(userID1, userID2 string) (models.Friendship, error)
	UpdateFriendshipStatus(userID, friendID, status string) (models.Friendship, error)
}

func NewFriendshipService(friendshipRepo database.FriendshipRepo, userRepo database.UserRepo) FriendshipService {
	return &friendshipService{
		friendshipRepo: friendshipRepo,
		userRepo:       userRepo,
	}
}

func (s *friendshipService) IsFriends(userID1, userID2 string) (bool, error) {
	return s.friendshipRepo.IsFriends(userID1, userID2)
}

func (s *friendshipService) GetFriends(userID, status string) ([]models.User, error) {
	friendships, err := s.friendshipRepo.GetFriendshipsByStatus(userID, status)
	if err != nil {
		return nil, err
	}

	friendIDs := []string{}
	for _, f := range friendships {
		if f.UserID1 == userID {
			friendIDs = append(friendIDs, f.UserID2)
		} else {
			friendIDs = append(friendIDs, f.UserID1)
		}
	}

	return s.userRepo.GetUsersByIDs(friendIDs)
}

func (s *friendshipService) AddFriendship(userID1, userID2 string) (models.Friendship, error) {
	if userID1 == userID2 {
		return models.Friendship{}, fmt.Errorf("cannot send friend request to yourself")
	}

	return s.friendshipRepo.CreateFriendship(userID1, userID2)
}

func (s *friendshipService) UpdateFriendshipStatus(userID, friendID, status string) (models.Friendship, error) {
	friendship, err := s.friendshipRepo.FindPendingFriendshipBetween(userID, friendID)
	if err != nil {
		return models.Friendship{}, err
	}
	if friendship.Status != models.FriendshipPending {
		return models.Friendship{}, fmt.Errorf("friendship is not in pending status")
	}

	return s.friendshipRepo.UpdateFriendshipStatus(friendship.ID.Hex(), models.FriendshipAccepted)
}
