package utils

import (
	"html"
	"regexp"
	"strings"
)

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
