package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"provolo-api/internal/env"
	"provolo-api/internal/types"
	"provolo-api/internal/utils"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

type AuthHandler struct {
	client *auth.Client
}

type LoginRequest struct {
	IdToken string `json:"idToken" binding:"required"`
}

type User struct {
	UID         string `json:"uid"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	PhotoURL    string `json:"photoURL"`
}

// NewAuthHandler creates a new Firebase auth handler
func NewAuthHandler() (*AuthHandler, error) {
	ctx := context.Background()

	// Try to get encoded config from environment first
	encodedConfig := env.GetEnvString("FIREBASE_ENCODED_CONFIG", "")
	secretKey := env.GetEnvString("FIREBASE_SECRET_KEY", "")

	var opt option.ClientOption

	if encodedConfig != "" && secretKey != "" {
		// Decode the Firebase config
		configData, err := utils.DecodeFirebaseConfig(encodedConfig, secretKey)
		if err != nil {
			return nil, fmt.Errorf("error decoding Firebase config: %v", err)
		}

		// Create credentials from JSON data
		opt = option.WithCredentialsJSON(configData)
		log.Println("Using encoded Firebase configuration")
	} else {
		// Fallback to encoded file if environment variables are not set
		if _, err := os.Stat("firebase_config_encoded.txt"); err == nil {
			// Read encoded config from file
			encodedData, err := os.ReadFile("firebase_config_encoded.txt")
			if err != nil {
				return nil, fmt.Errorf("error reading encoded config file: %v", err)
			}

			if secretKey == "" {
				return nil, fmt.Errorf("FIREBASE_SECRET_KEY environment variable is required")
			}

			// Decode the Firebase config
			configData, err := utils.DecodeFirebaseConfig(string(encodedData), secretKey)
			if err != nil {
				return nil, fmt.Errorf("error decoding Firebase config: %v", err)
			}

			// Create credentials from JSON data
			opt = option.WithCredentialsJSON(configData)
			log.Println("Using encoded Firebase configuration from file")
		} else {
			// Final fallback to original file (for development)
			opt = option.WithCredentialsFile("firebaseConfig.json")
			log.Println("Using original Firebase configuration file (development mode)")
		}
	}

	// Initialize Firebase Admin SDK
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %v", err)
	}

	return &AuthHandler{client: client}, nil
}

// Login endpoint - creates session cookie
// @Summary User login
// @Description Authenticates user with Firebase ID token and creates session cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "Login request with ID token"
// @Success 200 {object} types.APIResponse
// @Failure 400 {object} types.APIResponse
// @Failure 401 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"Invalid Request",
			"Invalid request body: "+err.Error(),
		))
		return
	}

	ctx := context.Background()

	// Verify the ID token
	token, err := h.client.VerifyIDToken(ctx, req.IdToken)
	if err != nil {
		log.Printf("Error verifying ID token: %v", err)
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			"Authentication Failed",
			"Invalid token",
		))
		return
	}

	// Create session cookie (expires in 5 days)
	expiresIn := time.Hour * 24 * 5
	cookie, err := h.client.SessionCookie(ctx, req.IdToken, expiresIn)
	if err != nil {
		log.Printf("Error creating session cookie: %v", err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"Session Creation Failed",
			"Failed to create session",
		))
		return
	}

	// Set HTTP-only cookie
	domain := ""
	secure := false
	host := c.Request.Host
	if env.GetEnvString("ENVIRONMENT", "development") == "production" {
		// Set cookie domain for allowed production domains
		switch host {
		case "www.provolo.org":
			domain = ".provolo.org"
		case "provolo-front-end-dev-env.vercel.app":
			domain = ".vercel.app"
		default:
			domain = ""
		}
		secure = true
	} else if host == "localhost:5173" {
		domain = "localhost"
		secure = false
	}
	c.SetCookie(
		"session",
		cookie,
		int(expiresIn.Seconds()),
		"/",
		domain,
		secure,
		true,
	)

	// Get user info
	userRecord, err := h.client.GetUser(ctx, token.UID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"User Info Failed",
			"Failed to get user info",
		))
		return
	}

	user := &User{
		UID:         userRecord.UID,
		Email:       userRecord.Email,
		DisplayName: userRecord.DisplayName,
		PhotoURL:    userRecord.PhotoURL,
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Login Successful",
		"User authenticated successfully",
		user,
	))
}

// VerifySession endpoint
// @Summary Verify user session
// @Description Verifies the current user session and returns user info
// @Tags auth
// @Produce json
// @Success 200 {object} types.APIResponse
// @Failure 401 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/auth/verify [get]
func (h *AuthHandler) VerifySession(c *gin.Context) {
	cookie, err := c.Cookie("session")
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			"No Session",
			"No session found",
		))
		return
	}

	ctx := context.Background()
	token, err := h.client.VerifySessionCookie(ctx, cookie)
	if err != nil {
		log.Printf("Error verifying session: %v", err)
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			"Invalid Session",
			"Session is invalid or expired",
		))
		return
	}

	// Get user info
	userRecord, err := h.client.GetUser(ctx, token.UID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"User Info Failed",
			"Failed to get user info",
		))
		return
	}

	user := &User{
		UID:         userRecord.UID,
		Email:       userRecord.Email,
		DisplayName: userRecord.DisplayName,
		PhotoURL:    userRecord.PhotoURL,
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Session Valid",
		"Session is valid",
		user,
	))
}

// GetUserProfile handles the user profile request
// @Summary Get user profile
// @Description Retrieves the authenticated user's profile information
// @Tags auth
// @Produce json
// @Success 200 {object} types.APIResponse
// @Failure 401 {object} types.APIResponse
// @Router /api/v1/protected/profile [get]
func (h *AuthHandler) GetUserProfile(c *gin.Context) {
	userID := c.GetString("userID")
	userEmail := c.GetString("userEmail")
	displayName := c.GetString("displayName")

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Profile Retrieved",
		"User profile data",
		gin.H{
			"userID":      userID,
			"email":       userEmail,
			"displayName": displayName,
		},
	))
}

// Logout endpoint
// @Summary User logout
// @Description Logs out the user by clearing the session cookie
// @Tags auth
// @Produce json
// @Success 200 {object} types.APIResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear the session cookie
	c.SetCookie(
		"session",
		"",
		-1,
		"/",
		"",   // domain
		true, // secure
		true, // httpOnly
	)

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Logout Successful",
		"User logged out successfully",
		nil,
	))
}

// GetClient returns the Firebase auth client for use in middleware
func (h *AuthHandler) GetClient() *auth.Client {
	return h.client
}
