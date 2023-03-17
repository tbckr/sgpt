package file

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const defaultImageFiletype = ".png"

func getFilename(defaultFilename string) (string, error) {
	var filename string
	var err error

	// Generate a filename, if no default filename was provided
	if defaultFilename == "" {
		filename, err = generateFilename()
		if err != nil {
			return "", err
		}
	} else {
		filename = defaultFilename
	}

	// Check, if file already exists
	var exists bool
	baseFilename := strings.Clone(filename)
	for {
		exists, err = fileExists(filename)
		if err != nil {
			return "", err
		}
		// If file exists, append random string
		if exists {
			filename, err = makeUnique(baseFilename)
			if err != nil {
				return "", err
			}
		} else {
			break
		}
	}
	return filename, nil
}

func generateFilename() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return u.String() + defaultImageFiletype, nil
}

func fileExists(filename string) (bool, error) {
	// If file exists, append random string
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func makeUnique(filename string) (string, error) {
	// Get filename and extension
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]

	if ext == "" {
		ext = defaultImageFiletype
	}

	// Generate a random byte slice
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode the byte slice as a string using base64 encoding
	randomSuffix := base64.URLEncoding.EncodeToString(b)

	if name == "" {
		return fmt.Sprintf("%s%s", randomSuffix, ext), nil
	}
	return fmt.Sprintf("%s_%s%s", name, randomSuffix, ext), nil
}
