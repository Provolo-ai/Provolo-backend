package tests

import (
	"os"
	"testing"
)

// TestMain sets up the test environment before running tests
func TestMain(m *testing.M) {
	// Set test environment variables
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("PORT", "8080")
	os.Setenv("FIREBASE_ENCODED_CONFIG", "test-config")
	os.Setenv("FIREBASE_SECRET_KEY", "test-secret")
	os.Setenv("GEMINI_API_KEY", "test-gemini-key")
	os.Setenv("MAX_PROMPT_LIMIT", "2")
	
	// Run tests
	exitCode := m.Run()
	
	// Clean up
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("PORT")
	os.Unsetenv("FIREBASE_ENCODED_CONFIG")
	os.Unsetenv("FIREBASE_SECRET_KEY")
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("MAX_PROMPT_LIMIT")
	
	os.Exit(exitCode)
}
