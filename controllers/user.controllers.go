package controllers

import (
	"net/http"

	"github.com/Meeyok-Chat/backend/models"
	service "github.com/Meeyok-Chat/backend/services/user"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type userController struct {
	userService service.UserService
}

type UserController interface {
	GetUsers(c *gin.Context)
	GetUserByID(c *gin.Context)
	GetUserByToken(c *gin.Context)
	CreateUser(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
}

func NewUserController(userService service.UserService) UserController {
	return &userController{
		userService: userService,
	}
}

func (uc userController) GetUsers(c *gin.Context) {
	result, err := uc.userService.GetUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (uc userController) GetUserByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Params.ByName("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request parameter"})
		return
	}

	result, err := uc.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (uc userController) GetUserByToken(c *gin.Context) {
	username, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "User not found"})
		return
	}
	result, err := uc.userService.GetUserByUsername(username.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (uc userController) CreateUser(c *gin.Context) {
	userDTO := models.User{}
	if err := c.ShouldBindJSON(&userDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := uc.userService.CreateUser(userDTO); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User created"})
}

func (uc userController) UpdateUser(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request parameter"})
		return
	}

	userDTO := models.User{}
	if err := c.ShouldBindJSON(&userDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	userDTO.ID = id
	if err := uc.userService.UpdateUser(userDTO); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}

func (uc userController) DeleteUser(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request parameter"})
		return
	}

	if err := uc.userService.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
