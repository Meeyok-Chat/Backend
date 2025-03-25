package controllers

import (
	"net/http"

	"github.com/Meeyok-Chat/backend/models"
	service "github.com/Meeyok-Chat/backend/services/post"
	"github.com/gin-gonic/gin"
)

type postController struct {
	postService service.PostService
}

type PostController interface {
	GetPosts(ctx *gin.Context)
	GetPostByID(c *gin.Context)
	CreatePost(c *gin.Context)
	UpdatePost(c *gin.Context)
	DeletePost(c *gin.Context)
}

func NewPostController(postService service.PostService) PostController {
	return &postController{
		postService: postService,
	}
}

// GetAllPostsHandler godoc
// @Summary      Get all posts
// @Description  Retrieves all posts from the database
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {array}   models.Post
// @Failure      500  {object}  models.HTTPError
// @Router       /posts [get]
func (p *postController) GetPosts(ctx *gin.Context) {
	posts, err := p.postService.GetPosts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, posts)
}

// @Summary Get a post by ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} models.Post
// @Failure 400 {object} map[string]string
// @Router /posts/{id} [get]
func (p *postController) GetPostByID(c *gin.Context) {
	id := c.Param("id")
	post, err := p.postService.GetPostByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, post)
}

// @Summary Create a new post
// @Tags posts
// @Accept json
// @Produce json
// @Param post body models.Post true "Post details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /posts [post]
func (p *postController) CreatePost(c *gin.Context) {
	post := models.Post{}
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	createdPost, err := p.postService.CreatePost(post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = p.postService.AddPostToUser(post.UserID, createdPost.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post created successfully"})
}

// @Summary Update a post
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param post body map[string]string true "Updated content"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /posts/{id} [put]
func (p *postController) UpdatePost(c *gin.Context) {
	id := c.Param("id")
	post := models.Post{}
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if err := p.postService.UpdatePost(id, post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post updated successfully"})
}

// @Summary Delete a post
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /posts/{id} [delete]
func (p *postController) DeletePost(c *gin.Context) {
	id := c.Param("id")
	if err := p.postService.DeletePost(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}
