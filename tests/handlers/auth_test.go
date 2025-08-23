package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"provolo-api/internal/handlers"
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

func (m *MockAuthClient) SessionCookie(ctx context.Context, idToken string, expiresIn time.Duration) (string, error) {
	args := m.Called(ctx, idToken, expiresIn)
	return args.String(0), args.Error(1)
}

func (m *MockAuthClient) VerifySessionCookie(ctx context.Context, sessionCookie string) (*auth.Token, error) {
	args := m.Called(ctx, sessionCookie)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Token), args.Error(1)
}

func (m *MockAuthClient) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserRecord), args.Error(1)
}

// Mock User Record
type MockUserRecord struct {
	UID         string
	Email       string
	DisplayName string
	PhotoURL    string
}

func (m *MockUserRecord) GetUID() string                    { return m.UID }
func (m *MockUserRecord) GetEmail() string                  { return m.Email }
func (m *MockUserRecord) GetDisplayName() string            { return m.DisplayName }
func (m *MockUserRecord) GetPhotoURL() string               { return m.PhotoURL }
func (m *MockUserRecord) GetPhoneNumber() string            { return "" }
func (m *MockUserRecord) GetProviderData() []*auth.UserInfo { return nil }
func (m *MockUserRecord) GetCustomClaims() map[string]interface{} {
	return map[string]interface{}{
		"email": m.Email,
		"name":  m.DisplayName,
	}
}
func (m *MockUserRecord) GetDisabled() bool                    { return false }
func (m *MockUserRecord) GetEmailVerified() bool               { return true }
func (m *MockUserRecord) GetCreationTimestamp() time.Time      { return time.Now() }
func (m *MockUserRecord) GetLastSignInTimestamp() time.Time    { return time.Now() }
func (m *MockUserRecord) GetLastRefreshTimestamp() time.Time   { return time.Now() }
func (m *MockUserRecord) GetUserMetadata() *auth.UserMetadata  { return nil }
func (m *MockUserRecord) GetProviderToUserInfo() map[string]*auth.UserInfo {
	return nil
}
func (m *MockUserRecord) GetMultiFactor() *auth.MultiFactor { return nil }
func (m *MockUserRecord) GetTenantID() string               { return "" }

// Mock Token
type MockToken struct {
	UID    string
	Claims map[string]interface{}
}

func (m *MockToken) GetUID() string                    { return m.UID }
func (m *MockToken) GetClaims() map[string]interface{} { return m.Claims }
func (m *MockToken) GetAuthTime() time.Time            { return time.Now() }
func (m *MockToken) GetIssuedAt() time.Time            { return time.Now() }
func (m *MockToken) GetExpirationTime() time.Time      { return time.Now().Add(time.Hour) }
func (m *MockToken) GetIssuer() string                 { return "test-issuer" }
func (m *MockToken) GetAudience() string               { return "test-audience" }
func (m *MockToken) GetSubject() string                { return m.UID }
func (m *MockToken) GetFirebase() *auth.FirebaseInfo   { return nil }

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthHandler_Login_Success(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	handler := &handlers.AuthHandler{Client: mockClient}
	router.POST("/login", handler.Login)

	// Mock successful token verification
	mockToken := &MockToken{
		UID: "test-user-123",
		Claims: map[string]interface{}{
			"email": "test@example.com",
			"name":  "Test User",
		},
	}
	
	mockUserRecord := &MockUserRecord{
		UID:         "test-user-123",
		Email:       "test@example.com",
		DisplayName: "Test User",
		PhotoURL:    "https://example.com/photo.jpg",
	}

	mockClient.On("VerifyIDToken", mock.Anything, "valid-token").Return(mockToken, nil)
	mockClient.On("SessionCookie", mock.Anything, "valid-token", mock.AnythingOfType("time.Duration")).Return("session-cookie", nil)
	mockClient.On("GetUser", mock.Anything, "test-user-123").Return(mockUserRecord, nil)

	// Test request
	loginReq := handlers.LoginRequest{IdToken: "valid-token"}
	jsonData, _ := json.Marshal(loginReq)
	
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Login Successful", response["title"])
	assert.Equal(t, "User authenticated successfully", response["message"])
	
	// Check if session cookie was set
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}
	assert.NotNil(t, sessionCookie)
	assert.Equal(t, "session-cookie", sessionCookie.Value)
	assert.True(t, sessionCookie.HttpOnly)

	mockClient.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidToken(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	handler := &handlers.AuthHandler{Client: mockClient}
	router.POST("/login", handler.Login)

	// Mock failed token verification
	mockClient.On("VerifyIDToken", mock.Anything, "invalid-token").Return(nil, assert.AnError)

	// Test request
	loginReq := handlers.LoginRequest{IdToken: "invalid-token"}
	jsonData, _ := json.Marshal(loginReq)
	
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Authentication Failed", response["title"])

	mockClient.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidRequest(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	handler := &handlers.AuthHandler{Client: mockClient}
	router.POST("/login", handler.Login)

	// Test request with invalid JSON
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Invalid Request", response["title"])
}

func TestAuthHandler_VerifySession_Success(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	handler := &handlers.AuthHandler{Client: mockClient}
	router.GET("/verify", handler.VerifySession)

	// Mock successful session verification
	mockToken := &MockToken{
		UID: "test-user-123",
		Claims: map[string]interface{}{
			"email": "test@example.com",
			"name":  "Test User",
		},
	}
	
	mockUserRecord := &MockUserRecord{
		UID:         "test-user-123",
		Email:       "test@example.com",
		DisplayName: "Test User",
		PhotoURL:    "https://example.com/photo.jpg",
	}

	mockClient.On("VerifySessionCookie", mock.Anything, "valid-session").Return(mockToken, nil)
	mockClient.On("GetUser", mock.Anything, "test-user-123").Return(mockUserRecord, nil)

	// Test request
	req, _ := http.NewRequest("GET", "/verify", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "valid-session"})
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Session Valid", response["title"])
	assert.Equal(t, "Session is valid", response["message"])

	mockClient.AssertExpectations(t)
}

func TestAuthHandler_VerifySession_NoSession(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	handler := &handlers.AuthHandler{Client: mockClient}
	router.GET("/verify", handler.VerifySession)

	// Test request without session cookie
	req, _ := http.NewRequest("GET", "/verify", nil)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "No Session", response["title"])
}

func TestAuthHandler_VerifySession_InvalidSession(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	handler := &handlers.AuthHandler{Client: mockClient}
	router.POST("/logout", handler.Logout)

	// Mock failed session verification
	mockClient.On("VerifySessionCookie", mock.Anything, "invalid-session").Return(nil, assert.AnError)

	// Test request
	req, _ := http.NewRequest("POST", "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "invalid-session"})
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Invalid Session", response["title"])

	mockClient.AssertExpectations(t)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	// Setup
	router := setupTestRouter()
	mockClient := &MockAuthClient{}
	
	handler := &handlers.AuthHandler{Client: mockClient}
	router.POST("/logout", handler.Logout)

	// Test request
	req, _ := http.NewRequest("POST", "/logout", nil)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Logout Successful", response["title"])
	assert.Equal(t, "User logged out successfully", response["message"])
	
	// Check if session cookie was cleared
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}
	assert.NotNil(t, sessionCookie)
	assert.Equal(t, "", sessionCookie.Value)
	assert.Equal(t, -1, sessionCookie.MaxAge)
}

func TestAuthHandler_NewAuthHandler_Error(t *testing.T) {
	// Test with invalid Firebase config
	// This would require mocking the Firebase initialization which is complex
	// For now, we'll test the basic structure
	
	// The actual NewAuthHandler function requires Firebase credentials
	// In a real test environment, you would use test credentials or mock the Firebase SDK
	t.Skip("Skipping NewAuthHandler test as it requires Firebase credentials")
}
