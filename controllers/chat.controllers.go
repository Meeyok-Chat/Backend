package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Meeyok-Chat/backend/models"
	"github.com/Meeyok-Chat/backend/services/chat"
	Websocket "github.com/Meeyok-Chat/backend/services/websocket"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type chatController struct {
	chatService      chat.ChatService
	websocketManager Websocket.ManagerService
}

type ChatController interface {
	GetChats(c *gin.Context)
	GetChatById(c *gin.Context)
	GetChatsByBatchId(c *gin.Context)
	CreateChat(c *gin.Context)
	UpdateChat(c *gin.Context)
	UpdateChatStatus(c *gin.Context)
	DeleteChat(c *gin.Context)
	GetMessages(c *gin.Context)
	SetPrompt(c *gin.Context)
	GetBackupChats(c *gin.Context)
	// summary sheet
	GetSummarysheetByChatId(c *gin.Context)
	UpdateSummarysheetByChatId(c *gin.Context)
}

func NewChatController(chatService chat.ChatService, websocketManager Websocket.ManagerService) ChatController {
	return &chatController{
		chatService:      chatService,
		websocketManager: websocketManager,
	}
}

func (cc *chatController) GetChats(c *gin.Context) {
	chats, err := cc.chatService.GetChats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chats)
}

func (cc *chatController) GetChatById(c *gin.Context) {
	chatId, exists := c.Get("chat")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Chat not found"})
		return
	}
	id := chatId.(primitive.ObjectID)

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request parameter"})
		return
	}

	numberOfMessages, err := strconv.Atoi(c.DefaultQuery("num-message", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request parameter"})
		return
	}

	chat, err := cc.chatService.GetChatById(id, page, numberOfMessages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chat)
}

func (cc *chatController) GetChatsByBatchId(c *gin.Context) {
	batchId := c.Params.ByName("batchId")

	chats, err := cc.chatService.GetChatsByBatchId(batchId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chats)
}

func (cc *chatController) CreateChat(c *gin.Context) {
	chatDTO := models.Chat{}
	if err := c.ShouldBindJSON(&chatDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	chat, err := cc.chatService.CreateChat(chatDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chat.ID)
}

func (cc *chatController) UpdateChat(c *gin.Context) {
	chatId, exists := c.Get("chat")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Chat not found"})
		return
	}
	id := chatId.(primitive.ObjectID)

	chatDTO := models.Chat{}
	if err := c.ShouldBindJSON(&chatDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	chatDTO.ID = id
	if err := cc.chatService.UpdateChat(chatDTO); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Chat updated"})
}

func (cc *chatController) UpdateChatStatus(c *gin.Context) {
	chatId, exists := c.Get("chat")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Chat not found"})
		return
	}
	id := chatId.(primitive.ObjectID)

	status := c.Params.ByName("status")
	if err := cc.chatService.UpdateChatStatus(id, status, time.Time{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Chat status updated"})
}

func (cc *chatController) DeleteChat(c *gin.Context) {
	chatId, exists := c.Get("chat")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Chat not found"})
		return
	}
	id := chatId.(primitive.ObjectID)

	if err := cc.chatService.DeleteChat(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Chat deleted"})
}

func (cc *chatController) GetMessages(c *gin.Context) {
	chatId, exists := c.Get("chat")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Chat not found"})
		return
	}
	id := chatId.(primitive.ObjectID)

	page, err := strconv.Atoi(c.DefaultQuery("page", "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request parameter"})
		return
	}

	numberOfMessages, err := strconv.Atoi(c.DefaultQuery("num-message", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request parameter"})
		return
	}

	messages, err := cc.chatService.GetMessages(id, page, numberOfMessages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (cc *chatController) SetPrompt(c *gin.Context) {
	chatId, exists := c.Get("chat")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Chat not found"})
		return
	}
	id := chatId.(primitive.ObjectID)

	version, err := strconv.Atoi(c.Params.ByName("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request parameter"})
		return
	}

	err = cc.chatService.UpdatePromptVersion(id, version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Prompt version updated"})
}

func (cc *chatController) GetBackupChats(c *gin.Context) {
	chats, err := cc.chatService.GetBackupChats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chats)
}

func (cc *chatController) GetSummarysheetByChatId(c *gin.Context) {
	chatId, exists := c.Get("chat")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Chat not found"})
		return
	}
	id := chatId.(primitive.ObjectID)

	summartSheet, err := cc.chatService.GetSummarysheetByChatId(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summartSheet)
}

func (cc *chatController) UpdateSummarysheetByChatId(c *gin.Context) {
	chatId := c.Params.ByName("id")
	id, err := primitive.ObjectIDFromHex(chatId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request parameter"})
		return
	}

	summartSheet := models.SummarySheet{}
	if err := c.ShouldBindJSON(&summartSheet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := cc.chatService.UpdateSummarysheet(id, summartSheet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	cc.websocketManager.NewSummarySheetHandler(chatId)
	c.JSON(http.StatusOK, gin.H{"message": "Summary Sheet updated"})
}
