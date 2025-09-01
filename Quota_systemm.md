# Quota System Documentation

## Overview

The new quota system uses the `quota_history` table instead of the old `user_prompt_limits` table. This provides better flexibility and allows for different quota limits per feature and tier.

## Key Components

### 1. Quota History Table Structure

The `quota_history` table contains:
- `userId`: User identifier
- `tierId`: User's current tier
- `features`: Array of features with usage tracking
- `lastSubscriptionDate`: When the user last subscribed
- `createdAt` and `updatedAt`: Timestamps

### 2. Feature Quota Structure

Each feature in the quota history includes:
- `name`: Feature name
- `description`: Feature description
- `slug`: Feature identifier (e.g., "profile_optimizer")
- `limited`: Whether the feature has quota limits
- `maxQuota`: Maximum allowed usage per period
- `recurringInterval`: Reset period (daily, weekly, monthly)
- `usageCount`: Current usage count
- `lastUsed`: Last usage timestamp

## Usage Examples

### Basic Quota Check

```go
// Check if user has quota available
result, err := utils.CheckUserQuota(ctx, app, userID, types.FeatureProfileOptimizer)
if err != nil {
    // Handle error
}

if !result.Allowed {
    // Return quota exceeded error
    return fmt.Errorf("quota exceeded: %d/%d", result.Count, result.Limit)
}
```

### Check and Update Quota

```go
// Check quota and update it if allowed
result, err := utils.CheckAndUpdateQuota(ctx, app, userID, types.FeatureProfileOptimizer)
if err != nil {
    // Handle error
}

if !result.Allowed {
    // Return quota exceeded error
    return fmt.Errorf("quota exceeded: %d/%d", result.Count, result.Limit)
}

// Quota was updated successfully
```

### Create Quota History for New User

```go
// Create quota history from user's tier
err := utils.CreateQuotaHistoryFromTier(ctx, app, userID)
if err != nil {
    // Handle error
}
```

## Migration from Old System

The migration script `migrate_prompt_quota.go` automatically:
1. Reads existing `user_prompt_limits` data
2. Creates `quota_history` entries for each user
3. Maps old prompt usage to the new quota system
4. Preserves user tier information

## Benefits of New System

1. **Feature-specific quotas**: Different limits for different features
2. **Tier-based limits**: Quotas automatically adjust based on user's subscription tier
3. **Flexible intervals**: Support for daily, weekly, and monthly reset periods
4. **Better tracking**: More detailed usage analytics
5. **Scalability**: Easier to add new features and modify quotas

## Implementation in Handlers

The profile optimizer handler now uses:

```go
// Check quota before processing
quotaResult, err := utils.CheckUserQuota(ctx, firebaseApp, userIDStr, types.FeatureProfileOptimizer)
if err != nil {
    // Handle error
}

if !quotaResult.Allowed {
    c.JSON(http.StatusTooManyRequests, types.NewErrorResponse(
        "Quota Exceeded",
        fmt.Sprintf("Quota limit exceeded for profile optimizer. Current usage: %d/%d. Try again in the next period.", 
            quotaResult.Count, quotaResult.Limit),
    ))
    return
}

// ... process request ...

// Update quota after successful processing
if err := utils.UpdateUserQuota(ctx, firebaseApp, userIDStr, types.FeatureProfileOptimizer); err != nil {
    log.Printf("Warning: Failed to update quota for user %s: %v", userIDStr, err)
}
```

## Available Features

Current feature slugs:
- `profile_optimizer`: Profile optimization service
- `ai_proposal_credit`: AI proposal generation credits
- `ai_proposals_unlimited`: Unlimited AI proposals
- `linkedin_profile_optimizer`: LinkedIn profile optimization
- `priority_support`: Priority customer support
- `resume_generator`: Resume generation service
- `advanced_ai_insights`: Advanced AI insights
- `early_access_features`: Early access to new features
- `direct_support_chat`: Direct support chat access
- `standard_support`: Standard support access
