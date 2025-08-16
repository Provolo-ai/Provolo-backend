package main

import (
	"fmt"
	"log"
	"os"
	"provolo-api/internal/utils"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run cmd/encode/main.go <firebase-config-file> <32-char-secret-key>")
		fmt.Println("Example: go run cmd/encode/main.go firebaseConfig.json myverysecretkey1234567890123456")
		os.Exit(1)
	}

	configFile := os.Args[1]
	secretKey := os.Args[2]

	if len(secretKey) != 32 {
		log.Fatal("Secret key must be exactly 32 characters long")
	}

	// Read the Firebase config file
	configData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Encode the config
	encodedConfig, err := utils.EncodeFirebaseConfig(configData, secretKey)
	if err != nil {
		log.Fatalf("Error encoding config: %v", err)
	}

	// Write to encoded file
	encodedFile := "firebase_config_encoded.txt"
	err = os.WriteFile(encodedFile, []byte(encodedConfig), 0644)
	if err != nil {
		log.Fatalf("Error writing encoded file: %v", err)
	}

	fmt.Printf("Firebase config encoded successfully!\n")
	fmt.Printf("Encoded config saved to: %s\n", encodedFile)
	fmt.Printf("You can now commit this encoded file to your repository.\n")
	fmt.Printf("\nEncoded config:\n%s\n", encodedConfig)
}
