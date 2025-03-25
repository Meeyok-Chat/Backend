package controllers

import (
	"net/http"

	service "github.com/Meeyok-Chat/backend/services/friendship"

	"github.com/gin-gonic/gin"
)

type friendshipController struct {
	friendshipService service.FriendshipService
}

type FriendshipController interface {
	AddFriendshipHandler(ctx *gin.Context)
	AcceptFriendshipHandler(ctx *gin.Context)
	RejectFriendshipHandler(ctx *gin.Context)
}

func NewFriendshipController(friendshipService service.FriendshipService) FriendshipController {
	return &friendshipController{
		friendshipService: friendshipService,
	}
}

func (c *friendshipController) AddFriendshipHandler(ctx *gin.Context) {
	userID1 := ctx.Query("id1")
	userID2 := ctx.Query("id2")

	friendship, err := c.friendshipService.AddFriendship(userID1, userID2)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, friendship)
}

func (c *friendshipController) AcceptFriendshipHandler(ctx *gin.Context) {
	friendshipID := ctx.Param("id")

	friendship, err := c.friendshipService.AcceptFriendship(friendshipID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, friendship)
}

func (c *friendshipController) RejectFriendshipHandler(ctx *gin.Context) {
	friendshipID := ctx.Param("id")

	friendship, err := c.friendshipService.RejectFriendship(friendshipID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, friendship)
}
