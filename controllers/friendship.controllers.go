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

// AddFriendshipHandler godoc
// @Summary      Send a friend request
// @Description  Sends a friend request from one user to another
// @Tags         friendship
// @Accept       json
// @Produce      json
// @Param        id1   query     string  true  "User ID of the requester"
// @Param        id2   query     string  true  "User ID of the recipient"
// @Security     Bearer
// @Success      200   {object}  models.Friendship
// @Failure      400   {object}  models.HTTPError
// @Failure      500   {object}  models.HTTPError
// @Router       /friendships [post]
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

// AcceptFriendshipHandler godoc
// @Summary      Accept a friend request
// @Description  Accepts a pending friend request by updating the status
// @Tags         friendship
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Friendship ID"
// @Security     Bearer
// @Success      200   {object}  models.Friendship
// @Failure      400   {object}  models.HTTPError
// @Failure      500   {object}  models.HTTPError
// @Router       /friendships/{id}/accept [patch]
func (c *friendshipController) AcceptFriendshipHandler(ctx *gin.Context) {
	friendshipID := ctx.Param("id")

	friendship, err := c.friendshipService.AcceptFriendship(friendshipID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, friendship)
}

// RejectFriendshipHandler godoc
// @Summary      Reject a friend request
// @Description  Rejects a pending friend request and removes it from the system
// @Tags         friendship
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Friendship ID"
// @Security     Bearer
// @Success      200   {object}  models.Friendship
// @Failure      400   {object}  models.HTTPError
// @Failure      500   {object}  models.HTTPError
// @Router       /friendships/{id}/reject [patch]
func (c *friendshipController) RejectFriendshipHandler(ctx *gin.Context) {
	friendshipID := ctx.Param("id")

	friendship, err := c.friendshipService.RejectFriendship(friendshipID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, friendship)
}
