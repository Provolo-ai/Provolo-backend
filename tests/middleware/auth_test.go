package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"provolo-api/internal/middleware"
)

// Mock Firebase Auth Client
type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Token), args.Error(1)
}

func (m *MockAuthClient) VerifySessionCookie(ctx context.Context, sessionCookie string) (*auth.Token, error) {
	args := m.Called(ctx, sessionCookie)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Token), args.Error(1)
}

// Mock Token
type MockToken struct {
	UID    string
	Claims map[string]interface{}
}

func (m *MockToken) GetUID() string                    { return m.UID }
func (m *MockToken) GetClaims() map[string]interface{} { return m.Claims }
func (m *MockToken) GetAuthTime() int64                { return 0 }
func (m *MockToken) GetIssuedAt() int64                { return 0 }
func (m *MockToken) GetExpirationTime() int64          { return 0 }
func (m *MockToken) GetIssuer() string                 { return "test-issuer" }
func (m *MockToken) GetAudience() string               { return "test-audience" }
func (m *MockToken) GetSubject() string                { return m.UID }
func (m *MockToken) GetFirebase() *auth.FirebaseInfo   { return nil }

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthMiddleware_BearerToken_Success(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	middleware := middleware.AuthMiddleware(mockClient)
	router.Use(middleware)
	
	router.GET("/protected", func(c *gin.Context) {
		userID := c.GetString("userID")
		userEmail := c.GetString("userEmail")
		displayName := c.GetString("userDisplayName")
		
		c.JSON(http.StatusOK, gin.H{
			"userID":      userID,
			"email":       userEmail,
			"displayName": displayName,
		})
	})

	// Mock successful token verification
	mockToken := &MockToken{
		UID: "test-user-123",
		Claims: map[string]interface{}{
			"email": "test@example.com",
			"name":  "Test User",
		},
	}

	mockClient.On("VerifyIDToken", mock.Anything, "valid-bearer-token").Return(mockToken, nil)

	// Test request with Bearer token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-bearer-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "test-user-123", response["userID"])
	assert.Equal(t, "test@example.com", response["email"])
	assert.Equal(t, "Test User", response["displayName"])

	mockClient.AssertExpectations(t)
}

func TestAuthMiddleware_SessionCookie_Success(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	middleware := middleware.AuthMiddleware(mockClient)
	router.Use(middleware)
	
	router.GET("/protected", func(c *gin.Context) {
		userID := c.GetString("userID")
		userEmail := c.GetString("userEmail")
		displayName := c.GetString("userDisplayName")
		
		c.JSON(http.StatusOK, gin.H{
			"userID":      userID,
			"email":       userEmail,
			"displayName": displayName,
		})
	})

	// Mock successful session verification
	mockToken := &MockToken{
		UID: "test-user-123",
		Claims: map[string]interface{}{
			"email": "test@example.com",
			"name":  "Test User",
		},
	}

	mockClient.On("VerifySessionCookie", mock.Anything, "valid-session").Return(mockToken, nil)

	// Test request with session cookie
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "valid-session"})
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "test-user-123", response["userID"])
	assert.Equal(t, "test@example.com", response["email"])
	assert.Equal(t, "Test User", response["displayName"])

	mockClient.AssertExpectations(t)
}

func TestAuthMiddleware_NoAuth_Unauthorized(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	middleware := middleware.AuthMiddleware(mockClient)
	router.Use(middleware)
	
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test request without any authentication
	req, _ := http.NewRequest("GET", "/protected", nil)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Unauthorized", response["title"])
	assert.Equal(t, "No authentication provided", response["message"])
}

func TestAuthMiddleware_InvalidBearerToken_Unauthorized(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	middleware := middleware.AuthMiddleware(mockClient)
	router.Use(middleware)
	
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Mock failed token verification
	mockClient.On("VerifyIDToken", mock.Anything, "invalid-bearer-token").Return(nil, assert.AnError)

	// Test request with invalid Bearer token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-bearer-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Unauthorized", response["title"])
	assert.Equal(t, "Invalid or expired token", response["message"])

	mockClient.AssertExpectations(t)
}

func TestAuthMiddleware_InvalidSessionCookie_Unauthorized(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	middleware := middleware.AuthMiddleware(mockClient)
	router.Use(middleware)
	
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Mock failed session verification
	mockClient.On("VerifySessionCookie", mock.Anything, "invalid-session").Return(nil, assert.AnError)

	// Test request with invalid session cookie
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "invalid-session"})
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Unauthorized", response["title"])
	assert.Equal(t, "Invalid or expired session", response["message"])

	mockClient.AssertExpectations(t)
}

func TestAuthMiddleware_MalformedAuthHeader_Unauthorized(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	middleware := middleware.AuthMiddleware(mockClient)
	router.Use(middleware)
	
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test request with malformed Authorization header
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Unauthorized", response["title"])
	assert.Equal(t, "No authentication provided", response["message"])
}

func TestAuthMiddleware_EmptyAuthHeader_Unauthorized(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	middleware := middleware.AuthMiddleware(mockClient)
	router.Use(middleware)
	
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test request with empty Authorization header
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Unauthorized", response["title"])
	assert.Equal(t, "No authentication provided", response["message"])
}

func TestAuthMiddleware_ClaimsHandling(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	middleware := middleware.AuthMiddleware(mockClient)
	router.Use(middleware)
	
	router.GET("/protected", func(c *gin.Context) {
		userID := c.GetString("userID")
		userEmail := c.GetString("userEmail")
		displayName := c.GetString("userDisplayName")
		
		c.JSON(http.StatusOK, gin.H{
			"userID":      userID,
			"email":       userEmail,
			"displayName": displayName,
		})
	})

	// Mock token with missing claims
	mockToken := &MockToken{
		UID: "test-user-123",
		Claims: map[string]interface{}{
			// No email or name claims
		},
	}

	mockClient.On("VerifyIDToken", mock.Anything, "valid-token").Return(mockToken, nil)

	// Test request
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "test-user-123", response["userID"])
	assert.Equal(t, "", response["email"])
	assert.Equal(t, "", response["displayName"])

	mockClient.AssertExpectations(t)
}
