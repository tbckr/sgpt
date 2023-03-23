package cli

import (
	"fmt"
	"strings"

	"github.com/tbckr/sgpt/api"
	"github.com/tbckr/sgpt/filesystem"

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
		Args: cobra.ExactArgs(1),
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
			suffix, err := filesystem.CreateRandomFileSuffix(10)
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
