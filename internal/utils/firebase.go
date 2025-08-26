package utils

import (
	"context"
	"fmt"
	"os"
	"provolo-api/internal/env"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

// GetFirebaseApp creates and returns a Firebase app instance with proper configuration
func GetFirebaseApp(ctx context.Context) (*firebase.App, error) {
	// Try to get encoded config from environment first
	encodedConfig := env.GetEnvString("FIREBASE_ENCODED_CONFIG", "")
	secretKey := env.GetEnvString("FIREBASE_SECRET_KEY", "")

	var opt option.ClientOption

	if encodedConfig != "" && secretKey != "" {
		// Decode the Firebase config
		configData, err := DecodeFirebaseConfig(encodedConfig, secretKey)
		if err != nil {
			return nil, fmt.Errorf("error decoding Firebase config: %v", err)
		}

		// Create credentials from JSON data
		opt = option.WithCredentialsJSON(configData)
	} else {
		// Fallback to encoded file if environment variables are not set
		if _, err := os.Stat("firebase_config_encoded.txt"); err == nil {
			// Read encoded config from file
			encodedData, err := os.ReadFile("firebase_config_encoded.txt")
			if err != nil {
				return nil, fmt.Errorf("error reading encoded config file: %v", err)
			}

			if secretKey == "" {
				return nil, fmt.Errorf("FIREBASE_SECRET_KEY environment variable is required")
			}

			// Decode the Firebase config
			configData, err := DecodeFirebaseConfig(string(encodedData), secretKey)
			if err != nil {
				return nil, fmt.Errorf("error decoding Firebase config: %v", err)
			}

			// Create credentials from JSON data
			opt = option.WithCredentialsJSON(configData)
		} else {
			// Final fallback to original file (for development)
			opt = option.WithCredentialsFile("firebaseConfig.json")
		}
	}

	// Initialize Firebase Admin SDK
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase app: %v", err)
	}

	return app, nil
}
