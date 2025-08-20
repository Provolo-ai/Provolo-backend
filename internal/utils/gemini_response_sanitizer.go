package utils

import (
	"encoding/json"
	"errors"
	"log"
	"provolo-api/internal/types"
	"strings"
)

// validatePortfolioResponse validates that required fields are present
func validatePortfolioResponse(response *types.OptimizerResponse) error {
	if response.FullAnalysis == "" {
		return errors.New("fullAnalysis is required")
	}
	if response.WeaknessesAndOptimization == "" {
		return errors.New("weaknessesAndOptimization is required")
	}
	if response.OptimizedProfileOverview == "" {
		return errors.New("optimized profile overview is required")
	}
	if response.SuggestedProjectTitles == "" {
		return errors.New("suggestedProjectTitles is required")
	}
	if response.RecommendedVisuals == "" {
		return errors.New("recommendedVisuals is required")
	}
	if response.BeforeAfterComparison == "" {
		return errors.New("beforeAfterComparison is required")
	}
	return nil
}

func cleanGeminiResponse(text string) string {
	// Find the JSON block within code fences
	jsonStart := strings.Index(text, "```json")
	if jsonStart == -1 {
		// Try without the language specifier
		jsonStart = strings.Index(text, "```")
		if jsonStart == -1 {
			return strings.TrimSpace(text)
		}
	}

	// Find the start of actual JSON content
	startPos := strings.Index(text[jsonStart:], "\n")
	if startPos == -1 {
		startPos = strings.Index(text[jsonStart:], "{")
	} else {
		startPos = jsonStart + startPos + 1
		// Find the opening brace
		bracePos := strings.Index(text[startPos:], "{")
		if bracePos != -1 {
			startPos = startPos + bracePos
		}
	}

	// Find the closing code fence
	endPos := strings.Index(text[startPos:], "```")
	if endPos == -1 {
		// If no closing fence, find the last closing brace
		lastBrace := strings.LastIndex(text, "}")
		if lastBrace != -1 {
			return strings.TrimSpace(text[startPos : lastBrace+1])
		}
		return strings.TrimSpace(text[startPos:])
	}

	// Extract JSON content between braces
	jsonContent := text[startPos : startPos+endPos]

	// Find the actual JSON object boundaries
	firstBrace := strings.Index(jsonContent, "{")
	lastBrace := strings.LastIndex(jsonContent, "}")

	if firstBrace != -1 && lastBrace != -1 && lastBrace > firstBrace {
		return strings.TrimSpace(jsonContent[firstBrace : lastBrace+1])
	}

	return strings.TrimSpace(jsonContent)
}

// ParseGeminiJSONBlock parses the Gemini response, extracts JSON for structured fields, and stores full text
func ParseGeminiJSONBlock(text string) (*types.OptimizerResponse, error) {
	if text == "" {
		return nil, errors.New("empty response text")
	}

	// Store the full response text
	fullAnalysis := text

	// Try to extract and parse JSON block
	cleaned := cleanGeminiResponse(text)

	// Debug log the cleaned JSON
	log.Printf("Cleaned JSON for parsing: %s", cleaned[:min(200, len(cleaned))])

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &jsonData); err != nil {
		return nil, errors.New("failed to parse JSON from response: " + err.Error())
	}

	// Extract fields from JSON, with type checking
	getStringField := func(key string) string {
		if val, ok := jsonData[key]; ok {
			if str, ok := val.(string); ok {
				return str
			}
		}
		return ""
	}

	// Create response with both full analysis and parsed JSON fields
	response := &types.OptimizerResponse{
		FullAnalysis:              fullAnalysis, // Complete raw response
		WeaknessesAndOptimization: getStringField("weaknessesAndOptimization"),
		OptimizedProfileOverview:  getStringField("optimizedProfileOverview"),
		SuggestedProjectTitles:    getStringField("suggestedProjectTitles"),
		RecommendedVisuals:        getStringField("recommendedVisuals"),
		BeforeAfterComparison:     getStringField("beforeAfterComparison"),
	}

	// Validate required fields
	if err := validatePortfolioResponse(response); err != nil {
		return nil, err
	}

	return response, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
