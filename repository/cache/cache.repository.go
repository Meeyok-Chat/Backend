package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Meeyok-Chat/backend/configs"
	"github.com/Meeyok-Chat/backend/models"
)

type cacheRepo struct {
	cache *configs.RedisClient
}

type CacheRepo interface {
	UpdateCache(chatData models.Chat, clientData models.ClientData) error
	UpdateChatCache(chatId string, chatData models.Chat)
	UpdateMessagesCache(chatId string, messages []models.Message) error
	AppendMessageCache(chatId string, message models.Message) error
	UpdateClientCache(chatId string, clientData models.ClientData)
	CheckCache(chatId string) (models.CacheData, error)
}

func NewCacheRepo(cache *configs.RedisClient) CacheRepo {
	return &cacheRepo{
		cache: cache,
	}
}

func (c *cacheRepo) UpdateCache(chatData models.Chat, clientData models.ClientData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cacheData := &models.CacheData{
		ChatData:   chatData,
		ClientData: clientData,
	}

	response := map[string]interface{}{
		"data": cacheData,
	}
	mr, err := json.Marshal(&response)
	if err != nil {
		return err
	}

	err = c.cache.Client.Set(ctx, cacheData.ChatData.ID.Hex(), mr, 0).Err()
	if err != nil {
		return err
	}
	log.Println("Value set in cache successfully")
	return nil
}

func (c *cacheRepo) UpdateChatCache(chatId string, chatData models.Chat) {
	cacheData, err := c.CheckCache(chatId)
	if err != nil {
		c.UpdateCache(chatData, models.ClientData{})
		return
	}
	c.UpdateCache(chatData, cacheData.ClientData)
}

func (c *cacheRepo) UpdateMessagesCache(chatId string, messages []models.Message) error {
	cacheData, err := c.CheckCache(chatId)
	if err != nil {
		return fmt.Errorf("error : cannot update messages, because there is no chat (%s) in cache", chatId)
	}
	cacheData.ChatData.Messages = messages
	c.UpdateCache(cacheData.ChatData, cacheData.ClientData)
	return nil
}

func (c *cacheRepo) AppendMessageCache(chatId string, message models.Message) error {
	cacheData, err := c.CheckCache(chatId)
	if err != nil {
		return fmt.Errorf("error : cannot append message, because there is no chat (%s) in cache", chatId)
	}
	cacheData.ChatData.Messages = append(cacheData.ChatData.Messages, message)
	c.UpdateCache(cacheData.ChatData, cacheData.ClientData)
	return nil
}

func (c *cacheRepo) UpdateClientCache(chatId string, clientData models.ClientData) {
	cacheData, err := c.CheckCache(chatId)
	if err != nil {
		c.UpdateCache(models.Chat{}, clientData)
		return
	}
	c.UpdateCache(cacheData.ChatData, clientData)
}

func (c *cacheRepo) CheckCache(chatId string) (models.CacheData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("check chat in cache ...")
	value, err := c.cache.Client.Get(ctx, chatId).Result()
	if err != nil {
		// Key does not exist in cache
		log.Println("Key does not exist in cache")
		return models.CacheData{}, fmt.Errorf("error : cannot find chat (%s) in cache", chatId)
	}

	log.Println("Key exists in cache, retrieve data ...")
	// Key exist in cache
	var retrievedResponse map[string]interface{}
	err = json.Unmarshal([]byte(value), &retrievedResponse)
	if err != nil {
		return models.CacheData{}, err
	}

	data, ok := retrievedResponse["data"].(map[string]interface{})
	if !ok {
		return models.CacheData{}, fmt.Errorf("error : cannot find data in cache")
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return models.CacheData{}, err
	}

	var Retrieved models.CacheData
	err = json.Unmarshal(dataBytes, &Retrieved)
	if err != nil {
		return models.CacheData{}, err
	}
	return Retrieved, nil
}
