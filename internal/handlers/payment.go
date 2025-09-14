package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"provolo-api/internal/types"
	"provolo-api/internal/utils"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
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
		log.Printf("Failed to initialize Firebase for payment tiers request: %v", err)
		errorResponse := types.NewErrorResponse(
			"Service Error",
			"Unable to retrieve pricing information. Please contact support if this continues.",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Get Firestore client
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Printf("Failed to get Firestore client for payment tiers request: %v", err)
		errorResponse := types.NewErrorResponse(
			"Service Error",
			"Unable to retrieve pricing information. Please contact support if this continues.",
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
			log.Printf("Failed to fetch payment tiers from Firestore: %v", err)
			errorResponse := types.NewErrorResponse(
				"Service Error",
				"Unable to retrieve pricing information. Please contact support if this continues.",
			)
			c.JSON(http.StatusInternalServerError, errorResponse)
			return
		}

		var tier types.Tier
		if err := doc.DataTo(&tier); err != nil {
			log.Printf("Failed to map tier data from Firestore: %v", err)
			errorResponse := types.NewErrorResponse(
				"Service Error",
				"Unable to retrieve pricing information. Please contact support if this continues.",
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
		log.Printf("Failed to initialize Firebase for payment tier request (%s): %v", slug, err)
		errorResponse := types.NewErrorResponse(
			"Service Error",
			"Unable to retrieve pricing information. Please contact support if this continues.",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Get Firestore client
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Printf("Failed to get Firestore client for payment tier request (%s): %v", slug, err)
		errorResponse := types.NewErrorResponse(
			"Service Error",
			"Unable to retrieve pricing information. Please contact support if this continues.",
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
		log.Printf("Failed to fetch payment tier (%s) from Firestore: %v", slug, err)
		errorResponse := types.NewErrorResponse(
			"Service Error",
			"Unable to retrieve pricing information. Please contact support if this continues.",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	var tier types.Tier
	if err := doc.DataTo(&tier); err != nil {
		log.Printf("Failed to map payment tier data (%s) from Firestore: %v", slug, err)
		errorResponse := types.NewErrorResponse(
			"Service Error",
			"Unable to retrieve pricing information. Please contact support if this continues.",
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
	var webhookData map[string]interface{}

	if err := c.ShouldBindJSON(&webhookData); err != nil {
		log.Printf("Invalid JSON payload received in payment webhook: %v", err)
		errorResponse := types.NewErrorResponse(
			"Invalid Request",
			"Invalid payment webhook data format.",
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

	// Extract event type and data
	eventType, _ := webhookData["type"].(string)
	data, _ := webhookData["data"].(map[string]interface{})
	checkoutID, _ := data["checkout_id"].(string)
	status, _ := data["status"].(string)
	createdAt, _ := data["created_at"].(string)
	updatedAt, _ := data["modified_at"].(string)
	if updatedAt == "" {
		updatedAt, _ = data["updated_at"].(string)
	}

	// Only persist subscription.created, subscription.updated, order.created, and order.updated
	if eventType == "subscription.created" || eventType == "subscription.updated" || eventType == "order.created" || eventType == "order.updated" {
		app, err := utils.GetFirebaseApp(context.Background())
		if err != nil {
			log.Printf("Failed to initialize Firebase for payment webhook (%s): %v", eventType, err)
			errorResponse := types.NewErrorResponse(
				"Service Error",
				"Unable to process payment event. Please contact support if this continues.",
			)
			c.JSON(http.StatusInternalServerError, errorResponse)
			return
		}
		client, err := app.Firestore(context.Background())
		if err != nil {
			log.Printf("Failed to get Firestore client for payment webhook (%s): %v", eventType, err)
			errorResponse := types.NewErrorResponse(
				"Service Error",
				"Unable to process payment event. Please contact support if this continues.",
			)
			c.JSON(http.StatusInternalServerError, errorResponse)
			return
		}
		defer client.Close()

		// Upsert logic: store all events as keys in a single 'events' map
		docRef := client.Collection("billing_history").Doc(checkoutID)

		// Read existing document to preserve previous events
		events := map[string]interface{}{}
		docSnap, err := docRef.Get(context.Background())
		if err == nil && docSnap.Exists() {
			if existing, err := docSnap.DataAt("events"); err == nil {
				if existingMap, ok := existing.(map[string]interface{}); ok {
					events = existingMap
				}
			}
		}
		events[eventType] = data

		_, err = docRef.Set(context.Background(), map[string]interface{}{
			"checkout_id":    checkoutID,
			"current_status": status,
			"created_at":     createdAt,
			"updated_at":     updatedAt,
			"events":         events,
		}, firestore.MergeAll)
		if err != nil {
			log.Printf("Failed to persist billing event (%s) for checkout %s: %v", eventType, checkoutID, err)
			errorResponse := types.NewErrorResponse(
				"Service Error",
				"Unable to process payment event. Please contact support if this continues.",
			)
			c.JSON(http.StatusInternalServerError, errorResponse)
			return
		}

		// Handle order.updated event for subscription management
		if eventType == "order.updated" {
			if err := handleOrderUpdated(context.Background(), app, data); err != nil {
				fmt.Printf("Failed to process order.updated event: %v\n", err)
			}
		}
	} else {
		// Log all other events but do not persist
		fmt.Printf("Webhook event %s received and logged only.\n", eventType)
	}

	// Return success using the standard APIResponse pattern
	resp := types.NewSuccessResponse(
		"Payment Webhook",
		"Webhook received and processed successfully - any data structure accepted",
		webhookData,
	)
	c.JSON(http.StatusOK, resp)
}

// handleOrderUpdated processes order.updated events for subscription management
func handleOrderUpdated(ctx context.Context, app *firebase.App, data map[string]interface{}) error {
	// Extract user identification
	var userID string
	var customerEmail string

	// Try to get user_id from metadata
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		if uid, exists := metadata["user_id"].(string); exists && uid != "" {
			userID = uid
		}
	}

	// If no user_id in metadata, try customer email
	if userID == "" {
		if customer, ok := data["customer"].(map[string]interface{}); ok {
			if email, exists := customer["email"].(string); exists && email != "" {
				customerEmail = email
			}
		}
	}

	// Skip if we have neither user_id nor customer email
	if userID == "" && customerEmail == "" {
		return fmt.Errorf("no user identification found: missing both metadata.user_id and customer.email")
	}

	// Extract product_id and status
	productID, _ := data["product_id"].(string)
	status, _ := data["status"].(string)

	if productID == "" {
		return fmt.Errorf("missing product_id in order data")
	}

	// Define accepted "paid" statuses
	paidStatuses := map[string]bool{
		"pending":            true,
		"paid":               true,
		"active":             true,
		"completed":          true,
		"refunded":           false, // Include but mark as not paid
		"partially_refunded": false,
		"canceled":           false,
	}

	isPaidStatus, isValidStatus := paidStatuses[status]
	if !isValidStatus {
		return fmt.Errorf("unknown order status: %s", status)
	}

	// Get Firestore client
	client, err := app.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Firestore client: %v", err)
	}
	defer client.Close()

	// Check for double processing - look for existing order with same product_id and status
	checkoutID, _ := data["checkout_id"].(string)
	if checkoutID != "" {
		billingDoc, err := client.Collection("billing_history").Doc(checkoutID).Get(ctx)
		if err == nil && billingDoc.Exists() {
			if events, err := billingDoc.DataAt("events"); err == nil {
				if eventsMap, ok := events.(map[string]interface{}); ok {
					if orderUpdated, exists := eventsMap["order.updated"].(map[string]interface{}); exists {
						if existingStatus, ok := orderUpdated["status"].(string); ok && existingStatus == status {
							// return fmt.Errorf("order already processed with status: %s", status)
						}
					}
				}
			}
		}
	}

	// Find tier by product_id (polarRefId)
	tierQuery := client.Collection("tiers").Where("polarRefId", "==", productID).Limit(1)
	tierDocs, err := tierQuery.Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("failed to query tiers: %v", err)
	}
	if len(tierDocs) == 0 {
		return fmt.Errorf("no tier found with polarRefId: %s", productID)
	}

	var tier types.Tier
	if err := tierDocs[0].DataTo(&tier); err != nil {
		return fmt.Errorf("failed to parse tier data: %v", err)
	}

	// Find user by userID or email
	var userDoc *firestore.DocumentSnapshot
	var userDocRef *firestore.DocumentRef

	if userID != "" {
		// Query by userId field
		userQuery := client.Collection("users").Where("userId", "==", userID).Limit(1)
		userDocs, err := userQuery.Documents(ctx).GetAll()
		if err != nil {
			return fmt.Errorf("failed to query users by userId: %v", err)
		}
		if len(userDocs) > 0 {
			userDoc = userDocs[0]
			userDocRef = userDocs[0].Ref
		}
	}

	if userDoc == nil && customerEmail != "" {
		// Query by email field
		userQuery := client.Collection("users").Where("email", "==", customerEmail).Limit(1)
		userDocs, err := userQuery.Documents(ctx).GetAll()
		if err != nil {
			return fmt.Errorf("failed to query users by email: %v", err)
		}
		if len(userDocs) > 0 {
			userDoc = userDocs[0]
			userDocRef = userDocs[0].Ref
			// Update userID for subsequent operations
			if userData := userDoc.Data(); userData != nil {
				if uid, ok := userData["userId"].(string); ok {
					userID = uid
				}
			}
		}
	}

	if userDoc == nil {
		return fmt.Errorf("user not found with userID: %s or email: %s", userID, customerEmail)
	}

	// Only update tier if status indicates a paid subscription
	if isPaidStatus {
		// Update user's tierId
		_, err = userDocRef.Update(ctx, []firestore.Update{
			{Path: "tierId", Value: tier.Slug},
			{Path: "updatedAt", Value: time.Now()},
		})
		if err != nil {
			return fmt.Errorf("failed to update user tier: %v", err)
		}

		// Archive current quota history before updating to new tier
		if err := archiveQuotaHistory(ctx, client, userID); err != nil {
			fmt.Printf("Warning: Failed to archive quota history for user %s: %v\n", userID, err)
		}

		// Update quota history using existing utility function
		if err := utils.CreateQuotaHistoryFromTier(ctx, app, userID); err != nil {
			return fmt.Errorf("failed to update quota history: %v", err)
		}

		fmt.Printf("Successfully updated user %s to tier %s (status: %s)\n", userID, tier.Slug, status)
	} else {
		// For non-paid statuses (refunded, canceled), potentially downgrade to starter
		// But only if the current tier matches the product being refunded/canceled
		currentUserData := userDoc.Data()
		if currentTierID, ok := currentUserData["tierId"].(string); ok && currentTierID == tier.Slug {
			// Archive current quota history before downgrading
			if err := archiveQuotaHistory(ctx, client, userID); err != nil {
				fmt.Printf("Warning: Failed to archive quota history for user %s before downgrade: %v\n", userID, err)
			}

			// Downgrade to starter tier
			_, err = userDocRef.Update(ctx, []firestore.Update{
				{Path: "tierId", Value: types.DefaultTierID},
				{Path: "updatedAt", Value: time.Now()},
			})
			if err != nil {
				return fmt.Errorf("failed to downgrade user tier: %v", err)
			}

			// Update quota history for starter tier
			if err := utils.CreateQuotaHistoryFromTier(ctx, app, userID); err != nil {
				return fmt.Errorf("failed to update quota history for downgrade: %v", err)
			}

			fmt.Printf("Successfully downgraded user %s to %s tier (status: %s)\n", userID, types.DefaultTierID, status)
		} else {
			fmt.Printf("User %s tier not affected by %s status (current tier: %s, order tier: %s)\n",
				userID, status, currentTierID, tier.Slug)
		}
	}

	return nil
}

// archiveQuotaHistory moves current quota_history to archive collection before tier updates
func archiveQuotaHistory(ctx context.Context, client *firestore.Client, userID string) error {
	// Get current quota history
	quotaDoc, err := client.Collection("quota_history").Doc(userID).Get(ctx)
	if err != nil {
		// If quota history doesn't exist, nothing to archive
		if err.Error() == "not found" || err.Error() == "document not found" {
			fmt.Printf("No existing quota history to archive for user %s\n", userID)
			return nil
		}
		return fmt.Errorf("failed to get current quota history: %v", err)
	}

	if !quotaDoc.Exists() {
		fmt.Printf("No existing quota history to archive for user %s\n", userID)
		return nil
	}

	// Parse current quota history
	var currentQuota types.QuotaHistory
	if err := quotaDoc.DataTo(&currentQuota); err != nil {
		return fmt.Errorf("failed to parse current quota history: %v", err)
	}

	// Get or create archive document
	archiveDocRef := client.Collection("quota_archive").Doc(userID)
	archiveDoc, err := archiveDocRef.Get(ctx)

	var archiveData map[string]interface{}
	now := time.Now()

	if err != nil || !archiveDoc.Exists() {
		// Create new archive document
		archiveData = map[string]interface{}{
			"userId": userID,
			"prev_quotas": []interface{}{
				map[string]interface{}{
					"archived_at":            now,
					"tier_id":                currentQuota.TierId,
					"last_subscription_date": currentQuota.LastSubscriptionDate,
					"features":               currentQuota.Features,
					"created_at":             currentQuota.CreatedAt,
					"updated_at":             currentQuota.UpdatedAt,
				},
			},
			"createdAt": now,
			"updatedAt": now,
		}
	} else {
		// Get existing archive data
		archiveData = archiveDoc.Data()

		// Append new quota history to prev_quotas array
		prevQuotas, ok := archiveData["prev_quotas"].([]interface{})
		if !ok {
			prevQuotas = []interface{}{}
		}

		newQuotaEntry := map[string]interface{}{
			"archived_at":            now,
			"tier_id":                currentQuota.TierId,
			"last_subscription_date": currentQuota.LastSubscriptionDate,
			"features":               currentQuota.Features,
			"created_at":             currentQuota.CreatedAt,
			"updated_at":             currentQuota.UpdatedAt,
		}

		prevQuotas = append(prevQuotas, newQuotaEntry)

		archiveData["prev_quotas"] = prevQuotas
		archiveData["updatedAt"] = now
	}

	// Save archive document
	_, err = archiveDocRef.Set(ctx, archiveData)
	if err != nil {
		return fmt.Errorf("failed to save quota archive: %v", err)
	}

	fmt.Printf("Successfully archived quota history for user %s (tier: %s)\n", userID, currentQuota.TierId)
	return nil
}
