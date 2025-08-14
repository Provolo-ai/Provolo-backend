package handlers

import (
	"net/http"
	"provolo-api/internal/types"

	"github.com/gin-gonic/gin"
)

// PaymentWebhook handles payment webhook requests
// @Summary Handle payment webhook
// @Description Handles payment webhook notifications from payment providers - accepts any JSON structure
// @Tags payments
// @Accept json
// @Produce json
// @Param request body interface{} true "Any JSON data structure"
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
