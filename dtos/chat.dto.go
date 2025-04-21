package dtos

type CreateChatRequest struct {
	Name  string   `json:"name" binding:"required" example:"Team Discussion"`
	Users []string `json:"users" binding:"required" example:"user123,user456"`
	Type  string   `json:"type" binding:"required,oneof=Individual Group" example:"Group"`
}

type AddUsersRequest struct {
	Users []string `json:"users" binding:"required" example:"user123,user456"`
}
