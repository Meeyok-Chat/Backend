package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/Meeyok-Chat/backend/controllers"
	"github.com/Meeyok-Chat/backend/middleware"
	"github.com/Meeyok-Chat/backend/services/user"
	"github.com/gin-gonic/gin"
)

func UserRoute(r *gin.Engine, middleware middleware.AuthMiddleware, client *auth.Client, userService user.UserService) {
	userController := controllers.NewUserController(userService)

	rgu := r.Group("/users")
	rgu.Use(middleware.Auth(client))
	{
		rgu.GET("/me", userController.GetUserByToken)

		rgu.GET("", userController.GetUsers)
		rgu.GET("/:id", userController.GetUserByID)
		rgu.GET("/username/:username", userController.GetUserByUsername)

		rgu.PUT("/:id", userController.UpdateUser)
		rgu.PATCH("/:id/username", userController.UpdateUsername)

		rgu.DELETE("/:id", userController.DeleteUser)
	}
}
