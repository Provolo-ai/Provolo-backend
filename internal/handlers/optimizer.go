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
// @Failure 500 {object} types.APIResponse "Internal Server Error - AI service or client creation failed"
// @Router /api/v1/optimize-profile [post]
// Profile Optimizer
func ProfileOptimizer(c *gin.Context) {
	var req PromptReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadGateway, types.NewErrorResponse(
			"Invalid Request",
			"Invalid request body: "+err.Error(),
		))
		return
	}

	// Validate input lengths and content
	if err := validateInput("Full Name", req.FullName, 100); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"Validation Error",
			err.Error(),
		))
		return
	}

	if err := validateInput("Professional Title", req.ProfessionalTitle, 200); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"Validation Error",
			err.Error(),
		))
		return
	}

	if err := validateInput("Profile", req.Profile, 5000); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"Validation Error",
			err.Error(),
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

	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: env.GetEnvString("GEMINI_API_KEY", ""),
	})
	if err != nil {
		log.Printf("Error creating Gemini client: %v", err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"Internal Server Error",
			"Failed to create Gemini client: "+err.Error(),
		))
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

	parsedResponse, err := utils.ParseGeminiJSONBlock(result.Text())
	if err != nil {
		log.Printf("Error parsing Gemini response: %v", err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"Processing Error",
			"Failed to process AI response: "+err.Error(),
		))
		return
	}

	log.Printf("Generated content: %v", result)
	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Optimization Successfully",
		"Profile optimized successfully",
		parsedResponse,
	))
}
