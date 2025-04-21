package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/Meeyok-Chat/backend/controllers"
	"github.com/Meeyok-Chat/backend/middleware"
	"github.com/Meeyok-Chat/backend/services/friendship"
	"github.com/gin-gonic/gin"
)

func FriendshipRoute(r *gin.Engine, middleware middleware.AuthMiddleware, client *auth.Client, friendshipService friendship.FriendshipService) {
	friendshipController := controllers.NewFriendshipController(friendshipService)

	rgc := r.Group("/friendships")
	rgc.Use(middleware.Auth(client))
	{
		rgc.GET(":status", friendshipController.GetFriendsByStatusHandler)
		rgc.POST("/:id", friendshipController.AddFriendshipHandler)
		rgc.PATCH("/accept/:userId", friendshipController.AcceptFriendshipHandler)
		rgc.PATCH("/reject/:userId", friendshipController.RejectFriendshipHandler)
	}
}
