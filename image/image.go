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

package image

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"os"

	"github.com/tbckr/sgpt/filesystem"
)

const DefaultExtension = ".png"

func SaveB64EncodedImage(filename, imageData string) error {
	imgBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(imgBytes)
	var img image.Image
	img, err = png.Decode(reader)
	if err != nil {
		return err
	}

	var exists bool
	exists, err = filesystem.FileExists(filename)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("file already exists")
	}

	var file *os.File
	file, err = os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = png.Encode(file, img); err != nil {
		return err
	}
	return nil
}
