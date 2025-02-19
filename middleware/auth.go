package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Meeyok-Chat/backend/configs"
	"github.com/Meeyok-Chat/backend/models"
	service "github.com/Meeyok-Chat/backend/services/user"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

type authMiddleware struct {
	userService service.UserService
}

type AuthMiddleware interface {
	Auth(client *auth.Client) gin.HandlerFunc
	processToken(ctx *gin.Context, client *auth.Client, token *auth.Token, tokenID string)
	InitAuth() (*auth.Client, error)
	RoleAuth(requiredRole ...string) gin.HandlerFunc
	AssignRole(ctx context.Context, client *auth.Client, c *gin.Context, email string, role string) error
}

func NewAuthMiddleware(userService service.UserService) AuthMiddleware {
	return &authMiddleware{
		userService: userService,
	}
}

func (s authMiddleware) Auth(client *auth.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()

		header := ctx.Request.Header.Get("Authorization")
		if header == "" {
			log.Println("Missing Authorization header")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, Invalid Token"})
			ctx.Abort()
			return
		}
		idToken := strings.Split(header, "Bearer ")
		if len(idToken) != 2 {
			log.Println("Invalid Authorization header")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, Invalid Token"})
			ctx.Abort()
			return
		}
		tokenID := idToken[1]

		token, err := client.VerifyIDToken(context.Background(), tokenID)
		if err != nil {
			log.Printf("Error verifying token. Error: %v\n", err)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, Invalid Token"})
			ctx.Abort()
			return
		}
		s.processToken(ctx, client, token, tokenID)
		log.Println("Auth time:", time.Since(startTime))
	}
}

func (s authMiddleware) processToken(ctx *gin.Context, client *auth.Client, token *auth.Token, tokenID string) {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	email, ok := token.Claims["email"].(string)
	if !ok {
		log.Println("Email claim not found in token")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, Invalid Token"})
		ctx.Abort()
		return
	}
	log.Println("auth email is ", email)

	role, roleOk := token.Claims["role"].(string)
	username, usernameOk := token.Claims["name"].(string)

	if !usernameOk {
		username = email
	}

	if !roleOk {
		user, err := s.userService.GetUserByEmail(email)
		if err != nil {
			if email == adminEmail {
				err2 := s.userService.CreateUser(models.User{Email: adminEmail, Username: username, Role: "admin"})
				if err2 != nil {
					log.Printf("Error creating user: %v\n", err2)
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
					ctx.Abort()
					return
				}

				startTime := time.Now()
				if err := s.AssignRole(context.Background(), client, ctx, adminEmail, "admin"); err != nil {
					log.Printf("Error assigning admin role to %s: %v\n", adminEmail, err)
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
					ctx.Abort()
					return
				}
				log.Println("Admin role assigned in:", time.Since(startTime))
				role = "admin"
			} else {
				err2 := s.userService.CreateUser(models.User{Email: email, Username: username, Role: "user"})
				if err2 != nil {
					log.Printf("Error creating user: %v\n", err2)
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
					ctx.Abort()
					return
				}

				startTime := time.Now()
				if err := s.AssignRole(context.Background(), client, ctx, email, "user"); err != nil {
					log.Printf("Error assigning user role to %s: %v\n", email, err)
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
					ctx.Abort()
					return
				}
				log.Println("User role assigned in:", time.Since(startTime))
				role = "user"
			}
		} else {
			startTime := time.Now()
			if err := s.AssignRole(context.Background(), client, ctx, email, user.Role); err != nil {
				log.Printf("Error assigning user role to %s: %v\n", email, err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
				ctx.Abort()
				return
			}
			log.Println("User role assigned in:", time.Since(startTime))
			role = user.Role
		}
	}

	ctx.Set("email", email)
	ctx.Set("role", role)
	ctx.Set("token", tokenID)

	log.Println("Successfully authenticated")
	log.Printf("Email: %v\n", email)
	log.Printf("Role: %v\n", role)

	ctx.Next()
}

func (s authMiddleware) InitAuth() (*auth.Client, error) {
	credentialsPath := configs.GetFirebaseLocalCredentials()
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v", err)
		return nil, err
	}

	client, errAuth := app.Auth(context.Background())
	if errAuth != nil {
		log.Fatalf("error initializing firebase auth: %v", errAuth)
		return nil, errAuth
	}

	return client, nil
}

func (s authMiddleware) RoleAuth(allowedRoles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		username, exists := ctx.Get("user")
		if !exists {
			log.Println("User not found in context")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if username == "" {
			log.Println("Invalid user data in context")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		role, exists := ctx.Get("role")
		if !exists {
			log.Println("Role not found in context")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if role == "" {
			log.Println("User role not set")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Include "admin" in allowedRoles by default
		allRoles := append(allowedRoles, "admin")

		// Check if the role matches any of the allowed roles
		for _, allowedRole := range allRoles {
			if role == allowedRole {
				log.Printf("User with email %s and role %s authorized", username, role)
				ctx.Next()
				return
			}
		}

		log.Printf("User with email %s and role %s tried to access a route that requires roles: %v",
			username, role, allowedRoles)
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
}

func (s authMiddleware) AssignRole(ctx context.Context, client *auth.Client, c *gin.Context, email string, role string) error {
	startTime := time.Now()
	defer func() {
		log.Println("Role assigned in:", time.Since(startTime))
	}()
	user, err := client.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		log.Printf("Assign Error: User with email %s not found", email)
		return errors.New("user not found")
	}

	// userMongo, err := s.userService.GetUserByEmail(email)
	// if err != nil {
	// 	return fmt.Errorf("AssignRole Error: Error getting user: %w", err)
	// }
	// err2 := s.userService.UpdateUser(models.User{ID: userMongo.ID, Email: email, Role: role})
	// if err2 != nil {
	// 	return fmt.Errorf("AssignRole Error: Error updating user: %w", err2)
	// }

	currentCustomClaims := user.CustomClaims
	if currentCustomClaims == nil {
		currentCustomClaims = map[string]interface{}{}
	}
	currentCustomClaims["role"] = role
	if err := client.SetCustomUserClaims(ctx, user.UID, currentCustomClaims); err != nil {
		return fmt.Errorf("AssignRole Error: Error setting custom claims: %w", err)
	}
	log.Println(email, " with role ", role)
	return nil
}
