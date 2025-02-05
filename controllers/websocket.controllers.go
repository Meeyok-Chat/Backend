package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Meeyok-Chat/backend/configs"
	"github.com/Meeyok-Chat/backend/services/chat"
	Websocket "github.com/Meeyok-Chat/backend/services/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type websocketController struct {
	websocketManagerService Websocket.ManagerService
	chatService             chat.ChatService
}

type WebsocketController interface {
	InitWebsocket(c *gin.Context)
	ServeWS(c *gin.Context)
}

func NewWebsocketController(websocketManagerService Websocket.ManagerService, chatService chat.ChatService) WebsocketController {
	return &websocketController{
		websocketManagerService: websocketManagerService,
		chatService:             chatService,
	}
}

func (ws *websocketController) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	switch origin {
	case configs.GetEnv("FRONTEND_URL"):
		return true
	default:
		return false
	}
}

func (ws *websocketController) InitWebsocket(c *gin.Context) {
	chatId, exists := c.Get("chat")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Chat not found"})
		return
	}
	id := chatId.(primitive.ObjectID)

	if err := ws.websocketManagerService.CheckOldClient(id.Hex()); err != nil {
		c.Error(fmt.Errorf("error creating client: %w", err))
		c.JSON(http.StatusBadRequest, gin.H{"result": "fail"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

// serveWS is a HTTP Handler that has the Manager that allows connections
func (ws *websocketController) ServeWS(c *gin.Context) {
	encryptedToken := c.Param("token")

	encryptedToken = strings.Replace(encryptedToken, "-", "+", -1)
	encryptedToken = strings.Replace(encryptedToken, "_", "/", -1)
	encryptedToken = strings.Replace(encryptedToken, ".", "=", -1)

	token, err := ws.chatService.DecryptTokenFromFrontend(encryptedToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	_, id, err := ws.chatService.DecryptToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	// Begin by upgrading the HTTP request
	websocketUpgrader := websocket.Upgrader{
		// Apply the Origin Checker
		CheckOrigin:     ws.checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Add the newly created client to the manager
	ws.websocketManagerService.AddClient(conn, c, id)
}
