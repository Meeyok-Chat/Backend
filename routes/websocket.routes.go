package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/Meeyok-Chat/backend/controllers"
	"github.com/Meeyok-Chat/backend/middleware"
	"github.com/Meeyok-Chat/backend/services/chat"
	Websocket "github.com/Meeyok-Chat/backend/services/websocket"
	"github.com/gin-gonic/gin"
)

func WebsocketRoute(r *gin.Engine, middleware middleware.AuthMiddleware, client *auth.Client, managerService Websocket.ManagerService, chatService chat.ChatService) {
	websocketController := controllers.NewWebsocketController(managerService, chatService)

	r.GET("/initWs", middleware.Auth(client), websocketController.InitWebsocket)
	r.GET("/ws/:token", websocketController.ServeWS)
}
