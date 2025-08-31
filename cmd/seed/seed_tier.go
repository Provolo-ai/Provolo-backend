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
	"google.golang.org/api/option"

	_ "github.com/joho/godotenv/autoload"
)

// PREDEFINED TIERS
var tiers = []types.Tier{
	{
		Name:              "Starter (Freemium)",
		Slug:              "starter",
		Description:       "Perfect for new freelancers and those exploring the platform.",
		RecurringInterval: types.PlanMonthly,
		Price:             0,
		PolarRefId:        "d1173db4-8051-47a6-a3de-ba6296b2fb17",
		Features: []types.Feature{
			{
				Name:              "Profile Optimizer",
				Description:       "Full access to the Profile Optimizer feature.",
				Slug:              types.FeatureProfileOptimizer,
				Limited:           true,
				RecurringInterval: types.Daily,
				MaxQuota:          2,
			},
			{
				Name:              "AI Proposal Credit",
				Description:       "1 AI Proposal credit per month.",
				Slug:              types.FeatureAIProposalsUnlimited,
				Limited:           true,
				MaxQuota:          1,
				RecurringInterval: types.Monthly,
			},
			{
				Name:        "Standard Support",
				Description: "Support via Twitter.",
				Slug:        types.FeatureStandardSupport,
				Limited:     false,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		Name:              "Pro",
		Slug:              "pro",
		Description:       "For freelancers actively applying for jobs and serious about getting clients.",
		RecurringInterval: types.PlanMonthly,
		Price:             1999, // $19.99
		PolarRefId:        "503fe6a4-b148-41bb-b779-60334594794e",
		Features: []types.Feature{
			{
				Name:              "Profile Optimizer",
				Description:       "Full access to the Profile Optimizer feature.",
				Slug:              types.FeatureProfileOptimizer,
				Limited:           true,
				MaxQuota:          5,
				RecurringInterval: types.Daily,
			},
			{
				Name:        "AI Proposals",
				Description: "Unlimited AI Proposals per month.",
				Slug:        types.FeatureAIProposalsUnlimited,
				Limited:     false,
			},
			{
				Name:        "LinkedIn Profile Optimizer",
				Description: "Access to the upcoming LinkedIn Profile Optimizer feature.",
				Slug:        types.FeatureLinkedInOptimizer,
				Limited:     false,
			},
			{
				Name:        "Priority Support",
				Description: "Get faster support response times.",
				Slug:        types.FeaturePrioritySupport,
				Limited:     false,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		Name:              "Guru",
		Slug:              "guru",
		Description:       "For established freelancers scaling their business across multiple platforms.",
		RecurringInterval: types.PlanMonthly,
		Price:             4999, // $49.99
		PolarRefId:        "070d64d8-764f-4fbf-aacd-ea895d90ea48",
		Features: []types.Feature{
			{
				Name:        "Profile Optimizer",
				Description: "Full access to the Profile Optimizer feature.",
				Slug:        types.FeatureProfileOptimizer,
				Limited:     false,
			},
			{
				Name:        "AI Proposals",
				Description: "Unlimited AI Proposals per month.",
				Slug:        types.FeatureAIProposalsUnlimited,
				Limited:     false,
			},
			{
				Name:        "LinkedIn Profile Optimizer",
				Description: "Access to the upcoming LinkedIn Profile Optimizer feature.",
				Slug:        types.FeatureLinkedInOptimizer,
				Limited:     false,
			},
			{
				Name:        "Priority Support",
				Description: "Get faster support response times.",
				Slug:        types.FeaturePrioritySupport,
				Limited:     false,
			},
			{
				Name:        "Resume Generator",
				Description: "Generate professional resumes automatically.",
				Slug:        types.FeatureResumeGenerator,
				Limited:     false,
			},
			{
				Name:        "Advanced AI Insights",
				Description: "A/B testing proposals and tracking proposal performance.",
				Slug:        types.FeatureAdvancedAIInsights,
				Limited:     false,
			},
			{
				Name:        "Early Access",
				Description: "Be the first to access upcoming features.",
				Slug:        types.FeatureEarlyAccess,
				Limited:     false,
			},
			{
				Name:        "Direct Support Chat",
				Description: "Get direct access to support chat.",
				Slug:        types.FeatureDirectSupportChat,
				Limited:     false,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
}

func validateFeatures(features []types.Feature) error {
	for _, f := range features {
		if f.Limited {
			if f.MaxQuota <= 0 {
				return fmt.Errorf("feature %s is limited but MaxQuota is not set", f.Slug)
			}
			if f.RecurringInterval == "" {
				return fmt.Errorf("feature %s is limited but RecurringInterval is not set", f.Slug)
			}
		} else {
			// Enforce normalization for non-limited features
			if f.MaxQuota != 0 || f.RecurringInterval != "" {
				return fmt.Errorf("feature %s is not limited but MaxQuota/RecurringInterval set", f.Slug)
			}
		}
	}
	return nil
}

func main() {
	fmt.Println("started seeding")

	firebaseEncodedConfig := env.GetEnvString("FIREBASE_ENCODED_CONFIG", "")
	secretKey := env.GetEnvString("FIREBASE_SECRET_KEY", "")

	if firebaseEncodedConfig == "" || secretKey == "" {
		fmt.Println("Environment variables not found!")
		os.Exit(1)
	}

	// Decode service account config JSON
	configData, decodedErr := utils.DecodeFirebaseConfig(firebaseEncodedConfig, secretKey)
	if decodedErr != nil {
		fmt.Printf("Error decoding firebase config: %v\n", decodedErr)
		os.Exit(1)
	}

	// Initialize Firebase App
	ctx := context.Background()
	opt := option.WithCredentialsJSON(configData)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		fmt.Printf("Error initializing Firebase App: %v\n", err)
		os.Exit(1)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		fmt.Printf("Error creating Firestore client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Seed tiers
	for _, tier := range tiers {
		// Validate features
		if err := validateFeatures(tier.Features); err != nil {
			fmt.Printf("Validation failed for tier %s: %v\n", tier.Slug, err)
			continue
		}

		// Ensure updated timestamps
		tier.CreatedAt = time.Now()
		tier.UpdatedAt = time.Now()

		// Use slug as doc ID for idempotency
		_, err := client.Collection("tiers").Doc(tier.Slug).Set(ctx, tier)
		if err != nil {
			fmt.Printf("Error seeding tier %s: %v\n", tier.Slug, err)
		} else {
			fmt.Printf("Seeded tier: %s\n", tier.Slug)
		}
	}

	fmt.Println("Seeding complete ✅")
}
