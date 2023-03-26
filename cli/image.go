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

package cli

import (
	"fmt"
	"strings"

	"github.com/tbckr/sgpt/api"
	"github.com/tbckr/sgpt/fs"

	"github.com/spf13/cobra"
	"github.com/tbckr/sgpt/image"
	"github.com/tbckr/sgpt/shell"
)

var imageArgs struct {
	count      int
	size       string
	download   bool
	filePrefix string
}

func imageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image <prompt>",
		Short: "Create an AI generated image with DALLE",
		Long: strings.TrimSpace(`
Create an AI generated image with the DALLE API.

Downloaded images have the filename pattern: <prefix>-<random suffix>.png
`),
		RunE: runImage,
	}
	fs := cmd.Flags()
	fs.IntVarP(&imageArgs.count, "count", "c", 1, "number of images to generate")
	fs.StringVar(&imageArgs.size, "size", api.DefaultImageSize, "image size")
	fs.BoolVarP(&imageArgs.download, "download", "d", false, "download generated images")
	fs.StringVar(&imageArgs.filePrefix, "prefix", "img", "file prefix for downloaded image")
	return cmd
}

func runImage(cmd *cobra.Command, args []string) error {
	prompt, err := shell.GetInput(args)
	if err != nil {
		return err
	}

	var responseFormat string
	if imageArgs.download {
		responseFormat = api.ImageData
	} else {
		responseFormat = api.ImageURL
	}

	options := api.ImageOptions{
		Count:          imageArgs.count,
		Size:           imageArgs.size,
		ResponseFormat: responseFormat,
	}

	var cli *api.Client
	cli, err = api.CreateClient()
	if err != nil {
		return err
	}

	var imageData []string
	imageData, err = cli.GetImage(cmd.Context(), options, prompt)
	if err != nil {
		return err
	}

	return handleImageURLs(imageData)
}

func handleImageURLs(images []string) error {
	var filename string
	for _, data := range images {
		if imageArgs.download {
			suffix, err := fs.CreateRandomFileSuffix(10)
			if err != nil {
				return err
			}
			filename = fmt.Sprintf("%s-%s%s", imageArgs.filePrefix, suffix, image.DefaultExtension)
			if err = image.SaveB64EncodedImage(filename, data); err != nil {
				return err
			}
		} else {
			filename = data
		}
		if _, err := fmt.Fprintln(stdout, filename); err != nil {
			return err
		}
	}
	return nil
}
