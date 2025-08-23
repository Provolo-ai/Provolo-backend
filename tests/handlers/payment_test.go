package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"provolo-api/internal/handlers"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestPaymentWebhook_Success(t *testing.T) {
	// Setup
	router := setupTestRouter()
	router.POST("/payment-webhook", handlers.PaymentWebhook)

	// Test request with valid JSON
	webhookData := map[string]interface{}{
		"event_type":     "payment.completed",
		"amount":         100.50,
		"currency":       "USD",
		"transaction_id": "txn_123456789",
		"customer_id":    "cust_abc123",
		"status":         "completed",
		"timestamp":      "2024-01-15T10:30:00Z",
		"payment_method": "credit_card",
		"metadata": map[string]interface{}{
			"order_id": "order_123",
			"user_id":  "user_456",
		},
	}

	jsonData, _ := json.Marshal(webhookData)

	req, _ := http.NewRequest("POST", "/payment-webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check response structure
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Payment Webhook", response["title"])
	assert.Equal(t, "Webhook received and processed successfully - any data structure accepted", response["message"])

	// Check that the data was returned as expected
	responseData := response["data"].(map[string]interface{})
	assert.Equal(t, "payment.completed", responseData["event_type"])
	assert.Equal(t, 100.50, responseData["amount"])
	assert.Equal(t, "USD", responseData["currency"])
	assert.Equal(t, "txn_123456789", responseData["transaction_id"])
}

func TestPaymentWebhook_SimpleData(t *testing.T) {
	// Setup
	router := setupTestRouter()
	router.POST("/payment-webhook", handlers.PaymentWebhook)

	// Test request with simple data
	webhookData := map[string]interface{}{
		"status": "success",
		"id":     "webhook_123",
	}

	jsonData, _ := json.Marshal(webhookData)

	req, _ := http.NewRequest("POST", "/payment-webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	
	// Check that the simple data was returned
	responseData := response["data"].(map[string]interface{})
	assert.Equal(t, "success", responseData["status"])
	assert.Equal(t, "webhook_123", responseData["id"])
}

func TestPaymentWebhook_EmptyData(t *testing.T) {
	// Setup
	router := setupTestRouter()
	router.POST("/payment-webhook", handlers.PaymentWebhook)

	// Test request with empty data
	webhookData := map[string]interface{}{}

	jsonData, _ := json.Marshal(webhookData)

	req, _ := http.NewRequest("POST", "/payment-webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	
	// Check that empty data was returned
	responseData := response["data"].(map[string]interface{})
	assert.Empty(t, responseData)
}

func TestPaymentWebhook_InvalidJSON(t *testing.T) {
	// Setup
	router := setupTestRouter()
	router.POST("/payment-webhook", handlers.PaymentWebhook)

	// Test request with invalid JSON
	req, _ := http.NewRequest("POST", "/payment-webhook", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Payment Webhook Error", response["title"])
	assert.Contains(t, response["message"], "Invalid JSON payload")
}

func TestPaymentWebhook_MalformedJSON(t *testing.T) {
	// Setup
	router := setupTestRouter()
	router.POST("/payment-webhook", handlers.PaymentWebhook)

	// Test request with malformed JSON
	req, _ := http.NewRequest("POST", "/payment-webhook", bytes.NewBuffer([]byte(`{"key": "value",}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Payment Webhook Error", response["title"])
	assert.Contains(t, response["message"], "Invalid JSON payload")
}

func TestPaymentWebhook_NoContentType(t *testing.T) {
	// Setup
	router := setupTestRouter()
	router.POST("/payment-webhook", handlers.PaymentWebhook)

	// Test request without Content-Type header
	webhookData := map[string]interface{}{
		"status": "success",
	}

	jsonData, _ := json.Marshal(webhookData)

	req, _ := http.NewRequest("POST", "/payment-webhook", bytes.NewBuffer(jsonData))
	// No Content-Type header

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
}

func TestPaymentWebhook_EmptyBody(t *testing.T) {
	// Setup
	router := setupTestRouter()
	router.POST("/payment-webhook", handlers.PaymentWebhook)

	// Test request with empty body
	req, _ := http.NewRequest("POST", "/payment-webhook", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	
	// Check that nil data was returned
	assert.Nil(t, response["data"])
}

func TestPaymentWebhook_ResponseFormat(t *testing.T) {
	// Setup
	router := setupTestRouter()
	router.POST("/payment-webhook", handlers.PaymentWebhook)

	// Test request
	webhookData := map[string]interface{}{
		"test": "data",
	}

	jsonData, _ := json.Marshal(webhookData)

	req, _ := http.NewRequest("POST", "/payment-webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify Content-Type header
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	// Verify response structure
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check that all required fields are present
	assert.NotEmpty(t, response["title"])
	assert.NotEmpty(t, response["message"])
	assert.NotEmpty(t, response["status"])
	assert.NotNil(t, response["data"])
}
