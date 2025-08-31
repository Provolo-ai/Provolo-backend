package main

import (
	"context"
	"fmt"
	"os"
	"provolo-api/internal/env"
	"provolo-api/internal/types"
	"provolo-api/internal/utils"
	"time"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	_ "github.com/joho/godotenv/autoload"
)

func migratePromptQuota() {
	fmt.Println("Starting prompt quota migration...")

	// Load env config
	firebaseEncodedConfig := env.GetEnvString("FIREBASE_ENCODED_CONFIG", "")
	secretKey := env.GetEnvString("FIREBASE_SECRET_KEY", "")
	if firebaseEncodedConfig == "" || secretKey == "" {
		fmt.Println("Env config not found")
		os.Exit(1)
	}

	// Decode firebase config
	firebaseConfig, err := utils.DecodeFirebaseConfig(firebaseEncodedConfig, secretKey)
	if err != nil {
		fmt.Printf("Failed to decode firebase config: %v\n", err)
		os.Exit(1)
	}

	// Initialize Firebase
	ctx := context.Background()
	sa := option.WithCredentialsJSON(firebaseConfig)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		fmt.Printf("Failed to create firebase app: %v\n", err)
		os.Exit(1)
	}

	// Initialize Firestore client
	client, err := app.Firestore(ctx)
	if err != nil {
		fmt.Printf("Failed to create firestore client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	iter := client.Collection("user_prompt_limits").Documents(ctx)

	migrated := 0
	skipped := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Failed to read user_prompt_limits: %v\n", err)
			skipped++
			continue
		}

		var upl struct {
			UserId       string    `firestore:"userId"`
			PromptCount  int       `firestore:"promptCount"`
			LastPromptAt time.Time `firestore:"lastPromptAt"`
		}
		if err := doc.DataTo(&upl); err != nil {
			fmt.Printf("Failed to map user_prompt_limits: %v\n", err)
			skipped++
			continue
		}

		// Default tier
		tierId := "starter"

		userDoc, err := client.Collection("users").Doc(upl.UserId).Get(ctx)
		if err == nil {
			if tid, ok := userDoc.Data()["tierId"].(string); ok && tid != "" {
				tierId = tid
			}
		}

		// Build tier features
		features := []types.QuotaFeature{}
		tierDoc, err := client.Collection("tiers").Doc(tierId).Get(ctx)
		if err == nil {
			if feats, ok := tierDoc.Data()["features"].([]interface{}); ok {
				for _, f := range feats {
					if fmap, ok := f.(map[string]interface{}); ok {
						feature := types.Feature{
							Name:              fmap["name"].(string),
							Description:       fmap["description"].(string),
							Slug:              types.FeatureSlug(fmap["slug"].(string)),
							Limited:           fmap["limited"].(bool),
							MaxQuota:          int(fmap["maxQuota"].(int64)),
							RecurringInterval: types.RecurringInterval(fmap["recurringInterval"].(string)),
						}

						features = append(features, types.QuotaFeature{
							Feature:    feature,
							UsageCount: 0,
							LastUsed:   nil,
						})
					}
				}
			}
		}

		// Apply old prompt usage to correct feature
		for i, f := range features {
			if f.Slug == types.FeatureProfileOptimizer {
				features[i].UsageCount = upl.PromptCount
				features[i].LastUsed = &upl.LastPromptAt
			}
		}

		quotaDoc := client.Collection("quota_history").Doc(upl.UserId)

		// If already exists keep CreatedAt
		snapshot, err := quotaDoc.Get(ctx)
		createdAt := time.Now()
		if err == nil && snapshot.Exists() {
			var existing types.QuotaHistory
			if err := snapshot.DataTo(&existing); err == nil {
				createdAt = existing.CreatedAt
			}
		}

		quotaHistory := types.QuotaHistory{
			UserId:               upl.UserId,
			TierId:               tierId,
			LastSubscriptionDate: time.Now(),
			Features:             features,
			CreatedAt:            createdAt,
			UpdatedAt:            time.Now(),
		}

		// upsert quota history
		_, err = quotaDoc.Set(ctx, quotaHistory)
		if err != nil {
			fmt.Printf("Failed to upsert quota_history for user %s: %v\n", upl.UserId, err)
			skipped++
			continue
		}

		migrated++
		fmt.Printf("Upserted quota_history for user %s\n", upl.UserId)
	}

	fmt.Printf("Migration completed ✔️ Migrated: %d, Skipped: %d\n", migrated, skipped)
	os.Exit(0)
}
