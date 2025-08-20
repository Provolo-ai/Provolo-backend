package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"provolo-api/internal/env"
	"provolo-api/internal/types"
	"provolo-api/internal/utils"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

type PromptReq struct {
	FullName          string `json:"full_name" binding:"required" example:"John Doe" validate:"max=100" description:"Freelancer's full name (max 100 characters)"`
	ProfessionalTitle string `json:"professional_title" binding:"required" example:"Full Stack Developer" validate:"max=200" description:"Professional title or role (max 200 characters)"`
	Profile           string `json:"profile" binding:"required" example:"Experienced developer with 5+ years in web development..." validate:"max=1000" description:"Profile description or bio (max 1000 characters)"`
}

// ValidateInput checks input length and content
func validateInput(fieldName, input string, maxLength int) error {
	if len(strings.TrimSpace(input)) == 0 {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	if !utf8.ValidString(input) {
		return fmt.Errorf("%s exceeds maximum length of %d characters", fieldName, maxLength)
	}

	// Check for suspicious patterns
	suspiciousPatterns := regexp.MustCompile(`(?i)(javascript:|data:|vbscript:|file:|ftp:|http://|https://)`)
	if suspiciousPatterns.MatchString(input) {
		return fmt.Errorf("%s contains suspicious content", fieldName)
	}

	return nil
}

// @Summary Optimize freelancer profile using AI
// @Description Analyzes and optimizes a freelancer's profile content using AI to improve client attraction and profile effectiveness
// @Tags Profile Optimization
// @Accept json
// @Produce json
// @Param request body PromptReq true "Profile optimization request containing freelancer details"
// @Success 200 {object} types.APIResponse{data=types.OptimizerResponse} "Profile optimization completed successfully"
// @Failure 400 {object} types.APIResponse "Bad Request - Invalid input validation"
// @Failure 401 {object} types.APIResponse "Unauthorized - Unauthorized"
// @Failure 429 {object} types.APIResponse "Too Many Requests - Daily limit exceeded"
// @Failure 500 {object} types.APIResponse "Internal Server Error - AI service or client creation failed"
// @Router /api/v1/optimize-profile [post]
// Profile Optimizer
func ProfileOptimizer(c *gin.Context) {
	// Get user ID from auth context (set by AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			"Unauthorized",
			"User not authenticated",
		))
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"Internal Error",
			"Invalid user ID format",
		))
		return
	}

	// Check rate limit before processing
	ctx := context.Background()
	firebaseApp, err := utils.GetFirebaseApp(ctx)
	if err != nil {
		log.Printf("Error getting Firebase app: %v", err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"Internal Server Error",
			"Failed to initialize Firebase",
		))
		return
	}

	// Get current limit from environment
	dailyLimit := env.GetEnvInt("MAX_PROMPT_LIMIT", 2)
	limitResult, err := utils.CheckUserPromptLimit(ctx, firebaseApp, userIDStr, dailyLimit)
	if err != nil {
		log.Printf("Error checking rate limit: %v", err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"Internal Server Error",
			"Failed to check rate limit",
		))
		return
	}

	if !limitResult.Allowed {
		c.JSON(http.StatusTooManyRequests, types.NewErrorResponse(
			"Rate Limit Exceeded",
			fmt.Sprintf("Daily limit of %d prompts exceeded. Current count: %d. Try again tomorrow.", limitResult.Limit, limitResult.Count),
		))
		return
	}

	var req PromptReq
	if bindErr := c.ShouldBind(&req); bindErr != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"Invalid Request",
			"Invalid request body: "+bindErr.Error(),
		))
		return
	}

	// Validate input lengths and content
	if validateErr := validateInput("Full Name", req.FullName, 100); validateErr != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"Validation Error",
			validateErr.Error(),
		))
		return
	}

	if validateErr := validateInput("Professional Title", req.ProfessionalTitle, 200); validateErr != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"Validation Error",
			validateErr.Error(),
		))
		return
	}

	if validateErr := validateInput("Profile", req.Profile, 5000); validateErr != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"Validation Error",
			validateErr.Error(),
		))
		return
	}

	// Sanitize all inputs
	sanitizedFullName := utils.SanitizeInput(req.FullName)
	sanitizedTitle := utils.SanitizeInput(req.ProfessionalTitle)
	sanitizedProfile := utils.SanitizeInput(req.Profile)

	inputContent := strings.TrimSpace(
		"Freelancer Name: " + sanitizedFullName +
			"\nTitle: " + sanitizedTitle +
			"\n\n Profile Description:\n" + sanitizedProfile,
	)

	content := utils.OptimizerPrompt(inputContent)

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: env.GetEnvString("GEMINI_API_KEY", ""),
	})
	if err != nil {
		log.Printf("Error creating Gemini client: %v", err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"Internal Server Error",
			"Failed to create Gemini client: "+err.Error(),
		))
		return
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(content),
		nil,
	)
	if err != nil {
		log.Printf("Error generating content: %v", err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"AI Service Error",
			"Failed to generate content",
		))
		return
	}

	// After successful Gemini response parsing, but BEFORE sending JSON response:
	log.Printf("Gemini response: %s", result.Text())
	parsedResponse, err := utils.ParseGeminiJSONBlock(result.Text())
	if err != nil {
		log.Printf("Error parsing Gemini response: %v", err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"Processing Error",
			"Failed to process AI response: "+err.Error(),
		))
		return
	}

	// Update the count after successful processing but before sending response
	if err := utils.UpdateUserPromptLimit(ctx, firebaseApp, userIDStr); err != nil {
		// Log the error but don't fail the request since we already got the response
		log.Printf("Warning: Failed to update prompt count for user %s: %v", userIDStr, err)
	}

	log.Printf("Generated content: %v", result)
	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Optimization Successfully",
		"Profile optimized successfully",
		parsedResponse,
	))
}
