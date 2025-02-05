package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Meeyok-Chat/backend/models"
	"github.com/Meeyok-Chat/backend/services/chat"

	"github.com/gin-gonic/gin"
)

type authMiddleware struct {
	chatService chat.ChatService
}

type AuthMiddleware interface {
	Auth() gin.HandlerFunc
}

func NewAuthMiddleware(chatService chat.ChatService) AuthMiddleware {
	return &authMiddleware{
		chatService: chatService,
	}
}

func (s *authMiddleware) Auth() gin.HandlerFunc {
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
		tokenId := idToken[1]

		_, chatId, err := s.chatService.DecryptToken(tokenId)
		if err != nil {
			log.Println("Invalid Token", err)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, Invalid Token"})
			ctx.Abort()
			return
		}

		chat, err := s.chatService.GetChatById(chatId, 0, 10)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err})
			ctx.Abort()
			return
		}
		if chat.Status == models.ChatTerminated {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, Token Expired"})
			ctx.Abort()
			return
		}

		if chat.Status == models.ChatNotStarted {
			startTime := time.Now()
			if chat.Role == models.SpecialistRole {
				startTime = startTime.Add(time.Second * 31536000)
			}

			err := s.chatService.UpdateChatStatus(chat.ID, models.ChatProcessing, startTime)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
				ctx.Abort()
				return
			}
		} else if time.Since(chat.StartTime) > 24*time.Hour {
			if chat.Status == models.ChatProcessing {
				err := s.chatService.UpdateChatStatus(chat.ID, models.ChatCompleted, time.Time{})
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
					ctx.Abort()
					return
				}
			}
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, Token Expired"})
			ctx.Abort()
			return
		}

		ctx.Set("chat", chatId)
		log.Println("Successfully authenticated")
		ctx.Next()
		log.Println("Auth time:", time.Since(startTime))
	}
}
