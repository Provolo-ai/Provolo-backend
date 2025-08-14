package handlers

import (
	"net/http"
	"provolo-api/internal/types"

	"github.com/gin-gonic/gin"
)

// PaymentWebhookSample represents a sample structure for payment webhook (for documentation only)
type PaymentWebhookSample struct {
	EventType     string                 `json:"event_type" example:"payment.completed"`
	Amount        float64                `json:"amount" example:"100.50"`
	Currency      string                 `json:"currency" example:"USD"`
	TransactionID string                 `json:"transaction_id" example:"txn_123456789"`
	CustomerID    string                 `json:"customer_id" example:"cust_abc123"`
	Status        string                 `json:"status" example:"completed"`
	Timestamp     string                 `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	PaymentMethod string                 `json:"payment_method" example:"credit_card"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// PaymentWebhook handles payment webhook requests
// @Summary Handle payment webhook
// @Description Handles payment webhook notifications from payment providers - accepts any JSON structure
// @Tags payments
// @Accept json
// @Produce json
// @Param request body PaymentWebhookSample true "Sample payment webhook structure (accepts any JSON)"
// @Success 200 {object} types.APIResponse
// @Failure 400 {object} types.APIResponse
// @Router /api/v1/payment-webhook [post]
func PaymentWebhook(c *gin.Context) {
	// Handle completely dynamic JSON data - accepts any structure
	var webhookData interface{}

	if err := c.ShouldBindJSON(&webhookData); err != nil {
		// Return error using the standard APIResponse pattern
		errorResponse := types.NewErrorResponse(
			"Payment Webhook Error",
			"Invalid JSON payload: "+err.Error(),
		)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Return success using the standard APIResponse pattern
	response := types.NewSuccessResponse(
		"Payment Webhook",
		"Webhook received and processed successfully - any data structure accepted",
		webhookData, // Return whatever data was sent
	)

	c.JSON(http.StatusOK, response)
}
