package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tbckr/sgpt/client"
	"github.com/tbckr/sgpt/file"
	"github.com/tbckr/sgpt/shell"
)

var imageArgs struct {
	count    int
	size     string
	download bool
	filename string
}

func imageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image <prompt>",
		Short: "Create an AI generated image with DALLE",
		Long: strings.TrimSpace(`
Create an AI generated image with the DALLE API. 
`),
		RunE: runImage,
		Args: cobra.ExactArgs(1),
	}
	fs := cmd.Flags()
	fs.IntVarP(&imageArgs.count, "count", "c", 1, "number of images to generate")
	fs.StringVar(&imageArgs.size, "size", client.DefaultImageSize, "image size")
	fs.BoolVarP(&imageArgs.download, "download", "d", false, "download generated images")
	fs.StringVarP(&imageArgs.filename, "output", "o", "", "filename including path to file - might be used for base name, if multiple images are created")
	return cmd
}

func runImage(cmd *cobra.Command, args []string) error {
	prompt, err := shell.GetInput(args)
	if err != nil {
		return err
	}

	var responseFormat string
	if imageArgs.download {
		responseFormat = client.ImageData
	} else {
		responseFormat = client.ImageURL
	}

	options := client.ImageOptions{
		Count:          imageArgs.count,
		Size:           imageArgs.size,
		ResponseFormat: responseFormat,
	}

	var cli *client.Client
	cli, err = client.CreateClient()
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
