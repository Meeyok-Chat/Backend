package friendship

import (
	"fmt"

	"github.com/Meeyok-Chat/backend/models"
	"github.com/Meeyok-Chat/backend/repository/database"
)

type friendshipService struct {
	friendshipRepo database.FriendshipRepo
}

type FriendshipService interface {
	AddFriendship(userID1, userID2 string) (models.Friendship, error)
	AcceptFriendship(friendshipID string) (models.Friendship, error)
	RejectFriendship(friendshipID string) (models.Friendship, error)
}

func NewFriendshipService(friendshipRepo database.FriendshipRepo) FriendshipService {
	return &friendshipService{
		friendshipRepo: friendshipRepo,
	}
}

func (s *friendshipService) AddFriendship(userID1, userID2 string) (models.Friendship, error) {
	if userID1 == userID2 {
		return models.Friendship{}, fmt.Errorf("cannot send friend request to yourself")
	}

	return s.friendshipRepo.CreateFriendship(userID1, userID2)
}

func (s *friendshipService) AcceptFriendship(friendshipID string) (models.Friendship, error) {
	return s.friendshipRepo.UpdateFriendshipStatus(friendshipID, models.FriendshipAccepted)
}

func (s *friendshipService) RejectFriendship(friendshipID string) (models.Friendship, error) {
	return s.friendshipRepo.UpdateFriendshipStatus(friendshipID, models.FriendshipRejected)
}
