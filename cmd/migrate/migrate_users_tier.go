package main

import (
	"context"
	"fmt"
	"os"
	"provolo-api/internal/env"
	"provolo-api/internal/types"
	"provolo-api/internal/utils"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func migrateUsersTier() {
	fmt.Println("Starting user tiers migration...")

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

	// starter tier ref
	starterTierRef := client.Collection("tiers").Doc(types.DefaultTierID)

	// Query users with empty string tierId
	fmt.Println("Querying users with tierId == \"\" ...")
	userIterEmpty := client.Collection("users").Where("tierId", "==", "").Documents(ctx)
	processUsers(ctx, userIterEmpty, starterTierRef)

	// Query all users and filter those with missing tierId
	fmt.Println("Querying users with missing tierId ...")
	userIterAll := client.Collection("users").Documents(ctx)
	for {
		doc, err := userIterAll.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Failed to read user document: %v\n", err)
			break
		}
		data := doc.Data()
		if _, ok := data["tierId"]; !ok {
			// tierId is missing, update this user
			processSingleUser(ctx, doc.Ref, starterTierRef)
		}
	}

	fmt.Println("Migration completed ✔️")
}

// processUsers iterates through users and updates their tierId
func processUsers(ctx context.Context, iter *firestore.DocumentIterator, starterTierRef *firestore.DocumentRef) {
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Failed to read user document: %v\n", err)
			break
		}

		fmt.Printf("Updating user %s -> set tierId to %s\n", doc.Ref.ID, starterTierRef.ID)

		_, err = doc.Ref.Update(ctx, []firestore.Update{
			{Path: "tierId", Value: starterTierRef.ID},
			{Path: "updatedAt", Value: time.Now()},
		})
		if err != nil {
			fmt.Printf("Failed to update user %s: %v\n", doc.Ref.ID, err)
		} else {
			fmt.Printf("User %s updated successfully\n", doc.Ref.ID)
		}
	}
}

// processSingleUser updates the tierId of a single user document
func processSingleUser(ctx context.Context, userRef *firestore.DocumentRef, starterTierRef *firestore.DocumentRef) {
	_, err := userRef.Update(ctx, []firestore.Update{
		{Path: "tierId", Value: starterTierRef.ID},
		{Path: "updatedAt", Value: time.Now()},
	})
	if err != nil {
		fmt.Printf("Failed to update user %s: %v\n", userRef.ID, err)
	} else {
		fmt.Printf("User %s updated successfully\n", userRef.ID)
	}
}
