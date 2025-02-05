package main

import (
	"log"

	"github.com/Meeyok-Chat/backend/configs"
	"github.com/Meeyok-Chat/backend/middleware"
	"github.com/Meeyok-Chat/backend/repository/cache"
	"github.com/Meeyok-Chat/backend/repository/database"
	"github.com/Meeyok-Chat/backend/repository/queue/queuePublisher"
	"github.com/Meeyok-Chat/backend/routes"
	"github.com/Meeyok-Chat/backend/services/chat"
	Websocket "github.com/Meeyok-Chat/backend/services/websocket"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize a new Redis client. (Cache)
	redisClient, err := configs.NewRedisClient()
	if err != nil {
		log.Fatalf("Could not create Redis client: %v", err)
	}
	// Initialize a new MongoDB client. (Database)
	mongoClient, err := configs.NewMongoClient()
	if err != nil {
		log.Fatalf("Could not create MongoDB client: %v", err)
	}

	// Initialize a new queue publisher
	queuePublisher := queuePublisher.NewQueuePublisher()

	// Initialize a new chat repository
	chatRepo := database.NewChatRepo(mongoClient.Chats, mongoClient.BackupChats)
	// Initialize a new cache repository
	cacheRepo := cache.NewCacheRepo(redisClient)
	// Initialize a new chat service
	chatService := chat.NewChatService(chatRepo)
	go chatService.HandleBackupChat()

	// Initialize a websocket manager
	websocketManager := Websocket.NewManagerService(cacheRepo, chatRepo, queuePublisher)
	// go websocketManager.SummaryChat()

	middleware := middleware.NewAuthMiddleware(chatService)

	// Initialize a queue manager
	// queueReceiver := queueReceiver.NewConsumerManager(websocketManager, cacheRepo)
	// go queueReceiver.ReadResult()
	// go queueReceiver.ReadDLQ()

	// Initialize a routes
	r := gin.Default()
	r.Use(configs.EnableCORS())
	routes.WebsocketRoute(r, middleware, websocketManager, chatService)
	routes.ChatRoute(r, middleware, chatService, websocketManager)

	log.Fatal(r.Run(":" + configs.GetEnv("CHATBOT_PORT")))
}
