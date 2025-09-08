package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"provolo-api/internal/types"
	"provolo-api/internal/utils"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
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

// GetPaymentTiers retrieves the available payment tiers
// @Summary Get payment tiers
// @Description Retrieves all available payment tiers
// @Tags payments
// @Produce json
// @Success 200 {object} types.APIResponse{data=[]types.Tier} "Payment tiers retrieved successfully"
// @Failure 500 {object} types.APIResponse "Internal Server Error - Failed to retrieve payment tiers"
// @Router /api/v1/payment/tiers [get]
func GetPaymentTiers(c *gin.Context) {
	ctx := context.Background()

	// Get Firebase app instance
	app, err := utils.GetFirebaseApp(ctx)
	if err != nil {
		errorResponse := types.NewErrorResponse(
			"Firebase Error",
			"Failed to initialize Firebase: "+err.Error(),
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Get Firestore client
	client, err := app.Firestore(ctx)
	if err != nil {
		errorResponse := types.NewErrorResponse(
			"Firestore Error",
			"Failed to get Firestore client: "+err.Error(),
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}
	defer client.Close()

	// Query all tiers from the "tiers" collection
	iter := client.Collection("tiers").Documents(ctx)
	defer iter.Stop()

	var tiers []types.Tier

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			errorResponse := types.NewErrorResponse(
				"Firestore Query Error",
				"Failed to fetch tiers: "+err.Error(),
			)
			c.JSON(http.StatusInternalServerError, errorResponse)
			return
		}

		var tier types.Tier
		if err := doc.DataTo(&tier); err != nil {
			errorResponse := types.NewErrorResponse(
				"Data Mapping Error",
				"Failed to map tier data: "+err.Error(),
			)
			c.JSON(http.StatusInternalServerError, errorResponse)
			return
		}

		tiers = append(tiers, tier)
	}

	// Sort tiers by price (ascending: Starter, Pro, Guru)
	sort.Slice(tiers, func(i, j int) bool {
		return tiers[i].Price < tiers[j].Price
	})

	// Return success response with tiers
	response := types.NewSuccessResponse(
		"Payment Tiers",
		"Payment tiers retrieved successfully",
		tiers,
	)

	c.JSON(http.StatusOK, response)
}

// GetPaymentTierBySlug retrieves a specific payment tier by slug
// @Summary Get payment tier by slug
// @Description Retrieves a specific payment tier using its slug
// @Tags payments
// @Produce json
// @Param slug path string true "Tier slug"
// @Success 200 {object} types.APIResponse{data=types.Tier} "Payment tier retrieved successfully"
// @Failure 404 {object} types.APIResponse "Tier not found"
// @Failure 500 {object} types.APIResponse "Internal Server Error"
// @Router /api/v1/payment/tiers/{slug} [get]
func GetPaymentTierBySlug(c *gin.Context) {
	ctx := context.Background()
	slug := c.Param("slug")

	if slug == "" {
		errorResponse := types.NewErrorResponse(
			"Invalid Request",
			"Tier slug is required",
		)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Get Firebase app instance
	app, err := utils.GetFirebaseApp(ctx)
	if err != nil {
		errorResponse := types.NewErrorResponse(
			"Firebase Error",
			"Failed to initialize Firebase: "+err.Error(),
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Get Firestore client
	client, err := app.Firestore(ctx)
	if err != nil {
		errorResponse := types.NewErrorResponse(
			"Firestore Error",
			"Failed to get Firestore client: "+err.Error(),
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}
	defer client.Close()

	// Query tier by slug
	iter := client.Collection("tiers").Where("slug", "==", slug).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		errorResponse := types.NewErrorResponse(
			"Tier Not Found",
			"No tier found with slug: "+slug,
		)
		c.JSON(http.StatusNotFound, errorResponse)
		return
	}
	if err != nil {
		errorResponse := types.NewErrorResponse(
			"Firestore Query Error",
			"Failed to fetch tier: "+err.Error(),
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	var tier types.Tier
	if err := doc.DataTo(&tier); err != nil {
		errorResponse := types.NewErrorResponse(
			"Data Mapping Error",
			"Failed to map tier data: "+err.Error(),
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Return success response with tier
	response := types.NewSuccessResponse(
		"Payment Tier",
		"Payment tier retrieved successfully",
		tier,
	)

	c.JSON(http.StatusOK, response)
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
// @Router /api/v1/payment/webhook [post]
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

	// Log webhookData as JSON
	webhookJSON, err := json.MarshalIndent(webhookData, "", "  ")
	if err != nil {
		fmt.Printf("Payment Webhook received (marshal error): %+v\n", webhookData)
	} else {
		fmt.Printf("Payment Webhook received: %s\n", webhookJSON)
	}

	app, err := utils.GetFirebaseApp(context.Background())
	if err != nil {
		errorResponse := types.NewErrorResponse(
			"Firebase Error",
			"Failed to initialize Firebase: "+err.Error(),
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	client, err := app.Firestore(context.Background())
	if err != nil {
		errorResponse := types.NewErrorResponse(
			"Firestore Error",
			"Failed to get Firestore client: "+err.Error(),
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}
	defer client.Close()

	docData := map[string]interface{}{
		"payload":     webhookData,
		"received_at": time.Now().UTC(),
	}
	// include raw json string when marshaling succeeded
	if webhookJSON != nil {
		docData["raw_json"] = string(webhookJSON)
	}

	if _, _, err := client.Collection("billing_history").Add(context.Background(), docData); err != nil {
		errorResponse := types.NewErrorResponse(
			"Firestore Write Error",
			"Failed to persist webhook payload: "+err.Error(),
		)
		// Log the error server-side as well
		fmt.Printf("Failed to write billing_history document: %v\n", err)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Return success using the standard APIResponse pattern
	response := types.NewSuccessResponse(
		"Payment Webhook",
		"Webhook received and processed successfully - any data structure accepted",
		webhookData,
	)
	c.JSON(http.StatusOK, response)
}
