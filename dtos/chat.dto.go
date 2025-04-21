package dtos

type CreateChatRequest struct {
	Name  string   `json:"name" binding:"required" example:"My Group Chat"`
	Users []string `json:"users" binding:"required" example:"[\"user1\", \"user2\"]"`
	Type  string   `json:"type" binding:"required,oneof=Individual Group" example:"Group"`
}

type AddUsersRequest struct {
	Users []string `json:"users" binding:"required" example:"[\"67a4665f38d8f842368969ad\"]"`
}
