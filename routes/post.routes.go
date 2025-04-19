package routes

import (
	"firebase.google.com/go/v4/auth"
	"github.com/Meeyok-Chat/backend/controllers"
	"github.com/Meeyok-Chat/backend/middleware"
	"github.com/Meeyok-Chat/backend/services/post"
	"github.com/gin-gonic/gin"
)

func PostRoute(r *gin.Engine, middleware middleware.AuthMiddleware, client *auth.Client, postService post.PostService) {
	postController := controllers.NewPostController(postService)

	rgc := r.Group("/posts")
	rgc.Use(middleware.Auth(client))
	{
		rgc.GET("/", postController.GetPosts)
		rgc.GET("/:id", postController.GetPostByID)
		rgc.POST("", postController.CreatePost)
		rgc.PUT("/:id", postController.UpdatePost)
		rgc.DELETE("/:id", postController.DeletePost)
	}
}
