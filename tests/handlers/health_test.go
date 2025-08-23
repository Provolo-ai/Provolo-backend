package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"provolo-api/internal/handlers"
	"provolo-api/internal/types"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestGetHealthCheck_Success(t *testing.T) {
	// Setup
	router := setupTestRouter()
	
	config := types.Config{
		Environment: "test",
		Port:        8080,
	}
	
	router.GET("/health", handlers.GetHealthCheck(config))

	// Test request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// Check response structure
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Health Check", response["title"])
	assert.Equal(t, "API is running successfully", response["message"])
	
	// Check data fields
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "running", data["uptime"])
	assert.Equal(t, "1.0.0", data["version"])
	assert.Equal(t, "test", data["env"])
	assert.Equal(t, float64(8080), data["port"])
}

func TestGetHealthCheck_ProductionConfig(t *testing.T) {
	// Setup
	router := setupTestRouter()
	
	config := types.Config{
		Environment: "production",
		Port:        443,
	}
	
	router.GET("/health", handlers.GetHealthCheck(config))

	// Test request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "production", data["env"])
	assert.Equal(t, float64(443), data["port"])
}

func TestGetHealthCheck_DevelopmentConfig(t *testing.T) {
	// Setup
	router := setupTestRouter()
	
	config := types.Config{
		Environment: "development",
		Port:        3000,
	}
	
	router.GET("/health", handlers.GetHealthCheck(config))

	// Test request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "development", data["env"])
	assert.Equal(t, float64(3000), data["port"])
}

func TestGetHealthCheck_ResponseFormat(t *testing.T) {
	// Setup
	router := setupTestRouter()
	
	config := types.Config{
		Environment: "test",
		Port:        8080,
	}
	
	router.GET("/health", handlers.GetHealthCheck(config))

	// Test request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify Content-Type header
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	
	// Verify response structure matches types.APIResponse
	var response types.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// Check that all required fields are present
	assert.NotEmpty(t, response.Title)
	assert.NotEmpty(t, response.Message)
	assert.NotEmpty(t, response.Status)
	assert.NotNil(t, response.Data)
}
