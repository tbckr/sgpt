package files

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"path"
)

const defaultDirPermissions = 0750

func GetAppCacheDir(applicationName string) (string, error) {
	// Get user specific config dir
	baseConfigDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	// Application specific cache dir
	configPath := path.Join(baseConfigDir, applicationName)
	_, err = os.Stat(configPath)
	// Check, if application cache dir exists
	if os.IsNotExist(err) {
		// Create application cache dir
		if err = os.MkdirAll(configPath, defaultDirPermissions); err != nil {
			return "", err
		}
	}
	return configPath, nil
}

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
