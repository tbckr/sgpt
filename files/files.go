package files

import (
	"crypto/rand"
	"encoding/base64"
	"os"
)

func Exists(filename string) (bool, error) {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func CreateRandomSuffix(size int) (string, error) {
	// Generate a random byte slice
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Encode the byte slice as a string using base64 encoding
	randomSuffix := base64.URLEncoding.EncodeToString(b)
	return randomSuffix, nil
}
