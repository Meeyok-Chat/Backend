package configs

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Middleware function to handle CORS
func EnableCORS() gin.HandlerFunc {
	// Parse allowed origins from environment variable
	allowedOrigins := strings.Split(GetEnv("FRONTEND_URLS"), ",")
	allowedOrigins = append(allowedOrigins, "https://meeyok-frontend.pages.dev/") // TODO: move this to env

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if the origin is in the allowed list
		allowed := false
		for _, o := range allowedOrigins {
			if o == origin {
				allowed = true
				break
			}
		}

		// Set CORS headers if the origin is allowed
		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			if allowed {
				c.AbortWithStatus(http.StatusOK)
			} else {
				c.AbortWithStatus(http.StatusForbidden)
			}
			return
		}

		// Call the next handler
		c.Next()
	}
}
