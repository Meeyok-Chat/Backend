package routes

import (
	"github.com/Meeyok-Chat/backend/controllers"
	"github.com/Meeyok-Chat/backend/middleware"
	"github.com/Meeyok-Chat/backend/services/chat"
	Websocket "github.com/Meeyok-Chat/backend/services/websocket"
	"github.com/gin-gonic/gin"
)

func ChatRoute(r *gin.Engine, middleware middleware.AuthMiddleware, chatService chat.ChatService, websocketManager Websocket.ManagerService) {
	chatController := controllers.NewChatController(chatService, websocketManager)

	rgc := r.Group("/chat")
	rgc.Use(middleware.Auth())
	{
		rgc.GET("all", chatController.GetChats)
		rgc.GET("", chatController.GetChatById)
		rgc.GET("/batch/:batchId", chatController.GetChatsByBatchId)
		rgc.POST("", chatController.CreateChat)
		rgc.PATCH("", chatController.UpdateChat)
		rgc.PATCH("/status/:status", chatController.UpdateChatStatus)
		rgc.DELETE("", chatController.DeleteChat)
		rgc.GET("/message", chatController.GetMessages)
		rgc.POST("/prompt/:version", chatController.SetPrompt)

		rgc.GET("/backup", chatController.GetBackupChats)

		rgc.GET("/summarysheet", chatController.GetSummarysheetByChatId)
	}
	r.PATCH("/chat/summarysheet/:id", chatController.UpdateSummarysheetByChatId)
}
