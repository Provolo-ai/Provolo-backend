package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"provolo-api/internal/types"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Custom error type for unauthorized content
type UnauthorizedContentError struct {
	Message string
	Code    string
}

func (e *UnauthorizedContentError) Error() string {
	return e.Message
}

// ParseGeminiJSONBlock parses the Gemini response and handles both success and error formats
func ParseGeminiJSONBlock(text string) (*types.OptimizerResponse, error) {
	if text == "" {
		return nil, errors.New("empty response text")
	}

	// Store the full response text for debugging and fallback
	fullAnalysis := text

	// Try to extract and parse JSON block
	cleaned := cleanGeminiResponse(text)

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from response: %v", err)
	}

	// Check if this is an error response
	if errorFlag, exists := jsonData["error"]; exists {
		if isError, ok := errorFlag.(bool); ok && isError {
			message := "Content not authorized for processing"
			code := "UNAUTHORIZED_CONTENT"

			if msg, ok := jsonData["message"].(string); ok && msg != "" {
				message = msg
			}
			if c, ok := jsonData["code"].(string); ok && c != "" {
				code = c
			}

			return nil, &UnauthorizedContentError{
				Message: message,
				Code:    code,
			}
		}
	}

	// Extract fields from JSON for success response
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
		FullAnalysis:              fullAnalysis,
		WeaknessesAndOptimization: getStringField("weaknessesAndOptimization"),
		OptimizedProfileOverview:  getStringField("optimizedProfileOverview"),
		SuggestedProjectTitles:    getStringField("suggestedProjectTitles"),
		RecommendedVisuals:        getStringField("recommendedVisuals"),
		BeforeAfterComparison:     getStringField("beforeAfterComparison"),
	}

	// Validate required fields
	if err := validatePortfolioResponse(response); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}

	return response, nil
}

// cleanGeminiResponse extracts and sanitizes the JSON block from the response
func cleanGeminiResponse(text string) string {
	// Step 1: Trim whitespace first
	text = strings.TrimSpace(text)

	// Step 2: Check if the entire response is already valid JSON
	if json.Valid([]byte(text)) {
		return text
	}

	// Step 3: Try to extract JSON block between ```json and ```
	re := regexp.MustCompile("(?s)```json\n(.*?)\n```")
	matches := re.FindStringSubmatch(text)
	jsonBlock := text
	if len(matches) >= 2 {
		jsonBlock = matches[1]
		jsonBlock = strings.TrimSpace(jsonBlock)

		// Check if extracted block is valid
		if json.Valid([]byte(jsonBlock)) {
			return jsonBlock
		}
	}

	// Step 4: Final check - if it looks like JSON structure but isn't valid, try sanitization
	if !strings.HasPrefix(text, "{") || !strings.HasSuffix(text, "}") {
		return fmt.Sprintf("{\"raw_response\":%q}", text)
	}

	// Step 5: Pre-process the JSON to fix common escape sequence issues
	// Use the original text for sanitization
	jsonBlock = text

	// First, temporarily replace valid escape sequences to protect them
	jsonBlock = strings.ReplaceAll(jsonBlock, "\\\"", "__QUOTE__")
	jsonBlock = strings.ReplaceAll(jsonBlock, "\\n", "__NEWLINE__")
	jsonBlock = strings.ReplaceAll(jsonBlock, "\\r", "__RETURN__")
	jsonBlock = strings.ReplaceAll(jsonBlock, "\\t", "__TAB__")

	// Fix invalid escape sequences (like \' which is not valid in JSON)
	jsonBlock = strings.ReplaceAll(jsonBlock, "\\\\'", "'")

	// Now sanitize string values
	reString := regexp.MustCompile(`"(.*?)"`)
	cleaned := reString.ReplaceAllStringFunc(jsonBlock, func(s string) string {
		// Extract content between quotes
		content := s[1 : len(s)-1]

		// Escape special characters that aren't already escaped
		content = strings.ReplaceAll(content, "\"", "\\\"")
		content = strings.ReplaceAll(content, "\n", "\\n")
		content = strings.ReplaceAll(content, "\r", "\\r")
		content = strings.ReplaceAll(content, "\t", "\\t")

		// Restore protected sequences
		content = strings.ReplaceAll(content, "__QUOTE__", "\\\"")
		content = strings.ReplaceAll(content, "__NEWLINE__", "\\n")
		content = strings.ReplaceAll(content, "__RETURN__", "\\r")
		content = strings.ReplaceAll(content, "__TAB__", "\\t")

		return fmt.Sprintf("\"%s\"", content)
	})

	// Step 6: Validate that the result is valid JSON
	if !json.Valid([]byte(cleaned)) {
		return fmt.Sprintf("{\"raw_response\":%q}", text)
	}

	return cleaned
}

// validatePortfolioResponse checks if required fields are present using validator
func validatePortfolioResponse(resp *types.OptimizerResponse) error {
	validate := validator.New()
	if err := validate.Struct(resp); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}
	return nil
}
