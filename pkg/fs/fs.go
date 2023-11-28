// Copyright (c) 2023 Tim <tbckr>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
// SPDX-License-Identifier: MIT

package fs

import (
	"bufio"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

const (
	defaultDirPermissions = 0755
	appName               = "sgpt"
)

func createPath(dirs ...string) (string, error) {
	appPath := filepath.Join(dirs...)
	// if app dir does not exist, create it
	if _, err := os.Stat(appPath); errors.Is(err, os.ErrNotExist) {
		slog.Debug("Creating directory: " + appPath)
		if err = os.MkdirAll(appPath, defaultDirPermissions); err != nil {
			return "", err
		}
	}
	return appPath, nil
}

func GetAppConfigPath() (string, error) {
	configPath, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return createPath(configPath, appName)
}

func GetAppCacheDir() (string, error) {
	// Get user specific config dir
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return createPath(cacheDir, appName)
}

func GetPersonasPath() (string, error) {
	configPath, err := GetAppConfigPath()
	if err != nil {
		return "", err
	}
	return createPath(configPath, "personas")
}

func ReadString(in io.Reader) (string, error) {
	var buf []byte
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		buf = append(buf, scanner.Bytes()...)
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	input := string(buf)
	return input, nil
}

// GetImageFileType returns the file type of images
func GetImageFileType(inputFile string) (string, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", err
	}

	// Reset the read pointer.
	_, _ = file.Seek(0, 0)

	// Always returns a valid content-type and "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

// LoadBase64ImageFromFile loads a base64 encoded image from a file
func LoadBase64ImageFromFile(inputFile string) (string, error) {
	// Load image from file
	imageBytes, err := os.ReadFile(inputFile)
	if err != nil {
		return "", err
	}
	// Convert image to base64
	b64Image := base64.StdEncoding.EncodeToString(imageBytes)
	return b64Image, nil
}
