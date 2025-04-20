package controllers

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/Meeyok-Chat/backend/configs"
	"github.com/Meeyok-Chat/backend/services/chat"
	Websocket "github.com/Meeyok-Chat/backend/services/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type websocketController struct {
	websocketManagerService Websocket.ManagerService
	chatService             chat.ChatService
}

type WebsocketController interface {
	InitWebsocket(c *gin.Context)
	ServeWS(c *gin.Context)
	GetClients(c *gin.Context)
}

func NewWebsocketController(websocketManagerService Websocket.ManagerService, chatService chat.ChatService) WebsocketController {
	return &websocketController{
		websocketManagerService: websocketManagerService,
		chatService:             chatService,
	}
}

func (ws *websocketController) checkOrigin(r *http.Request) bool {
	allowedOrigins := strings.Split(configs.GetEnv("FRONTEND_URLS"), ",")
	origin := r.Header.Get("Origin")

	return slices.Contains(allowedOrigins, origin)
}

// InitWebsocket godoc
// @Summary      Initialize WebSocket connection
// @Description  Prepares for WebSocket connection by checking existing client
// @Tags         websocket
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  models.HTTPError  "Bad Request"
// @Failure      500  {object}  models.HTTPError  "Internal Server Error"
// @Router       /ws/init [get]
func (ws *websocketController) InitWebsocket(c *gin.Context) {
	userID, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "User not found"})
		return
	}
	id := userID.(string)

	if err := ws.websocketManagerService.CheckOldClient(id); err != nil {
		c.Error(fmt.Errorf("error creating client: %w", err))
		c.JSON(http.StatusBadRequest, gin.H{"result": "fail"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

// ServeWS godoc
// @Summary      Establish WebSocket connection
// @Description  Upgrades HTTP connection to WebSocket for real-time communication
// @Tags         websocket
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        userID  path      string  true  "User ID"
// @Success      101     "Switching Protocols"
// @Failure      400     {object}  models.HTTPError  "Bad Request"
// @Failure      500     {object}  models.HTTPError  "Internal Server Error"
// @Router       /ws/{userID} [get]
func (ws *websocketController) ServeWS(c *gin.Context) {
	userID := c.Param("userID")

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
	ws.websocketManagerService.AddClient(conn, c, userID)
}

// GetClients godoc
// @Summary      List active WebSocket clients
// @Description  Retrieves a list of currently connected WebSocket clients
// @Tags         websocket
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  map[string][]string
// @Failure      500  {object}  models.HTTPError  "Internal Server Error"
// @Router       /ws/clients [get]
func (ws *websocketController) GetClients(c *gin.Context) {
	users := ws.websocketManagerService.GetClients()
	c.JSON(http.StatusOK, gin.H{"users": users})
}
