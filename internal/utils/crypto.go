package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// EncodeFirebaseConfig encodes the Firebase config JSON with AES encryption and base64
func EncodeFirebaseConfig(configJSON []byte, secretKey string) (string, error) {
	// Create AES cipher
	key := []byte(secretKey)
	if len(key) != 32 {
		return "", errors.New("secret key must be 32 bytes long")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, configJSON, nil)

	// Encode to base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecodeFirebaseConfig decodes the base64 encoded and encrypted Firebase config
func DecodeFirebaseConfig(encodedConfig, secretKey string) ([]byte, error) {
	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encodedConfig)
	if err != nil {
		return nil, err
	}

	// Create AES cipher
	key := []byte(secretKey)
	if len(key) != 32 {
		return nil, errors.New("secret key must be 32 bytes long")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
