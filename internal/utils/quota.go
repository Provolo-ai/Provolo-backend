package utils

import (
	"context"
	"fmt"
	"log"
	"provolo-api/internal/types"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
)

func ResetIfNewInterval(feature types.QuotaFeature, now time.Time) int {
	if feature.LastUsed == nil {
		return feature.UsageCount
	}

	switch feature.RecurringInterval {
	case types.Daily:
		if !isSameDay(*feature.LastUsed, now) {
			return 0
		}
	case types.Weekly:
		_, week1 := feature.LastUsed.ISOWeek()
		_, week2 := now.ISOWeek()
		if week1 != week2 || feature.LastUsed.Year() != now.Year() {
			return 0
		}
	case types.Monthly:
		y1, m1, _ := feature.LastUsed.Date()
		y2, m2, _ := now.Date()
		if y1 != y2 || m1 != m2 {
			return 0
		}
	}
	return feature.UsageCount
}

// QuotaResult represents the result of a quota check
type QuotaResult struct {
	Allowed bool `json:"allowed"`
	Count   int  `json:"count"`
	Limit   int  `json:"limit"`
}

// CheckUserQuota checks if user has quota available for a specific feature
func CheckUserQuota(ctx context.Context, app *firebase.App, userID string, slug types.FeatureSlug) (*QuotaResult, error) {
	// Log quota check attempt
	logQuotaAction(ctx, userID, "QUOTA_CHECK_ATTEMPT", fmt.Sprintf("Feature: %s", slug))

	client, err := app.Firestore(ctx)
	if err != nil {
		logQuotaAction(ctx, userID, "QUOTA_CHECK_CLIENT_FAILED", fmt.Sprintf("Feature: %s, Error: %v", slug, err))
		return nil, fmt.Errorf("failed to get Firestore client: %v", err)
	}
	defer client.Close()

	// Try to get existing quota history
	doc, err := client.Collection("quota_history").Doc(userID).Get(ctx)
	if err != nil || !doc.Exists() {
		// User doesn't have quota history, create it from their tier
		logQuotaAction(ctx, userID, "QUOTA_HISTORY_MISSING", fmt.Sprintf("Feature: %s, Creating new quota history", slug))

		// Add comprehensive debugging to help troubleshoot
		debugUserExistence(ctx, client, userID)
		debugFirestoreStructure(ctx, client, userID)

		// Check if user exists before trying to create quota history
		userExists, err := checkUserExistsInUsers(ctx, client, userID)
		if err != nil {
			logQuotaAction(ctx, userID, "USER_EXISTENCE_CHECK_FAILED", fmt.Sprintf("Error: %v", err))
			return nil, fmt.Errorf("failed to check user existence: %v", err)
		}

		if !userExists {
			logQuotaAction(ctx, userID, "USER_NOT_FOUND_IN_USERS", "User does not exist in users collection")
			return nil, fmt.Errorf("user not found in users collection: %s", userID)
		}

		return createQuotaHistoryFromTier(ctx, client, userID, slug)
	}

	var quotaHistory types.QuotaHistory
	if err := doc.DataTo(&quotaHistory); err != nil {
		return nil, fmt.Errorf("failed to parse quota history: %v", err)
	}

	// Find the specific feature
	for _, feature := range quotaHistory.Features {
		if feature.Slug == slug {
			now := time.Now()
			currentCount := ResetIfNewInterval(feature, now)

			result := &QuotaResult{
				Allowed: currentCount < feature.MaxQuota,
				Count:   currentCount,
				Limit:   feature.MaxQuota,
			}

			return result, nil
		}
	}

	// Feature not found in quota history
	return nil, fmt.Errorf("feature %s not found in quota history for user %s", slug, userID)
}

// createQuotaHistoryFromTier creates quota history for a user based on their tier
func createQuotaHistoryFromTier(ctx context.Context, client *firestore.Client, userID string, slug types.FeatureSlug) (*QuotaResult, error) {
	// Get user's tier
	query := client.Collection("users").Where("userId", "==", userID).Limit(1)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get user document: %v", err)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	userDoc := docs[0]

	tierID := "starter" // default tier
	if tid, ok := userDoc.Data()["tierId"].(string); ok && tid != "" {
		tierID = tid
	}

	// Get tier features
	tierDoc, err := client.Collection("tiers").Doc(tierID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tier document: %v", err)
	}

	var tier types.Tier
	if err := tierDoc.DataTo(&tier); err != nil {
		return nil, fmt.Errorf("failed to parse tier document: %v", err)
	}

	// Find the specific feature
	var targetFeature *types.Feature
	for _, feature := range tier.Features {
		if feature.Slug == slug {
			targetFeature = &feature
			break
		}
	}

	if targetFeature == nil {
		return nil, fmt.Errorf("feature %s not found in tier %s", slug, tierID)
	}

	// Create quota history with all tier features
	features := make([]types.QuotaFeature, len(tier.Features))
	for i, feature := range tier.Features {
		features[i] = types.QuotaFeature{
			Feature:    feature,
			UsageCount: 0,
			LastUsed:   nil,
		}
	}

	quotaHistory := types.QuotaHistory{
		UserId:               userID,
		TierId:               tierID,
		LastSubscriptionDate: time.Now(),
		Features:             features,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Save quota history
	_, err = client.Collection("quota_history").Doc(userID).Set(ctx, quotaHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to create quota history: %v", err)
	}

	// Return result for the requested feature
	return &QuotaResult{
		Allowed: true, // New user, first usage
		Count:   0,
		Limit:   targetFeature.MaxQuota,
	}, nil
}

// UpdateUserQuota increments the user's quota usage for a specific feature
func UpdateUserQuota(ctx context.Context, app *firebase.App, userID string, slug types.FeatureSlug) error {
	// Log quota update attempt
	logQuotaAction(ctx, userID, "QUOTA_UPDATE_ATTEMPT", fmt.Sprintf("Feature: %s", slug))

	client, err := app.Firestore(ctx)
	if err != nil {
		logQuotaAction(ctx, userID, "QUOTA_UPDATE_CLIENT_FAILED", fmt.Sprintf("Feature: %s, Error: %v", slug, err))
		return fmt.Errorf("failed to get Firestore client: %v", err)
	}
	defer client.Close()

	doc, err := client.Collection("quota_history").Doc(userID).Get(ctx)
	if err != nil || !doc.Exists() {
		return fmt.Errorf("no quota history found for user %s", userID)
	}

	var quotaHistory types.QuotaHistory
	if err := doc.DataTo(&quotaHistory); err != nil {
		return fmt.Errorf("failed to parse quota history: %v", err)
	}

	now := time.Now()
	updatedFeatures := quotaHistory.Features
	var newCount int

	for i, f := range updatedFeatures {
		if f.Slug == slug {
			resetCount := ResetIfNewInterval(f, now)
			updatedFeatures[i].UsageCount = resetCount + 1
			updatedFeatures[i].LastUsed = &now
			newCount = updatedFeatures[i].UsageCount
			break
		}
	}

	_, err = client.Collection("quota_history").Doc(userID).Set(ctx, map[string]interface{}{
		"features":  updatedFeatures,
		"updatedAt": now,
	}, firestore.MergeAll)

	if err != nil {
		logQuotaAction(ctx, userID, "QUOTA_UPDATE_FAILED", fmt.Sprintf("Feature: %s, Error: %v", slug, err))
		return fmt.Errorf("failed to update quota: %v", err)
	}

	// Log successful quota update
	logQuotaAction(ctx, userID, "QUOTA_UPDATE_SUCCESS", fmt.Sprintf("Feature: %s, New count: %d", slug, newCount))
	return nil
}

// CheckAndUpdateQuota checks quota and updates it if allowed
func CheckAndUpdateQuota(ctx context.Context, app *firebase.App, userID string, slug types.FeatureSlug) (*QuotaResult, error) {
	result, err := CheckUserQuota(ctx, app, userID, slug)
	if err != nil {
		return nil, err
	}

	if !result.Allowed {
		return result, nil
	}

	if err := UpdateUserQuota(ctx, app, userID, slug); err != nil {
		return nil, err
	}

	result.Count += 1
	return result, nil
}

// CreateQuotaHistoryFromTier creates quota history for a user based on their tier
// This is a public function that can be reused in migrations or other parts of the app
func CreateQuotaHistoryFromTier(ctx context.Context, app *firebase.App, userID string) error {
	client, err := app.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Firestore client: %v", err)
	}
	defer client.Close()

	// Get user's tier
	query := client.Collection("users").Where("userId", "==", userID).Limit(1)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("failed to get user document: %v", err)
	}
	if len(docs) == 0 {
		return fmt.Errorf("user not found: %s", userID)
	}
	userDoc := docs[0]

	tierID := "starter" // default tier
	if tid, ok := userDoc.Data()["tierId"].(string); ok && tid != "" {
		tierID = tid
	}

	// Get tier features
	tierDoc, err := client.Collection("tiers").Doc(tierID).Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tier document: %v", err)
	}

	var tier types.Tier
	if err := tierDoc.DataTo(&tier); err != nil {
		return fmt.Errorf("failed to parse tier document: %v", err)
	}

	// Create quota history with all tier features
	features := make([]types.QuotaFeature, len(tier.Features))
	for i, feature := range tier.Features {
		features[i] = types.QuotaFeature{
			Feature:    feature,
			UsageCount: 0,
			LastUsed:   nil,
		}
	}

	quotaHistory := types.QuotaHistory{
		UserId:               userID,
		TierId:               tierID,
		LastSubscriptionDate: time.Now(),
		Features:             features,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Save quota history
	_, err = client.Collection("quota_history").Doc(userID).Set(ctx, quotaHistory)
	if err != nil {
		return fmt.Errorf("failed to create quota history: %v", err)
	}

	return nil
}

// GetQuotaHistoryStatus returns the current status of a user's quota history
func GetQuotaHistoryStatus(ctx context.Context, app *firebase.App, userID string) (*types.QuotaHistory, error) {
	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firestore client: %v", err)
	}
	defer client.Close()

	doc, err := client.Collection("quota_history").Doc(userID).Get(ctx)
	if err != nil || !doc.Exists() {
		return nil, fmt.Errorf("no quota history found for user %s", userID)
	}

	var quotaHistory types.QuotaHistory
	if err := doc.DataTo(&quotaHistory); err != nil {
		return nil, fmt.Errorf("failed to parse quota history: %v", err)
	}

	return &quotaHistory, nil
}

// EnsureUserQuotaHistory ensures a user has quota history, creating it if it doesn't exist
// This is a convenience function that combines checking and seeding
func EnsureUserQuotaHistory(ctx context.Context, app *firebase.App, userID string) error {
	client, err := app.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Firestore client: %v", err)
	}
	defer client.Close()

	// Check if quota history exists
	doc, err := client.Collection("quota_history").Doc(userID).Get(ctx)
	if err == nil && doc.Exists() {
		// Quota history exists, check if it's valid
		var quotaHistory types.QuotaHistory
		if err := doc.DataTo(&quotaHistory); err != nil {
			return fmt.Errorf("failed to parse quota history: %v", err)
		}

		// Check if user's tier has changed
		query := client.Collection("users").Where("userId", "==", userID).Limit(1)
		docs, err := query.Documents(ctx).GetAll()
		if err != nil {
			return fmt.Errorf("failed to get user document: %v", err)
		}
		if len(docs) == 0 {
			return fmt.Errorf("user not found: %s", userID)
		}
		userDoc := docs[0]

		currentTierID := "starter"
		if tid, ok := userDoc.Data()["tierId"].(string); ok && tid != "" {
			currentTierID = tid
		}

		// If tier has changed, update quota history
		if quotaHistory.TierId != currentTierID {
			fmt.Printf("User %s tier changed from %s to %s, updating quota history...\n", userID, quotaHistory.TierId, currentTierID)
			return updateQuotaHistoryForTierChange(ctx, client, userID, currentTierID)
		}

		return nil // Quota history exists and is up to date
	}

	// Create new quota history
	fmt.Printf("Creating new quota history for user %s...\n", userID)
	return createNewQuotaHistory(ctx, client, userID)
}

// BulkSeedQuotaHistory seeds quota history for multiple users
// This is useful for admin operations or bulk migrations
func BulkSeedQuotaHistory(ctx context.Context, app *firebase.App, userIDs []string) (int, int, error) {
	successCount := 0
	failureCount := 0

	for _, userID := range userIDs {
		if err := EnsureUserQuotaHistory(ctx, app, userID); err != nil {
			fmt.Printf("Failed to seed quota history for user %s: %v\n", userID, err)
			failureCount++
		} else {
			successCount++
		}
	}

	return successCount, failureCount, nil
}

// createNewQuotaHistory creates a new quota history entry for a user
func createNewQuotaHistory(ctx context.Context, client *firestore.Client, userID string) error {
	// Validate user exists and is valid
	if err := validateUserExists(ctx, client, userID); err != nil {
		return fmt.Errorf("user validation failed: %v", err)
	}

	// Get user's tier
	userDoc, err := client.Collection("users").Doc(userID).Get(ctx)
	if err != nil {
		logQuotaAction(ctx, userID, "TIER_FETCH_FAILED", fmt.Sprintf("Error: %v", err))
		return fmt.Errorf("failed to get user document: %v", err)
	}

	tierID := "starter" // default tier
	if tid, ok := userDoc.Data()["tierId"].(string); ok && tid != "" {
		tierID = tid
	}

	// Validate tier access
	if err := validateUserTierAccess(ctx, client, userID, tierID); err != nil {
		return fmt.Errorf("tier validation failed: %v", err)
	}

	// Get tier features
	tierDoc, err := client.Collection("tiers").Doc(tierID).Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tier document: %v", err)
	}

	var tier types.Tier
	if err := tierDoc.DataTo(&tier); err != nil {
		return fmt.Errorf("failed to parse tier document: %v", err)
	}

	// Create quota history with all tier features
	features := make([]types.QuotaFeature, len(tier.Features))
	for i, feature := range tier.Features {
		features[i] = types.QuotaFeature{
			Feature:    feature,
			UsageCount: 0,
			LastUsed:   nil,
		}
	}

	quotaHistory := types.QuotaHistory{
		UserId:               userID,
		TierId:               tierID,
		LastSubscriptionDate: time.Now(),
		Features:             features,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Save quota history
	_, err = client.Collection("quota_history").Doc(userID).Set(ctx, quotaHistory)
	if err != nil {
		return fmt.Errorf("failed to create quota history: %v", err)
	}

	return nil
}

// updateQuotaHistoryForTierChange updates quota history when user's tier changes
func updateQuotaHistoryForTierChange(ctx context.Context, client *firestore.Client, userID string, newTierID string) error {
	// Log tier change attempt
	logQuotaAction(ctx, userID, "TIER_CHANGE_ATTEMPT", fmt.Sprintf("Changing from current tier to: %s", newTierID))

	// Get new tier features
	tierDoc, err := client.Collection("tiers").Doc(newTierID).Get(ctx)
	if err != nil {
		logQuotaAction(ctx, userID, "TIER_CHANGE_FETCH_FAILED", fmt.Sprintf("New tier: %s, Error: %v", newTierID, err))
		return fmt.Errorf("failed to get tier document: %v", err)
	}

	var tier types.Tier
	if err := tierDoc.DataTo(&tier); err != nil {
		logQuotaAction(ctx, userID, "TIER_CHANGE_PARSE_FAILED", fmt.Sprintf("New tier: %s, Error: %v", newTierID, err))
		return fmt.Errorf("failed to parse tier document: %v", err)
	}

	// Create new features array with reset usage counts
	features := make([]types.QuotaFeature, len(tier.Features))
	for i, feature := range tier.Features {
		features[i] = types.QuotaFeature{
			Feature:    feature,
			UsageCount: 0,
			LastUsed:   nil,
		}
	}

	// Update quota history
	_, err = client.Collection("quota_history").Doc(userID).Set(ctx, map[string]interface{}{
		"tierId":               newTierID,
		"features":             features,
		"lastSubscriptionDate": time.Now(),
		"updatedAt":            time.Now(),
	}, firestore.MergeAll)

	if err != nil {
		logQuotaAction(ctx, userID, "TIER_CHANGE_UPDATE_FAILED", fmt.Sprintf("New tier: %s, Error: %v", newTierID, err))
		return fmt.Errorf("failed to update quota history for tier change: %v", err)
	}

	// Log successful tier change
	logQuotaAction(ctx, userID, "TIER_CHANGE_SUCCESS", fmt.Sprintf("Successfully changed to tier: %s, Features: %d", newTierID, len(features)))
	return nil
}

// validateUserTierAccess validates that a user can access a specific tier
func validateUserTierAccess(ctx context.Context, client *firestore.Client, userID, tierID string) error {
	// Check if tier exists
	tierDoc, err := client.Collection("tiers").Doc(tierID).Get(ctx)
	if err != nil || !tierDoc.Exists() {
		logQuotaAction(ctx, userID, "TIER_VALIDATION_FAILED", fmt.Sprintf("Tier %s not found", tierID))
		return fmt.Errorf("invalid tier: %s", tierID)
	}

	// Basic tier validation - you can add more business logic here
	if tierID == "" {
		logQuotaAction(ctx, userID, "TIER_VALIDATION_FAILED", "Empty tier ID")
		return fmt.Errorf("tier ID cannot be empty")
	}

	// Log successful tier validation
	logQuotaAction(ctx, userID, "TIER_VALIDATION_SUCCESS", fmt.Sprintf("Tier %s validated", tierID))
	return nil
}

// validateUserExists checks if user exists and is valid
func validateUserExists(ctx context.Context, client *firestore.Client, userID string) error {
	if userID == "" || len(userID) > 100 {
		logQuotaAction(ctx, userID, "USER_VALIDATION_FAILED", "Invalid user ID format")
		return fmt.Errorf("invalid user ID format")
	}

	query := client.Collection("users").Where("userId", "==", userID).Limit(1)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		logQuotaAction(ctx, userID, "USER_VALIDATION_FAILED", "Error fetching user")
		return fmt.Errorf("error fetching user: %v", err)
	}
	if len(docs) == 0 {
		logQuotaAction(ctx, userID, "USER_VALIDATION_FAILED", "User not found")
		return fmt.Errorf("user not found: %s", userID)
	}
	userDoc := docs[0]

	// Check if user is active (you can add more validation here)
	userData := userDoc.Data()
	if status, ok := userData["status"].(string); ok && status == "suspended" {
		logQuotaAction(ctx, userID, "USER_VALIDATION_FAILED", "User account suspended")
		return fmt.Errorf("user account suspended")
	}

	logQuotaAction(ctx, userID, "USER_VALIDATION_SUCCESS", "User validated successfully")
	return nil
}

// checkUserExistsInUsers checks if a user exists in the users collection
func checkUserExistsInUsers(ctx context.Context, client *firestore.Client, userID string) (bool, error) {
	// Log the check attempt
	logQuotaAction(ctx, userID, "USER_EXISTENCE_CHECK_ATTEMPT", "Checking if user exists in users collection")

	query := client.Collection("users").Where("userId", "==", userID).Limit(1)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		logQuotaAction(ctx, userID, "USER_EXISTENCE_CHECK_ERROR", fmt.Sprintf("Error checking user existence: %v", err))
		return false, err // Error occurred
	}

	if len(docs) == 0 {
		logQuotaAction(ctx, userID, "USER_NOT_FOUND_IN_USERS", "User does not exist in users collection")
		return false, nil // User doesn't exist
	}

	logQuotaAction(ctx, userID, "USER_EXISTENCE_CHECK_SUCCESS", "User found in users collection")
	return true, nil // User exists
}

// debugUserExistence provides detailed debugging information about user existence
func debugUserExistence(ctx context.Context, client *firestore.Client, userID string) {
	logQuotaAction(ctx, userID, "DEBUG_USER_EXISTENCE", "Starting debug check")

	// Log the user ID being checked
	logQuotaAction(ctx, userID, "DEBUG_USER_ID", fmt.Sprintf("Checking user ID: '%s' (length: %d)", userID, len(userID)))

	// Try to get user document
	query := client.Collection("users").Where("userId", "==", userID).Limit(1)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		logQuotaAction(ctx, userID, "DEBUG_USER_ERROR", fmt.Sprintf("Error: %v, Type: %T", err, err))

		// Check if it's a permission issue
		if strings.Contains(err.Error(), "Permission denied") || strings.Contains(err.Error(), "permission") {
			logQuotaAction(ctx, userID, "DEBUG_PERMISSION_ERROR", "Permission denied accessing users collection")
		}
		return
	}

	if len(docs) > 0 {
		doc := docs[0]
		data := doc.Data()
		logQuotaAction(ctx, userID, "DEBUG_USER_DATA", fmt.Sprintf("User exists, Data keys: %v", getMapKeys(data)))

		// Log specific fields that might be relevant
		if userId, ok := data["userId"].(string); ok {
			logQuotaAction(ctx, userID, "DEBUG_USER_FIELD", fmt.Sprintf("userId field: '%s'", userId))
		}
		if tierId, ok := data["tierId"].(string); ok {
			logQuotaAction(ctx, userID, "DEBUG_USER_FIELD", fmt.Sprintf("tierId field: '%s'", tierId))
		}
	} else {
		logQuotaAction(ctx, userID, "DEBUG_USER_EXISTS", "User not found")
	}
}

// debugFirestoreStructure helps debug Firestore collection and document structure
func debugFirestoreStructure(ctx context.Context, client *firestore.Client, userID string) {
	logQuotaAction(ctx, userID, "DEBUG_FIRESTORE_STRUCTURE", "Starting Firestore structure debug")

	// List all collections to see what exists
	collections := client.Collections(ctx)
	for {
		collection, err := collections.Next()
		if err != nil {
			break
		}
		logQuotaAction(ctx, userID, "DEBUG_COLLECTION_FOUND", fmt.Sprintf("Collection: %s", collection.ID))
	}

	// Try to get the user document with more detailed error handling
	query := client.Collection("users").Where("userId", "==", userID).Limit(1)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		logQuotaAction(ctx, userID, "DEBUG_DOCUMENT_ERROR", fmt.Sprintf("Document error: %v", err))

		// Check for specific error types
		if strings.Contains(err.Error(), "NotFound") {
			logQuotaAction(ctx, userID, "DEBUG_ERROR_TYPE", "Error type: NotFound")
		} else if strings.Contains(err.Error(), "Permission denied") {
			logQuotaAction(ctx, userID, "DEBUG_ERROR_TYPE", "Error type: Permission denied")
		} else if strings.Contains(err.Error(), "Invalid argument") {
			logQuotaAction(ctx, userID, "DEBUG_ERROR_TYPE", "Error type: Invalid argument")
		} else {
			logQuotaAction(ctx, userID, "DEBUG_ERROR_TYPE", fmt.Sprintf("Error type: Unknown - %s", err.Error()))
		}
		return
	}

	if len(docs) > 0 {
		doc := docs[0]
		data := doc.Data()
		logQuotaAction(ctx, userID, "DEBUG_DOCUMENT_EXISTS", fmt.Sprintf("Document exists with %d fields", len(data)))
		for key, value := range data {
			logQuotaAction(ctx, userID, "DEBUG_DOCUMENT_FIELD", fmt.Sprintf("Field: %s = %v (type: %T)", key, value, value))
		}
	} else {
		logQuotaAction(ctx, userID, "DEBUG_DOCUMENT_EMPTY", "User not found")
	}
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// logQuotaAction logs all quota-related actions for security monitoring
func logQuotaAction(ctx context.Context, userID, action, details string) {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	// Enhanced logging with more context
	log.Printf("QUOTA_SECURITY: [%s] User=%s | Action=%s | Details=%s | IP=%s",
		timestamp, userID, action, details, getClientIP(ctx))
}

// getClientIP extracts client IP from context (if available)
func getClientIP(ctx context.Context) string {
	return "unknown"
}
