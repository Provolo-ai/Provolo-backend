package middleware

import (
	"context"
	"log"
	"net/http"
	"provolo-api/internal/types"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates a middleware for Firebase authentication
func AuthMiddleware(authClient *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		var token *auth.Token
		var err error

		// Try Bearer token first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			idToken := strings.TrimPrefix(authHeader, "Bearer ")
			token, err = authClient.VerifyIDToken(ctx, idToken)
			if err != nil {
				log.Printf("Error verifying ID token: %v", err)
				c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
					"Unauthorized",
					"Invalid or expired token",
				))
				c.Abort()
				return
			}
		} else {
			// Fallback to session cookie
			cookie, err := c.Cookie("session")
			if err != nil {
				c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
					"Unauthorized",
					"No authentication provided",
				))
				c.Abort()
				return
			}

			token, err = authClient.VerifySessionCookie(ctx, cookie)
			if err != nil {
				log.Printf("Error verifying session: %v", err)
				c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
					"Unauthorized",
					"Invalid or expired session",
				))
				c.Abort()
				return
			}
		}

		// Add user info to request context
		c.Set("userID", token.UID)
		if email, ok := token.Claims["email"]; ok {
			c.Set("userEmail", email)
		}
		if displayName, ok := token.Claims["name"]; ok {
			c.Set("userDisplayName", displayName)
		}
		c.Next()
	}
}
