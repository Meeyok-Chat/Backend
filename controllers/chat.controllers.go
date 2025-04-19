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
	GetUserChats(c *gin.Context)
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

// GetChats godoc
// @Summary      List all chats
// @Description  Retrieves a list of all available chats
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {array}   models.Chat
// @Failure      500  {object}  models.HTTPError
// @Router       /chats [get]
func (cc *chatController) GetChats(c *gin.Context) {
	chats, err := cc.chatService.GetChats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chats)
}

// GetChatById godoc
// @Summary      Get a specific chat by ID
// @Description  Retrieves a chat with messages based on the provided ID
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id           path      string  true  "Chat ID"
// @Param        page         query     int     false "Page number for pagination" default(1)
// @Param        num-message  query     int     false "Number of messages per page" default(10)
// @Security     Bearer
// @Success      200  {object}  models.Chat
// @Failure      400  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPError
// @Router       /chats/{id} [get]
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

// GetUserChats godoc
// @Summary      Get user chats based on type
// @Description  Retrieves chats based on the given type (group, friend, non-friend)
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        type   path     string  true  "Type of chat (group, friend, non-friend)"
// @Security     Bearer
// @Success      200  {array}   models.Chat
// @Failure      400  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPError
// @Router       /chats/user/{type} [get]
func (cc *chatController) GetUserChats(c *gin.Context) {
	email, ok := c.Get("email")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "User not found"})
		return
	}
	user, err := cc.userService.GetUserByEmail(email.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	userID := user.ID.Hex()
	chatType := c.Param("type")

	chats, err := cc.chatService.GetUserChats(userID, chatType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	if chats == nil {
		chats = []models.Chat{}
	}
	c.JSON(http.StatusOK, chats)
}

// CreateChat godoc
// @Summary      Create a new chat
// @Description  Creates a new chat with the provided chat details
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        chat  body      models.Chat  true  "Chat details"
// @Security     Bearer
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  models.HTTPError
// @Failure      500   {object}  models.HTTPError
// @Router       /chats [post]
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

// AddUsersToChat godoc
// @Summary      Add users to a chat
// @Description  Adds specified users to an existing chat
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id     path      string    true  "Chat ID"
// @Param        users  body      object    true  "List of user IDs to add"  schema({"type":"object","properties":{"users":{"type":"array","items":{"type":"string"}}}})
// @Security     Bearer
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  models.HTTPError
// @Failure      500   {object}  models.HTTPError
// @Router       /chats/{id}/users [post]
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

// UpdateChat godoc
// @Summary      Update a chat
// @Description  Updates the details of an existing chat
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "Chat ID"
// @Param        chat  body      models.Chat  true  "Updated chat details"
// @Security     Bearer
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  models.HTTPError
// @Failure      500   {object}  models.HTTPError
// @Router       /chats/{id} [put]
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

// DeleteChat godoc
// @Summary      Delete a chat
// @Description  Deletes a chat by its ID
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Chat ID"
// @Security     Bearer
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  models.HTTPError
// @Router       /chats/{id} [delete]
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
