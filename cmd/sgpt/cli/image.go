package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tbckr/sgpt"
)

var imageCmd = &ffcli.Command{
	Name:       "image",
	ShortUsage: "sgpt image [command flags] <prompt>",
	ShortHelp:  "Create an AI generated image with dalle.",
	LongHelp: strings.TrimSpace(`
Create an AI generated image with the dalle API. 
`),
	Exec: runImage,
	FlagSet: (func() *flag.FlagSet {
		fs := newFlagSet("image")
		fs.IntVar(&imageArgs.count, "n", 1, "Number of images to generate")
		//fs.BoolVar(&imageArgs.open, "open", false, "Open image in default browser")
		//fs.BoolVar(&imageArgs.download, "download", false, "Download generated images")
		//fs.StringVar(&imageArgs.outputDirectory, "output", ".", "Path to folder to save the image in")
		fs.BoolVar(&imageArgs.json, "json", false, "Output image urls in json format")
		return fs
	})(),
}

var imageArgs struct {
	count           int
	open            bool
	download        bool
	outputDirectory string
	json            bool
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
	return handleImageURLs(imageUrls)
}

func handleImageURLs(imageUrls []string) error {
	if imageArgs.json {
		if err := printURLsAsJSON(imageUrls); err != nil {
			return err
		}
	} else {
		if err := printURLs(imageUrls); err != nil {
			return err
		}
	}

	//if imageArgs.download {
	//	if err := downloadImages(imageUrls); err != nil {
	//		return err
	//	}
	//}
	//if imageArgs.open {
	//	if err := openImages(imageUrls); err != nil {
	//		return err
	//	}
	//}
	return nil
}

func printURLs(urls []string) error {
	for _, element := range urls {
		if _, err := fmt.Fprintln(stdout, element); err != nil {
			return err
		}
	}
	return nil
}

func printURLsAsJSON(urls []string) error {
	urlMap := map[string][]string{
		"imageURLs": urls,
	}
	jsonData, err := json.Marshal(urlMap)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, jsonData)
	return err
}

//func openImages(_ []string) error {
//	return nil
//}
//
//func downloadImages(_ []string) error {
//	return nil
//}
