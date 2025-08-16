package middleware

import (
	"context"
	"log"
	"net/http"
	"provolo-api/internal/types"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates a middleware for Firebase authentication
func AuthMiddleware(authClient *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("session")
		if err != nil {
			c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
				"Unauthorized",
				"No session found",
			))
			c.Abort()
			return
		}

		ctx := context.Background()
		token, err := authClient.VerifySessionCookie(ctx, cookie)
		if err != nil {
			log.Printf("Error verifying session: %v", err)
			c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
				"Unauthorized",
				"Invalid or expired session",
			))
			c.Abort()
			return
		}

		// Add user ID to request context
		c.Set("userID", token.UID)
		c.Set("userEmail", token.Claims["email"])
		c.Next()
	}
}
