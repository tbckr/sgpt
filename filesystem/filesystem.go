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

package filesystem

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"path"
)

const defaultDirPermissions = 0750

func GetAppCacheDir(applicationName string) (string, error) {
	// Get user specific config dir
	baseCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	// Application specific cache dir
	configPath := path.Join(baseCacheDir, applicationName)
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

func FileExists(filename string) (bool, error) {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func CreateRandomFileSuffix(size int) (string, error) {
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
