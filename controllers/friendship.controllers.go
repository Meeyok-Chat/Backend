package controllers

import (
	"net/http"

	"github.com/Meeyok-Chat/backend/models"
	service "github.com/Meeyok-Chat/backend/services/friendship"

	"github.com/gin-gonic/gin"
)

type friendshipController struct {
	friendshipService service.FriendshipService
}

type FriendshipController interface {
	GetFriendsByStatusHandler(ctx *gin.Context)
	AddFriendshipHandler(ctx *gin.Context)
	AcceptFriendshipHandler(ctx *gin.Context)
	RejectFriendshipHandler(ctx *gin.Context)
}

func NewFriendshipController(friendshipService service.FriendshipService) FriendshipController {
	return &friendshipController{
		friendshipService: friendshipService,
	}
}

// GetFriendUsersByStatusHandler godoc
// @Summary      Get list of friends with status filter
// @Description  Returns a list of users who are friends or pending with the given user
// @Tags         friendship
// @Accept       json
// @Produce      json
// @Param        status  path     string  true  "Friendship status: accepted, pending, or rejected"
// @Security     Bearer
// @Success      200     {array}   models.User
// @Failure      400     {object}  models.HTTPError
// @Failure      500     {object}  models.HTTPError
// @Router       /friendships/{status} [get]
func (c *friendshipController) GetFriendsByStatusHandler(ctx *gin.Context) {
	userID, ok := ctx.Get("id")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "User not found"})
		return
	}
	status := ctx.Param("status")

	if userID == "" || (status != models.FriendshipPending && status != models.FriendshipAccepted && status != models.FriendshipRejected) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid userID or status"})
		return
	}

	users, err := c.friendshipService.GetFriends(userID.(string), status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(users) == 0 {
		ctx.JSON(http.StatusOK, []models.User{})
	} else {
		ctx.JSON(http.StatusOK, users)
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
	userID1, ok := ctx.Get("id")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "User not found"})
		return
	}
	userID2 := ctx.Param("id")

	friendship, err := c.friendshipService.AddFriendship(userID1.(string), userID2)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, friendship)
}

// AcceptFriendshipHandler godoc
// @Summary      Accept a friend request
// @Description  Accepts a pending friend request between the current user and the specified user
// @Tags         friendship
// @Accept       json
// @Produce      json
// @Param        userId  path      string  true  "Friend's user ID"
// @Security     Bearer
// @Success      200     {object}  models.Friendship
// @Failure      400     {object}  models.HTTPError
// @Failure      500     {object}  models.HTTPError
// @Router       /friendships/accept/{userId} [patch]
func (c *friendshipController) AcceptFriendshipHandler(ctx *gin.Context) {
	userID1, ok := ctx.Get("id")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "User not found"})
		return
	}
	userID2 := ctx.Param("userId")

	friendship, err := c.friendshipService.UpdateFriendshipStatus(userID1.(string), userID2, models.FriendshipAccepted)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, friendship)
}

// RejectFriendshipHandler godoc
// @Summary      Reject a friend request
// @Description  Rejects a pending friend request between the current user and the specified user
// @Tags         friendship
// @Accept       json
// @Produce      json
// @Param        userId  path      string  true  "Friend's user ID"
// @Security     Bearer
// @Success      200     {object}  models.Friendship
// @Failure      400     {object}  models.HTTPError
// @Failure      500     {object}  models.HTTPError
// @Router       /friendships/reject/{userId} [patch]
func (c *friendshipController) RejectFriendshipHandler(ctx *gin.Context) {
	userID1, ok := ctx.Get("id")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "User not found"})
		return
	}
	userID2 := ctx.Param("userId")

	friendship, err := c.friendshipService.UpdateFriendshipStatus(userID1.(string), userID2, models.FriendshipRejected)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, friendship)
}
