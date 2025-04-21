package dtos

type CreatePostRequest struct {
	Title   string `json:"title" binding:"required" example:"Title of the post"`
	Content string `json:"content" binding:"required" example:"Content of the post"`
}
