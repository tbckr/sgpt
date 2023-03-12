package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tbckr/sgpt"
)

var imageCmd = &ffcli.Command{
	Name:       "image",
	ShortUsage: "",
	ShortHelp:  "",
	LongHelp:   strings.TrimSpace(``),
	Exec:       runImage,
	FlagSet: (func() *flag.FlagSet {
		fs := newFlagSet("image")
		fs.IntVar(&imageArgs.count, "n", 1, "Number of images to generate")
		fs.BoolVar(&imageArgs.open, "open", false, "Open image in default browser")
		fs.BoolVar(&imageArgs.download, "download", false, "Download generated images")
		fs.StringVar(&imageArgs.directory, "output", ".", "Path to folder to save the image in")
		fs.BoolVar(&imageArgs.json, "json", false, "Output image urls in json format")
		return fs
	})(),
}

var imageArgs struct {
	count     int
	open      bool
	download  bool
	directory string
	json      bool
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

	client, err := sgpt.CreateClient()
	if err != nil {
		return err
	}

	var imageUrls []string
	imageUrls, err = sgpt.GetImage(ctx, client, options, prompt)
	if err != nil {
		return err
	}
	for _, element := range imageUrls {
		if _, err = fmt.Fprint(stdout, element); err != nil {
			return err
		}
	}
	return nil
}
