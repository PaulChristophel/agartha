package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

func generateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be a positive integer")
	}

	// Calculate how many bytes are needed to produce enough base64 characters
	encodedLength := 4 * ((length + 2) / 3)
	b := make([]byte, encodedLength)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	encoded := base64.URLEncoding.EncodeToString(b)

	// Ensure the encoded string is at least 'length' characters long
	if len(encoded) < length {
		return "", errors.New("encoded string is shorter than expected")
	}

	return encoded[:length], nil
}
