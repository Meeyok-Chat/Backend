package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/Meeyok-Chat/backend/controllers"
	"github.com/Meeyok-Chat/backend/middleware"
	"github.com/Meeyok-Chat/backend/services/chat"
	"github.com/Meeyok-Chat/backend/services/user"
	Websocket "github.com/Meeyok-Chat/backend/services/websocket"
	"github.com/gin-gonic/gin"
)

func ChatRoute(r *gin.Engine, middleware middleware.AuthMiddleware, client *auth.Client, userService user.UserService, chatService chat.ChatService, websocketManager Websocket.ManagerService) {
	chatController := controllers.NewChatController(chatService, userService, websocketManager)

	rgc := r.Group("/chats")
	rgc.Use(middleware.Auth(client))
	{
		rgc.GET("", chatController.GetChats)
		rgc.GET("/:id", chatController.GetChatById)
		rgc.GET("/user/:type", chatController.GetUserChats)

		rgc.POST("", chatController.CreateChat)
		rgc.PATCH("/addusers/:id", chatController.AddUsersToChat)

		rgc.PATCH("/:id", chatController.UpdateChat)
		rgc.DELETE("/:id", chatController.DeleteChat)

		// rgc.GET("/message", chatController.GetMessages)
	}
}
