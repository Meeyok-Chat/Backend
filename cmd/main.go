package main

import (
	"log"

	_ "github.com/Meeyok-Chat/backend/cmd/docs"
	"github.com/Meeyok-Chat/backend/configs"
	"github.com/Meeyok-Chat/backend/middleware"
	"github.com/Meeyok-Chat/backend/repository/database"
	"github.com/Meeyok-Chat/backend/repository/queue/queuePublisher"
	"github.com/Meeyok-Chat/backend/repository/queue/queueReceiver"
	"github.com/Meeyok-Chat/backend/routes"
	"github.com/Meeyok-Chat/backend/services/chat"
	"github.com/Meeyok-Chat/backend/services/friendship"
	"github.com/Meeyok-Chat/backend/services/user"
	Websocket "github.com/Meeyok-Chat/backend/services/websocket"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization

// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      https://meeyok-cloudrun-image-719562346977.asia-southeast1.run.app
// @BasePath

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	mongoClient, err := configs.NewMongoClient()
	if err != nil {
		log.Fatalf("Could not create MongoDB client: %v", err)
	}
	FirebaseClient, err := configs.NewFirebaseClient()
	if err != nil {
		log.Fatalf("Error initializing Firebase auth: %v", err)
	}

	// Initialize a new repositories
	chatRepo := database.NewChatRepo(mongoClient.Chats)
	userRepo := database.NewUserRepo(mongoClient.User)
	friendshipRepo := database.NewFriendshipRepo(mongoClient.Friendship)

	// Initialize a new services
	chatService := chat.NewChatService(chatRepo)
	userService := user.NewUserService(userRepo)
	friendshipService := friendship.NewFriendshipService(friendshipRepo)

	// Initialize a queue Publisher
	queuePublisher := queuePublisher.NewQueuePublisher()

	// Initialize a websocket manager
	websocketManager := Websocket.NewManagerService(queuePublisher, chatRepo, userRepo)

	// Initialize a queue manager Receiver
	queueReceiver := queueReceiver.NewConsumerManager(websocketManager)
	go queueReceiver.ReadResult()

	// Initialize a new client for firebase authentication
	middleware := middleware.NewAuthMiddleware(userService)

	// Initialize a routes
	r := gin.Default()
	r.Use(configs.EnableCORS())
	routes.WebsocketRoute(r, middleware, FirebaseClient, websocketManager, chatService)
	routes.ChatRoute(r, middleware, FirebaseClient, userService, chatService, websocketManager)
	routes.UserRoute(r, middleware, FirebaseClient, userService)
	routes.FriendshipRoute(r, middleware, FirebaseClient, friendshipService)
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.DefaultModelsExpandDepth(-1),
	))

	log.Fatal(r.Run(":" + configs.GetEnv("PORT")))
}
