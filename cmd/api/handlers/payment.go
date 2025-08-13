package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PaymentWebhook handles the payment webhook
// @Summary Handle payment webhook
// @Description Handles payment webhook events from the payment gateway
// @Tags payments
// @Accept json
// @Produce json
// @Param webhook body map[string]interface{} true "Payment Webhook Data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/v1/payment-webhook [post]
func PaymentWebhook(c *gin.Context) {
	// Handle dynamic JSON data - perfect for webhooks!
	var webhookData map[string]interface{}

	if err := c.ShouldBindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	// Log the received webhook data for testing
	log.Printf("Webhook received: %+v\n", webhookData)

	// You can access any field dynamically:
	if eventType, exists := webhookData["event_type"]; exists {
		log.Printf("Event type: %v\n", eventType)
	}

	if amount, exists := webhookData["amount"]; exists {
		log.Printf("Amount: %v\n", amount)
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"message":       "Payment webhook received successfully",
		"received_data": webhookData,
		"timestamp":     fmt.Sprintf("%v", webhookData["timestamp"]),
	})
}
