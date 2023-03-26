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
	"crypto/rand"
	"encoding/base64"
	"os"
	"path"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
)

const defaultDirPermissions = 0750

func GetAppCacheDir(applicationName string) (string, error) {
	// Get user specific config dir
	baseCacheDir, err := os.UserCacheDir()
	if err != nil {
		jww.ERROR.Println("Could not get user cache dir")
		return "", err
	}
	// Application specific cache dir
	configPath := path.Join(baseCacheDir, applicationName)
	jww.DEBUG.Println("Application cache dir:", configPath)
	_, err = os.Stat(configPath)
	// Check, if application cache dir exists
	if os.IsNotExist(err) {
		jww.ERROR.Println("Application cache dir does not exist - creating it")
		// Create application cache dir
		if err = os.MkdirAll(configPath, defaultDirPermissions); err != nil {
			jww.ERROR.Println("Could not create application cache dir")
			return "", err
		}
	}
	return configPath, nil
}

func FileExists(filename string) (bool, error) {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			jww.DEBUG.Println("File does not exist: ", filename)
			return false, nil
		}
		jww.ERROR.Println("Could not check if file exists: ", filename)
		return false, err
	}
	jww.DEBUG.Println("File exists: ", filename)
	return true, nil
}

func CreateRandomFileSuffix(size int) (string, error) {
	// Generate a random byte slice
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		jww.ERROR.Println("Could not generate random byte slice")
		return "", err
	}
	// Encode the byte slice as a string using base64 encoding
	randomSuffix := base64.URLEncoding.EncodeToString(b)
	randomSuffix = strings.TrimSuffix(randomSuffix, "==")
	jww.DEBUG.Println("Generated random file suffix: ", randomSuffix)
	return randomSuffix, nil
}
