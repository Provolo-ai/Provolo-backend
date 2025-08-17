package utils

import (
	"encoding/json"
	"errors"
	"provolo-api/internal/types"
	"regexp"
	"strings"
)

func cleanGeminiResponse(text string) string {
	// Remove leading/trailing whitespace
	cleaned := strings.TrimSpace(text)

	// Remove code block markers (case insensitive)
	jsonBlockRegex := regexp.MustCompile(`(?i)^\x60{3}json\s*`)
	cleaned = jsonBlockRegex.ReplaceAllString(cleaned, "")

	// Remove generic code block markers
	codeBlockRegex := regexp.MustCompile(`^\x60{3}\s*`)
	cleaned = codeBlockRegex.ReplaceAllString(cleaned, "")

	// Remove trailing code block markers
	trailingBlockRegex := regexp.MustCompile(`\x60{3}\s*$`)
	cleaned = trailingBlockRegex.ReplaceAllString(cleaned, "")

	// Final trim
	return strings.TrimSpace(cleaned)
}

// validatePortfolioResponse validates that all required fields are present and non-empty
func validatePortfolioResponse(response *types.OptimizerResponse) error {
	if response.WeaknessesAndOptimization == "" {
		return errors.New("weaknesses analysis is required")
	}
	if response.OptimizedProfileOverview == "" {
		return errors.New("optimized profile overview is required")
	}
	if response.SuggestedProjectTitles == "" {
		return errors.New("project title suggestions are required")
	}
	if response.RecommendedVisuals == "" {
		return errors.New("visual recommendations are required")
	}
	if response.BeforeAfterComparison == "" {
		return errors.New("before/after comparison is required")
	}
	return nil
}

// ParseGeminiJSONBlock cleans and parses JSON from Gemini response
// Removes code block markers, trims whitespace, and attempts JSON parsing
func ParseGeminiJSONBlock(text string) (*types.OptimizerResponse, error) {
	if text == "" {
		return nil, errors.New("empty response text")
	}

	// Clean the response text
	cleaned := cleanGeminiResponse(text)

	// Parse JSON
	var response types.OptimizerResponse
	err := json.Unmarshal([]byte(cleaned), &response)
	if err != nil {
		return nil, errors.New("failed to parse JSON: " + err.Error())
	}

	// Validate required fields
	if err := validatePortfolioResponse(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
