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
	GetUserByUsername(c *gin.Context)

	CreateUser(c *gin.Context)

	UpdateUser(c *gin.Context)
	UpdateUsername(c *gin.Context)

	DeleteUser(c *gin.Context)
}

func NewUserController(userService service.UserService) UserController {
	return &userController{
		userService: userService,
	}
}

// GetUsers godoc
// @Summary      List all users
// @Description  Retrieves a list of all users
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {array}   models.User
// @Failure      500  {object}  models.HTTPError
// @Router       /users [get]
func (uc userController) GetUsers(c *gin.Context) {
	result, err := uc.userService.GetUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Retrieves a specific user by their ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  models.User
// @Failure      500  {object}  models.HTTPError
// @Router       /users/{id} [get]
func (uc userController) GetUserByID(c *gin.Context) {
	id := c.Params.ByName("id")

	result, err := uc.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetUserByToken godoc
// @Summary      Get current user details
// @Description  Retrieves the details of the authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  models.User
// @Failure      401  {object}  models.HTTPError  "Unauthorized"
// @Failure      500  {object}  models.HTTPError
// @Router       /users/me [get]
func (uc userController) GetUserByToken(c *gin.Context) {
	email, ok := c.Get("email")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "User not found"})
		return
	}
	result, err := uc.userService.GetUserByEmail(email.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetUserByUsername godoc
// @Summary      Get user by username
// @Description  Retrieves a user's details by their username
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        username  path      string  true  "Username"
// @Success      200  {object}  models.User
// @Failure      500  {object}  models.HTTPError
// @Router       /users/username/{username} [get]
func (uc userController) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")

	result, err := uc.userService.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Registers a new user in the system
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "User details"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  models.HTTPError  "Bad Request"
// @Failure      500   {object}  models.HTTPError
// @Router       /users [post]
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

// UpdateUser godoc
// @Summary      Update user details
// @Description  Updates an existing user's information
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id    path      string      true  "User ID"
// @Param        user  body      models.User true  "Updated user details"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  models.HTTPError  "Bad Request"
// @Failure      500   {object}  models.HTTPError
// @Router       /users/{id} [put]
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

// UpdateUsername godoc
// @Summary      Update username
// @Description  Updates the username for a specific user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id    path      string      true  "User ID"
// @Param        user  body      models.User true  "New username"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  models.HTTPError  "Bad Request"
// @Failure      500   {object}  models.HTTPError
// @Router       /users/{id}/username [patch]
func (uc userController) UpdateUsername(c *gin.Context) {
	userID := c.Param("id")

	userDTO := models.User{}
	if err := c.ShouldBindJSON(&userDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := uc.userService.UpdateUsername(userID, userDTO.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Removes a user from the system
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  models.HTTPError  "Bad Request"
// @Failure      500  {object}  models.HTTPError
// @Router       /users/{id} [delete]
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
