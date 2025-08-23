package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"provolo-api/internal/routes"
	"provolo-api/internal/types"
)

func setupTestConfig() *types.Config {
	return &types.Config{
		Environment: "test",
		Port:        8080,
		SwaggerURL:  "http://localhost:8080/swagger/doc.json",
	}
}

func TestSetupRoutes_HealthEndpoint(t *testing.T) {
	// Setup
	config := setupTestConfig()
	handler := routes.SetupRoutes(config)

	// Test health endpoint
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Health Check", response["title"])
	assert.Equal(t, "API is running successfully", response["message"])
}

func TestSetupRoutes_SwaggerRedirect(t *testing.T) {
	// Setup
	config := setupTestConfig()
	handler := routes.SetupRoutes(config)

	// Test swagger redirect
	req, _ := http.NewRequest("GET", "/swagger/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusFound, w.Code) // 302 redirect
	assert.Equal(t, "/swagger/index.html", w.Header().Get("Location"))
}

func TestSetupRoutes_ProtectedProfileEndpoint_Unauthorized(t *testing.T) {
	// Setup
	config := setupTestConfig()
	handler := routes.SetupRoutes(config)

	// Test protected profile endpoint without auth
	req, _ := http.NewRequest("GET", "/api/v1/protected/profile", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Unauthorized", response["title"])
}

func TestSetupRoutes_OptimizeProfileEndpoint_Unauthorized(t *testing.T) {
	// Setup
	config := setupTestConfig()
	handler := routes.SetupRoutes(config)

	// Test optimize profile endpoint without auth
	req, _ := http.NewRequest("POST", "/api/v1/optimize-profile", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Unauthorized", response["title"])
}

func TestSetupRoutes_PaymentWebhookEndpoint(t *testing.T) {
	// Setup
	config := setupTestConfig()
	handler := routes.SetupRoutes(config)

	// Test payment webhook endpoint
	req, _ := http.NewRequest("POST", "/api/v1/payment-webhook", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Payment Webhook", response["title"])
}

func TestSetupRoutes_ProductionCORS(t *testing.T) {
	// Setup production config
	config := &types.Config{
		Environment: "production",
		Port:        443,
		SwaggerURL:  "https://provolo.org/swagger/doc.json",
	}
	handler := routes.SetupRoutes(config)

	// Test that CORS headers are set for production
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	req.Header.Set("Origin", "https://provolo.org")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Check CORS headers
	corsHeader := w.Header().Get("Access-Control-Allow-Origin")
	assert.NotEmpty(t, corsHeader)
}

func TestSetupRoutes_DevelopmentCORS(t *testing.T) {
	// Setup development config
	config := &types.Config{
		Environment: "development",
		Port:        3000,
		SwaggerURL:  "http://localhost:3000/swagger/doc.json",
	}
	handler := routes.SetupRoutes(config)

	// Test that CORS allows all origins in development
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Check CORS headers
	corsHeader := w.Header().Get("Access-Control-Allow-Origin")
	assert.NotEmpty(t, corsHeader)
}

func TestSetupRoutes_NonExistentEndpoint(t *testing.T) {
	// Setup
	config := setupTestConfig()
	handler := routes.SetupRoutes(config)

	// Test non-existent endpoint
	req, _ := http.NewRequest("GET", "/api/v1/nonexistent", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetupRoutes_InvalidMethod(t *testing.T) {
	// Setup
	config := setupTestConfig()
	handler := routes.SetupRoutes(config)

	// Test invalid method for health endpoint
	req, _ := http.NewRequest("POST", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetupRoutes_ApiVersioning(t *testing.T) {
	// Setup
	config := setupTestConfig()
	handler := routes.SetupRoutes(config)

	// Test that all endpoints are under /api/v1
	endpoints := []string{
		"/api/v1/health",
		"/api/v1/auth/login",
		"/api/v1/auth/verify",
		"/api/v1/auth/logout",
		"/api/v1/protected/profile",
		"/api/v1/optimize-profile",
		"/api/v1/payment-webhook",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req, _ := http.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			// Should not return 404 (endpoint exists)
			assert.NotEqual(t, http.StatusNotFound, w.Code, "Endpoint %s should exist", endpoint)
		})
	}
}

func TestSetupRoutes_ResponseHeaders(t *testing.T) {
	// Setup
	config := setupTestConfig()
	handler := routes.SetupRoutes(config)

	// Test health endpoint
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// Check common headers
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	
	// Check that CORS headers are present
	corsHeader := w.Header().Get("Access-Control-Allow-Origin")
	assert.NotEmpty(t, corsHeader)
}
