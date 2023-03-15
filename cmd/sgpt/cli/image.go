package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tbckr/sgpt"
	"github.com/tbckr/sgpt/image"
)

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
		fs.IntVar(&imageArgs.count, "n", 1, "Number of images to generate")
		fs.BoolVar(&imageArgs.download, "download", false, "Download generated images")
		fs.StringVar(&imageArgs.outputFilename, "output", "", "Filename including path to file - might be used for base name, if multiple images are created")
		return fs
	})(),
}

var imageArgs struct {
	count          int
	download       bool
	outputFilename string
}

func runImage(ctx context.Context, args []string) error {
	// Check, if prompt was provided via command line
	if len(args) != 1 {
		return ErrMissingPrompt
	}
	prompt := args[0]

	options := sgpt.ImageOptions{
		Count: imageArgs.count,
	}
	if err := sgpt.ValidateImageOptions(options); err != nil {
		return err
	}

	client, err := sgpt.CreateClient()
	if err != nil {
		return err
	}

	var imageUrls []string
	imageUrls, err = sgpt.GetImage(ctx, client, options, prompt)
	if err != nil {
		return err
	}
	return handleImageURLs(ctx, imageUrls)
}

func handleImageURLs(ctx context.Context, imageUrls []string) error {
	client := image.NewClient()
	client.SetDefaultFilename(imageArgs.outputFilename)

	for _, url := range imageUrls {

		if imageArgs.download {
			if err := client.DownloadImage(ctx, url); err != nil {
				return err
			}

		} else {
			if _, err := fmt.Fprintln(stdout, url); err != nil {
				return err
			}
		}
	}
	return nil
}
