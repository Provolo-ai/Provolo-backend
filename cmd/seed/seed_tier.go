package main

import (
	"context"
	"fmt"
	"os"
	"provolo-api/internal/env"
	"provolo-api/internal/utils"
	"time"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

	_ "github.com/joho/godotenv/autoload"
)

// ENUM DEFINITIONS
type RecurringInterval string
type PlanRecurringInterval string
type FeatureSlug string

// Enums
const (
	// Feature quota recurrence
	Daily   RecurringInterval = "daily"
	Weekly  RecurringInterval = "weekly"
	Monthly RecurringInterval = "monthly"

	// Plan recurrence
	PlanMonthly PlanRecurringInterval = "monthly"

	// Feature slugs
	FeatureProfileOptimizer     FeatureSlug = "profile_optimizer"
	FeatureAIProposalCredit     FeatureSlug = "ai_proposal_credit"
	FeatureAIProposalsUnlimited FeatureSlug = "ai_proposals_unlimited"
	FeatureLinkedInOptimizer    FeatureSlug = "linkedin_profile_optimizer"
	FeaturePrioritySupport      FeatureSlug = "priority_support"
	FeatureResumeGenerator      FeatureSlug = "resume_generator"
	FeatureAdvancedAIInsights   FeatureSlug = "advanced_ai_insights"
	FeatureEarlyAccess          FeatureSlug = "early_access_features"
	FeatureDirectSupportChat    FeatureSlug = "direct_support_chat"
	FeatureStandardSupport      FeatureSlug = "standard_support"
)

// Feature type def
type Feature struct {
	Name              string            `firestore:"name" json:"name"`
	Description       string            `firestore:"description" json:"description"`
	Slug              FeatureSlug       `firestore:"slug" json:"slug"`
	Limited           bool              `firestore:"limited" json:"limited"`
	MaxQuota          int               `firestore:"maxQuota" json:"maxQuota"`
	RecurringInterval RecurringInterval `firestore:"recurringInterval" json:"recurringInterval"`
}

// Tier type def
type Tier struct {
	Name              string                `firestore:"name" json:"name"`
	Slug              string                `firestore:"slug" json:"slug"`
	Price             int                   `firestore:"price" json:"price"` // price in cents
	Description       string                `firestore:"description" json:"description"`
	RecurringInterval PlanRecurringInterval `firestore:"recurringInterval" json:"recurringInterval"`
	Features          []Feature             `firestore:"features" json:"features"`
	CreatedAt         time.Time             `firestore:"createdAt" json:"createdAt"`
	UpdatedAt         time.Time             `firestore:"updatedAt" json:"updatedAt"`
}

// PREDEFINED TIERS
var tiers = []Tier{
	{
		Name:              "Starter (Freemium)",
		Slug:              "starter",
		Description:       "Perfect for new freelancers and those exploring the platform.",
		RecurringInterval: PlanMonthly,
		Price:             0,
		Features: []Feature{
			{
				Name:              "Profile Optimizer",
				Description:       "Full access to the Profile Optimizer feature.",
				Slug:              FeatureProfileOptimizer,
				Limited:           true,
				RecurringInterval: Daily,
				MaxQuota:          2,
			},
			{
				Name:              "AI Proposal Credit",
				Description:       "1 AI Proposal credit per month.",
				Slug:              FeatureAIProposalCredit,
				Limited:           true,
				MaxQuota:          1,
				RecurringInterval: Monthly,
			},
			{
				Name:        "Standard Support",
				Description: "Support via Twitter.",
				Slug:        FeatureStandardSupport,
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
		RecurringInterval: PlanMonthly,
		Price:             1999, // $19.99
		Features: []Feature{
			{
				Name:              "Profile Optimizer",
				Description:       "Full access to the Profile Optimizer feature.",
				Slug:              FeatureProfileOptimizer,
				Limited:           true,
				MaxQuota:          5,
				RecurringInterval: Daily,
			},
			{
				Name:        "AI Proposals",
				Description: "Unlimited AI Proposals per month.",
				Slug:        FeatureAIProposalsUnlimited,
				Limited:     false,
			},
			{
				Name:        "LinkedIn Profile Optimizer",
				Description: "Access to the upcoming LinkedIn Profile Optimizer feature.",
				Slug:        FeatureLinkedInOptimizer,
				Limited:     false,
			},
			{
				Name:        "Priority Support",
				Description: "Get faster support response times.",
				Slug:        FeaturePrioritySupport,
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
		RecurringInterval: PlanMonthly,
		Price:             4999, // $49.99
		Features: []Feature{
			{
				Name:        "Profile Optimizer",
				Description: "Full access to the Profile Optimizer feature.",
				Slug:        FeatureProfileOptimizer,
				Limited:     false,
			},
			{
				Name:        "AI Proposals",
				Description: "Unlimited AI Proposals per month.",
				Slug:        FeatureAIProposalsUnlimited,
				Limited:     false,
			},
			{
				Name:        "LinkedIn Profile Optimizer",
				Description: "Access to the upcoming LinkedIn Profile Optimizer feature.",
				Slug:        FeatureLinkedInOptimizer,
				Limited:     false,
			},
			{
				Name:        "Priority Support",
				Description: "Get faster support response times.",
				Slug:        FeaturePrioritySupport,
				Limited:     false,
			},
			{
				Name:        "Resume Generator",
				Description: "Generate professional resumes automatically.",
				Slug:        FeatureResumeGenerator,
				Limited:     false,
			},
			{
				Name:        "Advanced AI Insights",
				Description: "A/B testing proposals and tracking proposal performance.",
				Slug:        FeatureAdvancedAIInsights,
				Limited:     false,
			},
			{
				Name:        "Early Access",
				Description: "Be the first to access upcoming features.",
				Slug:        FeatureEarlyAccess,
				Limited:     false,
			},
			{
				Name:        "Direct Support Chat",
				Description: "Get direct access to support chat.",
				Slug:        FeatureDirectSupportChat,
				Limited:     false,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
}

func validateFeatures(features []Feature) error {
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
			fmt.Printf("❌ Validation failed for tier %s: %v\n", tier.Slug, err)
			continue // skip this tier instead of writing invalid data
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
