package main

import (
	"log"

	"github.com/Meeyok-Chat/backend/configs"
	"github.com/Meeyok-Chat/backend/middleware"
	"github.com/Meeyok-Chat/backend/repository/database"
	"github.com/Meeyok-Chat/backend/repository/queue/queuePublisher"
	"github.com/Meeyok-Chat/backend/repository/queue/queueReceiver"
	"github.com/Meeyok-Chat/backend/routes"
	"github.com/Meeyok-Chat/backend/services/chat"
	"github.com/Meeyok-Chat/backend/services/user"
	Websocket "github.com/Meeyok-Chat/backend/services/websocket"
	"github.com/gin-gonic/gin"
)

func main() {
	mongoClient, err := configs.NewMongoClient()
	if err != nil {
		log.Fatalf("Could not create MongoDB client: %v", err)
	}

	// Initialize a new repositories
	chatRepo := database.NewChatRepo(mongoClient.Chats)
	userRepo := database.NewUserRepo(mongoClient.User)

	// Initialize a new services
	chatService := chat.NewChatService(chatRepo)
	userService := user.NewUserService(userRepo)

	// Initialize a queue Publisher
	queuePublisher := queuePublisher.NewQueuePublisher()

	// Initialize a websocket manager
	websocketManager := Websocket.NewManagerService(queuePublisher, chatRepo, userRepo)

	// Initialize a queue manager Receiver
	queueReceiver := queueReceiver.NewConsumerManager(websocketManager)
	go queueReceiver.ReadResult()

	// Initialize a new client for firebase authentication
	middleware := middleware.NewAuthMiddleware(userService)
	client, err := middleware.InitAuth()
	if err != nil {
		log.Fatalf("Error initializing Firebase auth: %v", err)
	}

	// Initialize a routes
	r := gin.Default()
	r.Use(configs.EnableCORS())
	routes.WebsocketRoute(r, middleware, client, websocketManager, chatService)
	routes.ChatRoute(r, middleware, client, userService, chatService, websocketManager)
	routes.UserRoute(r, middleware, client, userService)

	log.Fatal(r.Run(":" + configs.GetEnv("PORT")))
}
