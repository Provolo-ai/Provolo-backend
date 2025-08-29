package types

import "time"

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
	PolarRefId        string                `firestore:"polarRefId" json:"polarRefId"`
	Price             int                   `firestore:"price" json:"price"`
	Description       string                `firestore:"description" json:"description"`
	RecurringInterval PlanRecurringInterval `firestore:"recurringInterval" json:"recurringInterval"`
	Features          []Feature             `firestore:"features" json:"features"`
	CreatedAt         time.Time             `firestore:"createdAt" json:"createdAt"`
	UpdatedAt         time.Time             `firestore:"updatedAt" json:"updatedAt"`
}
