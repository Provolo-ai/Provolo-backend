package utils

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ValidateInput checks input length and content
func ValidatePromptInput(fieldName, input string, maxLength int) error {
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

	// Check for HTML/XML tags that could manipulate response format
	htmlTagPatterns := regexp.MustCompile(`(?i)<\s*/?\s*(script|iframe|object|embed|form|input|textarea|select|button|link|meta|style|base|title|head|body|html|div|span|img|video|audio|source|track|canvas|svg|math|template|slot|shadow)\s*[^>]*>`)
	if htmlTagPatterns.MatchString(input) {
		return fmt.Errorf("%s contains illegal HTML tags", fieldName)
	}

	// Check for JSON manipulation attempts
	jsonManipulationPatterns := regexp.MustCompile(`(?i)(\"\s*[,}]|[,{]\s*\"|\\\"|\\n|\\r|\\t|\\\\|\\u[0-9a-f]{4})`)
	if jsonManipulationPatterns.MatchString(input) {
		return fmt.Errorf("%s contains potential JSON manipulation characters", fieldName)
	}

	// Check for format manipulation instructions
	formatManipulationPatterns := regexp.MustCompile(`(?i)(put.*in.*tag|embed.*into|wrap.*with|format.*as|output.*in|return.*as|generate.*in.*format|inside.*tag|within.*element)`)
	if formatManipulationPatterns.MatchString(input) {
		return fmt.Errorf("%s contains format manipulation instructions", fieldName)
	}

	// Check for system instruction override attempts
	systemOverridePatterns := regexp.MustCompile(`(?i)(ignore.*instruction|forget.*rule|override.*system|change.*format|modify.*response|alter.*output|bypass.*validation|disable.*check)`)
	if systemOverridePatterns.MatchString(input) {
		return fmt.Errorf("%s contains system override attempts", fieldName)
	}

	return nil
}

func SanitizeInput(input string) string {
	// Remove HTML tags and escape HTML entities
	input = html.EscapeString(input)

	// remove or replace protentiallyy dangerous characters
	dangerousChars := regexp.MustCompile(`[<>"'&\x00-\x1f\x7f-\x9f]`)
	input = dangerousChars.ReplaceAllString(input, "")

	// Remove SQL injection patterns
	sqlPatterns := regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|vbscript|onload|onerror|onclick)`)
	input = sqlPatterns.ReplaceAllString(input, "")

	// Remove excessive whitespace and normalize
	input = regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(input), " ")

	return input
}
