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

	"cloud.google.com/go/firestore"
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
	TierID      string `json:"tierId"`
}

// UserData represents the full user data from Firestore
type UserData struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	TierID      string    `json:"tierId"`
	Subscribed  bool      `json:"subscribed"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// getOrCreateUserData retrieves user data from Firestore or creates it if missing
func (h *AuthHandler) getOrCreateUserData(ctx context.Context, userRecord *auth.UserRecord, createIfMissing bool) (*UserData, error) {
	// Get Firebase app and Firestore client
	app, err := utils.GetFirebaseApp(ctx)
	if err != nil {
		log.Printf("Failed to initialize Firebase app for user %s: %v", userRecord.UID, err)
		return nil, fmt.Errorf("database connection error")
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Printf("Failed to get Firestore client for user %s: %v", userRecord.UID, err)
		return nil, fmt.Errorf("database connection error")
	}
	defer client.Close()

	// Query user by Firebase UID
	userQuery := client.Collection("users").Where("userId", "==", userRecord.UID).Limit(1)
	docs, err := userQuery.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Failed to query users collection for user %s: %v", userRecord.UID, err)
		return nil, fmt.Errorf("database query error")
	}

	// If user exists, return the data
	if len(docs) > 0 {
		doc := docs[0]
		data := doc.Data()

		userData := &UserData{
			ID:          doc.Ref.ID,
			UserID:      userRecord.UID,
			Email:       userRecord.Email,
			DisplayName: userRecord.DisplayName,
			TierID:      types.DefaultTierID, // default
			Subscribed:  true,
		}

		// Map Firestore data
		if tierID, ok := data["tierId"].(string); ok {
			userData.TierID = tierID
		}
		if subscribed, ok := data["subscribed"].(bool); ok {
			userData.Subscribed = subscribed
		}
		if createdAt, ok := data["createdAt"].(time.Time); ok {
			userData.CreatedAt = createdAt
		}
		if updatedAt, ok := data["updatedAt"].(time.Time); ok {
			userData.UpdatedAt = updatedAt
		}

		return userData, nil
	}

	// User doesn't exist in Firestore
	if !createIfMissing {
		return nil, fmt.Errorf("user not found in database")
	}

	// Create new user with starter tier (for auto-creation scenarios)
	log.Printf("Auto-creating user data for Firebase user: %s", userRecord.UID)

	newUserData := map[string]interface{}{
		"userId":      userRecord.UID,
		"email":       userRecord.Email,
		"displayName": userRecord.DisplayName,
		"tierId":      types.DefaultTierID,
		"subscribed":  true,
		"createdAt":   time.Now(),
		"updatedAt":   time.Now(),
	}

	// Add user to Firestore
	docRef, _, err := client.Collection("users").Add(ctx, newUserData)
	if err != nil {
		log.Printf("Failed to create user in database for UID %s: %v", userRecord.UID, err)
		return nil, fmt.Errorf("user creation error")
	}

	// Create quota history for new user
	if err := h.createQuotaHistoryForUser(ctx, client, userRecord.UID); err != nil {
		log.Printf("Warning: Failed to create quota history for auto-created user %s: %v", userRecord.UID, err)
	}

	return &UserData{
		ID:          docRef.ID,
		UserID:      userRecord.UID,
		Email:       userRecord.Email,
		DisplayName: userRecord.DisplayName,
		TierID:      types.DefaultTierID,
		Subscribed:  true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// createQuotaHistoryForUser creates quota history for a user (helper function)
func (h *AuthHandler) createQuotaHistoryForUser(ctx context.Context, client *firestore.Client, userID string) error {
	// Build tier features from the tiers collection
	features := []types.QuotaFeature{}
	tierDoc, err := client.Collection("tiers").Doc(types.DefaultTierID).Get(ctx)
	if err == nil {
		if feats, ok := tierDoc.Data()["features"].([]interface{}); ok {
			for _, f := range feats {
				if fmap, ok := f.(map[string]interface{}); ok {
					feature := types.Feature{
						Name:              fmap["name"].(string),
						Description:       fmap["description"].(string),
						Slug:              types.FeatureSlug(fmap["slug"].(string)),
						Limited:           fmap["limited"].(bool),
						MaxQuota:          int(fmap["maxQuota"].(int64)),
						RecurringInterval: types.RecurringInterval(fmap["recurringInterval"].(string)),
					}

					features = append(features, types.QuotaFeature{
						Feature:    feature,
						UsageCount: 0,
						LastUsed:   nil,
					})
				}
			}
		}
	}

	// Create quota history document
	quotaDoc := client.Collection("quota_history").Doc(userID)
	quotaHistory := types.QuotaHistory{
		UserId:               userID,
		TierId:               types.DefaultTierID,
		LastSubscriptionDate: time.Now(),
		Features:             features,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Save quota history
	_, err = quotaDoc.Set(ctx, quotaHistory)
	return err
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
		log.Printf("Invalid login request body: %v", err)
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"Invalid Request",
			"Please check your request format and try again.",
		))
		return
	}

	ctx := context.Background()

	// Verify the ID token
	token, err := h.client.VerifyIDToken(ctx, req.IdToken)
	if err != nil {
		log.Printf("Error verifying ID token for login: %v", err)
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			"Authentication Failed",
			"Unable to verify your authentication. Please try logging in again.",
		))
		return
	}

	// Create session cookie (expires in 5 days)
	expiresIn := time.Hour * 24 * 5
	cookie, err := h.client.SessionCookie(ctx, req.IdToken, expiresIn)
	if err != nil {
		log.Printf("Error creating session cookie for user %s: %v", token.UID, err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"Session Creation Failed",
			"Unable to create your session. Please contact support if this continues.",
		))
		return
	}

	// Set HTTP-only cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:   "session",
		Value:  cookie,
		MaxAge: int(expiresIn.Seconds()),
		Path:   "/",
		// Domain:   "provolo-front-end-dev-env.vercel.app",
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	})

	// Get user info from Firebase Auth
	userRecord, err := h.client.GetUser(ctx, token.UID)
	if err != nil {
		log.Printf("Error getting user from Firebase Auth during login for UID %s: %v", token.UID, err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"User Info Failed",
			"Unable to retrieve your account information. Please contact support if this continues.",
		))
		return
	}

	// Get or create user data in Firestore (strict mode - don't auto-create for login)
	userData, err := h.getOrCreateUserData(ctx, userRecord, false)
	if err != nil {
		log.Printf("User not found in database: %s, Error: %v", userRecord.UID, err)
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			"Account Setup Required",
			"Your account is not properly set up. Please contact support or sign up again to complete your account setup.",
		))
		return
	}

	user := &User{
		UID:         userRecord.UID,
		Email:       userRecord.Email,
		DisplayName: userRecord.DisplayName,
		PhotoURL:    userRecord.PhotoURL,
		TierID:      userData.TierID,
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

	// Get user info from Firebase Auth
	userRecord, err := h.client.GetUser(ctx, token.UID)
	if err != nil {
		log.Printf("Error getting user from Firebase Auth during session verification for UID %s: %v", token.UID, err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"User Info Failed",
			"Unable to retrieve your account information. Please contact support if this continues.",
		))
		return
	}

	// Get or create user data in Firestore (strict mode - don't auto-create for verify)
	userData, err := h.getOrCreateUserData(ctx, userRecord, false)
	if err != nil {
		log.Printf("User not found in database: %s, Error: %v", userRecord.UID, err)
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			"Account Setup Required",
			"Your account is not properly set up. Please contact support or sign up again to complete your account setup.",
		))
		return
	}

	user := &User{
		UID:         userRecord.UID,
		Email:       userRecord.Email,
		DisplayName: userRecord.DisplayName,
		PhotoURL:    userRecord.PhotoURL,
		TierID:      userData.TierID,
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
	http.SetCookie(c.Writer, &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: int(-1),
		Path:   "/",
		// Domain:   "provolo-front-end-dev-env.vercel.app",
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	})

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Logout Successful",
		"User logged out successfully",
		nil,
	))
}

// SignupOrEnsureUser endpoint - ensures user exists and sets up starter tier
// @Summary User signup or ensure existence
// @Description Checks if user exists in Firestore and creates profile with starter tier if not
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Firebase ID token"
// @Success 200 {object} types.APIResponse
// @Failure 400 {object} types.APIResponse
// @Failure 401 {object} types.APIResponse
// @Failure 500 {object} types.APIResponse
// @Router /api/v1/auth/signup [post]
func (h *AuthHandler) SignupOrEnsureUser(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid signup request body: %v", err)
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"Invalid Request",
			"Please check your request format and try again.",
		))
		return
	}

	// Verify the ID token
	token, err := h.client.VerifyIDToken(context.Background(), req.IdToken)
	if err != nil {
		log.Printf("Error verifying ID token during signup: %v", err)
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			"Authentication Failed",
			"Unable to verify your authentication. Please try again.",
		))
		return
	}

	userID := token.UID

	// Get Firebase user record
	ctx := context.Background()
	userRecord, err := h.client.GetUser(ctx, userID)
	if err != nil {
		log.Printf("Error getting user from Firebase Auth during signup for UID %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"User Info Failed",
			"Unable to retrieve your account information. Please contact support if this continues.",
		))
		return
	}

	// Get or create user data in Firestore (signup mode - allow auto-creation)
	userData, err := h.getOrCreateUserData(ctx, userRecord, true)
	if err != nil {
		log.Printf("Error creating or retrieving user data during signup for UID %s: %v", userRecord.UID, err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"Database Error",
			"Unable to set up your account. Please contact support if this continues.",
		))
		return
	}

	// Determine if this was a new user creation
	isNewUser := userData.CreatedAt.After(time.Now().Add(-time.Minute))

	// Create session cookie
	expiresIn := time.Hour * 24 * 5
	cookie, err := h.client.SessionCookie(context.Background(), req.IdToken, expiresIn)
	if err != nil {
		log.Printf("Error creating session cookie during signup for user %s: %v", userRecord.UID, err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"Session Error",
			"Unable to create your session. Please contact support if this continues.",
		))
		return
	}

	// Set the session cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session",
		Value:    cookie,
		MaxAge:   int(expiresIn.Seconds()),
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	})

	message := "User profile retrieved"
	if isNewUser {
		message = "New user created with starter tier and quota history"
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Signup/Login Successful",
		message,
		userData,
	))
}

// GetClient returns the Firebase auth client for use in middleware
func (h *AuthHandler) GetClient() *auth.Client {
	return h.client
}
