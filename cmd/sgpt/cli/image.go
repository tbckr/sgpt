package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/sashabaranov/go-openai"
	"github.com/tbckr/sgpt"
	"github.com/tbckr/sgpt/internal/file"
	"github.com/tbckr/sgpt/internal/shell"
)

const defaultImageFiletype = ".png"

var imageCmd = &ffcli.Command{
	Name:       "image",
	ShortUsage: "sgpt image [command flags] <prompt>",
	ShortHelp:  "Create an AI generated image with dalle.",
	LongHelp: strings.TrimSpace(`
Create an AI generated image with the DALLE API. 
`),
	Exec: runImage,
	FlagSet: (func() *flag.FlagSet {
		fs := newFlagSet("image")
		fs.IntVar(&imageArgs.count, "count", 1, "Number of images to generate")
		fs.StringVar(&imageArgs.size, "size", openai.CreateImageSize256x256, "Image size")
		fs.BoolVar(&imageArgs.download, "download", false, "Download generated images")
		fs.StringVar(&imageArgs.filename, "output", "", "Filename including path to file - might be used for base name, if multiple images are created")
		return fs
	})(),
}

var imageArgs struct {
	count    int
	size     string
	download bool
	filename string
}

func runImage(ctx context.Context, args []string) error {
	prompt, err := shell.GetPrompt(args)
	if err != nil {
		return err
	}

	options := sgpt.ImageOptions{
		Count: imageArgs.count,
		Size:  imageArgs.size,
	}

	var responseFormat string
	if imageArgs.download {
		responseFormat = openai.CreateImageResponseFormatB64JSON
	} else {
		responseFormat = openai.CreateImageResponseFormatURL
	}

	var client *openai.Client
	client, err = sgpt.CreateClient()
	if err != nil {
		return err
	}

	var imageData []string
	imageData, err = sgpt.GetImage(ctx, client, options, prompt, responseFormat)
	if err != nil {
		return err
	}

	return handleImageURLs(imageData)
}

func handleImageURLs(images []string) error {
	for _, data := range images {
		if imageArgs.download {
			// Save base64 encoded data to file
			if err := file.SaveEncodedImage(imageArgs.filename, data, stdout); err != nil {
				return err
			}
		} else {
			// Print url to stdout
			if _, err := fmt.Fprintln(stdout, data); err != nil {
				return err
			}
		}
	}
	return nil
}
