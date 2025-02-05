package chat

import (
	"crypto/aes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/Meeyok-Chat/backend/configs"
	"github.com/Meeyok-Chat/backend/models"
	"github.com/Meeyok-Chat/backend/repository/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type chatService struct {
	chatRepo database.ChatRepo
}

type ChatService interface {
	GetChats() ([]models.Chat, error)
	GetChatById(id primitive.ObjectID, page int, numberOfMessages int) (models.Chat, error)
	GetChatsByBatchId(batchId string) ([]models.Chat, error)
	GetMessages(id primitive.ObjectID, page int, numberOfMessages int) ([]models.Message, error)
	CreateChat(chat models.Chat) (models.Chat, error)
	UpdateChat(chat models.Chat) error
	UpdateChatStatus(id primitive.ObjectID, status string, startTime time.Time) error
	UpdatePromptVersion(id primitive.ObjectID, promptVersion int) error
	DeleteChat(id primitive.ObjectID) error
	TrimMessages(chat models.Chat) models.Chat
	DecryptToken(token string) (primitive.ObjectID, primitive.ObjectID, error)
	DecryptTokenFromFrontend(token string) (string, error)
	GetBackupChats() ([]models.BackupChat, error)
	HandleBackupChat()
	// Summaty sheet
	GetSummarysheetByChatId(id primitive.ObjectID) (models.SummarySheet, error)
	UpdateSummarysheet(id primitive.ObjectID, summarySheet models.SummarySheet) error
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

func (cs *chatService) GetChatById(id primitive.ObjectID, page int, numberOfMessages int) (models.Chat, error) {
	chat, err := cs.chatRepo.GetChatById(id)
	if err != nil {
		return models.Chat{}, err
	}
	if page == 0 {
		return chat, nil
	}
	chat.Messages = chat.Messages[max(0, len(chat.Messages)-page*numberOfMessages):max(0, len(chat.Messages)-page*numberOfMessages+numberOfMessages)]
	return chat, nil
}

func (cs *chatService) GetChatsByBatchId(batchId string) ([]models.Chat, error) {
	chats, err := cs.chatRepo.GetChatsByBatchId(batchId)
	if err != nil {
		return nil, err
	}
	return chats, nil
}

func (cs *chatService) GetMessages(id primitive.ObjectID, page int, numberOfMessages int) ([]models.Message, error) {
	chat, err := cs.chatRepo.GetChatById(id)
	if err != nil {
		return nil, err
	}
	if page == 0 {
		return chat.Messages, nil
	}
	messages := chat.Messages[max(0, len(chat.Messages)-page*numberOfMessages):max(0, len(chat.Messages)-(page-1)*numberOfMessages)]
	return messages, nil
}

func (cs *chatService) CreateChat(chat models.Chat) (models.Chat, error) {
	result, err := cs.chatRepo.CreateChat(chat)
	if err != nil {
		return models.Chat{}, err
	}
	return result, nil
}

func (cs *chatService) UpdateChat(chat models.Chat) error {
	err := cs.chatRepo.UpdateChat(chat)
	if err != nil {
		return err
	}
	return nil
}

func (cs *chatService) UpdateChatStatus(id primitive.ObjectID, status string, startTime time.Time) error {
	err := cs.chatRepo.UpdateChatStatus(id, status, startTime)
	if err != nil {
		return err
	}
	return nil
}

func (cs *chatService) UpdatePromptVersion(id primitive.ObjectID, promptVersion int) error {
	err := cs.chatRepo.UpdatePromptVersion(id, promptVersion)
	if err != nil {
		return err
	}
	return nil
}

func (cs *chatService) DeleteChat(id primitive.ObjectID) error {
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

func (cs *chatService) decodeBase62(encoded string) ([]byte, error) {
	const base62Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	num := big.NewInt(0)
	base := big.NewInt(62)

	for i := 0; i < len(encoded); i++ {
		index := strings.IndexByte(base62Chars, encoded[i])
		if index == -1 {
			return nil, fmt.Errorf("invalid character %c in Base62 string", encoded[i])
		}
		num.Mul(num, base)
		num.Add(num, big.NewInt(int64(index)))
	}

	return num.Bytes(), nil
}

func (cs *chatService) decryptDataDeterministic(encodedCiphertext string) ([]byte, error) {
	key := sha256.Sum256([]byte(configs.GetEnv("TOKEN_KEY")))

	ciphertext, err := cs.decodeBase62(encodedCiphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += aes.BlockSize {
		block.Decrypt(plaintext[i:i+aes.BlockSize], ciphertext[i:i+aes.BlockSize])
	}

	return plaintext, nil
}

func (cs *chatService) DecryptToken(token string) (primitive.ObjectID, primitive.ObjectID, error) {
	decryptedData, err := cs.decryptDataDeterministic(token)
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}

	ids := map[string]string{}
	if err := json.Unmarshal(decryptedData, &ids); err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}

	sessionId, err := primitive.ObjectIDFromHex(ids["id1"])
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}
	chatId, err := primitive.ObjectIDFromHex(ids["id2"])
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}
	return sessionId, chatId, nil
}

func (cs *chatService) DecryptTokenFromFrontend(token string) (string, error) {
	key := configs.GetEnv("TOKEN_KEY2")

	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}

	keyLen := len(key)
	output := make([]byte, len(decoded))

	for i := 0; i < len(decoded); i++ {
		output[i] = decoded[i] ^ key[i%keyLen]
	}
	return string(output), nil
}

func (cs *chatService) GetBackupChats() ([]models.BackupChat, error) {
	backupChats, err := cs.chatRepo.GetBackupChats()
	if err != nil {
		return nil, err
	}
	return backupChats, nil
}

func (cs *chatService) HandleBackupChat() {
	for {
		if time.Now().Local().Hour() == 17 && time.Now().Local().Minute() == 00 && time.Now().Local().Second() == 00 {
			log.Println("Backing up...")
			if err := cs.chatRepo.BackupChats(); err != nil {
				log.Println(err)
			}
			time.Sleep(time.Duration(time.Second))
			log.Println("Finish backing up")
		}
	}
}

func (cs *chatService) GetSummarysheetByChatId(id primitive.ObjectID) (models.SummarySheet, error) {
	summartSheet, err := cs.chatRepo.GetSummarysheetByChatId(id)
	if err != nil {
		return models.SummarySheet{}, err
	}
	return summartSheet, nil
}

func (cs *chatService) UpdateSummarysheet(id primitive.ObjectID, summarySheet models.SummarySheet) error {
	err := cs.chatRepo.UpdateSummarysheet(id, summarySheet)
	if err != nil {
		return err
	}
	return nil
}
