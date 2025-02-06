package controllers

import (
	"net/http"
	"strconv"

	"github.com/Meeyok-Chat/backend/models"
	"github.com/Meeyok-Chat/backend/services/chat"
	"github.com/Meeyok-Chat/backend/services/user"
	Websocket "github.com/Meeyok-Chat/backend/services/websocket"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type chatController struct {
	chatService      chat.ChatService
	userService      user.UserService
	websocketManager Websocket.ManagerService
}

type ChatController interface {
	GetChats(c *gin.Context)
	GetChatById(c *gin.Context)
	CreateChat(c *gin.Context)
	AddUsersToChat(c *gin.Context)
	UpdateChat(c *gin.Context)
	DeleteChat(c *gin.Context)
	GetMessages(c *gin.Context)
}

func NewChatController(chatService chat.ChatService, userService user.UserService, websocketManager Websocket.ManagerService) ChatController {
	return &chatController{
		chatService:      chatService,
		userService:      userService,
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
	chatId := c.Param("id")

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

	chat, err := cc.chatService.GetChatById(chatId, page, numberOfMessages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chat)
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

	err = cc.userService.AddChatToUser(chat.Users, chat.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Chat created"})
}

func (cc *chatController) AddUsersToChat(c *gin.Context) {
	chatID := c.Param("id")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "chatID is required"})
		return
	}

	var req struct {
		Users []string `json:"users"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err := cc.chatService.AddUsersToChat(chatID, req.Users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	err = cc.userService.AddChatToUser(req.Users, chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Users added to chat successfully"})
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

func (cc *chatController) DeleteChat(c *gin.Context) {
	chatId := c.Param("id")
	if err := cc.chatService.DeleteChat(chatId); err != nil {
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
	id := chatId.(string)

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
