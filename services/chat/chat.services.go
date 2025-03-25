package chat

import (
	"errors"

	"github.com/Meeyok-Chat/backend/models"
	"github.com/Meeyok-Chat/backend/repository/database"
)

type chatService struct {
	chatRepo database.ChatRepo
}

type ChatService interface {
	GetChats() ([]models.Chat, error)
	GetChatById(id string, page int, numberOfMessages int) (models.Chat, error)
	GetUserChats(userID string, chatType string) ([]models.Chat, error)
	GetMessages(id string, page int, numberOfMessages int) ([]models.Message, error)
	CreateChat(chat models.Chat) (models.Chat, error)
	AddUsersToChat(chatID string, users []string) error
	UpdateChat(chat models.Chat) error
	DeleteChat(id string) error
	TrimMessages(chat models.Chat) models.Chat
}

func NewChatService(chatRepo database.ChatRepo) ChatService {
	return &chatService{
		chatRepo: chatRepo,
	}
}

func (cs *chatService) GetChats() ([]models.Chat, error) {
	chats, err := cs.chatRepo.GetChats()
	if err != nil {
		return nil, err
	}
	return chats, nil
}

func (cs *chatService) GetChatById(id string, page int, numberOfMessages int) (models.Chat, error) {
	chat, err := cs.chatRepo.GetChatByID(id)
	if err != nil {
		return models.Chat{}, err
	}
	if page == 0 {
		return chat, nil
	}
	chat.Messages = chat.Messages[max(0, len(chat.Messages)-page*numberOfMessages):max(0, len(chat.Messages)-page*numberOfMessages+numberOfMessages)]
	return chat, nil
}

func (cs *chatService) GetMessages(id string, page int, numberOfMessages int) ([]models.Message, error) {
	chat, err := cs.chatRepo.GetChatByID(id)
	if err != nil {
		return nil, err
	}
	if page == 0 {
		return chat.Messages, nil
	}
	messages := chat.Messages[max(0, len(chat.Messages)-page*numberOfMessages):max(0, len(chat.Messages)-(page-1)*numberOfMessages)]
	return messages, nil
}

func (cs *chatService) GetUserChats(userID string, chatType string) ([]models.Chat, error) {
	switch chatType {
	case "group":
		return cs.chatRepo.GetGroupChats(userID)
	case "friend":
		return cs.chatRepo.GetFriendChats(userID)
	case "non-friend":
		return cs.chatRepo.GetNonFriendChats(userID)
	default:
		return nil, errors.New("invalid chat type")
	}
}

func (cs *chatService) CreateChat(chat models.Chat) (models.Chat, error) {
	result, err := cs.chatRepo.CreateChat(chat)
	if err != nil {
		return models.Chat{}, err
	}
	return result, nil
}

func (cs *chatService) AddUsersToChat(chatID string, users []string) error {
	err := cs.chatRepo.AddUsersToChat(chatID, users)
	if err != nil {
		return err
	}
	return nil
}

func (cs *chatService) UpdateChat(chat models.Chat) error {
	err := cs.chatRepo.UpdateChat(chat)
	if err != nil {
		return err
	}
	return nil
}

func (cs *chatService) DeleteChat(id string) error {
	err := cs.chatRepo.DeleteChat(id)
	if err != nil {
		return err
	}
	return nil
}

func (cs *chatService) TrimMessages(chat models.Chat) models.Chat {
	if len(chat.Messages) > 10 {
		chat.Messages = chat.Messages[len(chat.Messages)-10:]
	}
	return chat
}
